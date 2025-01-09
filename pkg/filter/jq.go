package filter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/itchyny/gojq"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// JqFactory creates jq filters.
type JqFactory struct{}

func (f JqFactory) CreatePreClone(_ CreateOptions, params params.Params) (Filter, error) {
	return nil, nil
}

// Create implements Factory.
func (f JqFactory) CreatePostClone(_ CreateOptions, params params.Params) (Filter, error) {
	expressionsRaw, err := params.StringSlice("expressions", []string{})
	if err != nil {
		return nil, err
	}

	if len(expressionsRaw) == 0 {
		return nil, fmt.Errorf("required parameter `expressions` not set")
	}

	jq := &Jq{
		ExprsString: strings.Join(expressionsRaw, ", "),
	}
	for idx, exprRaw := range expressionsRaw {
		query, err := gojq.Parse(exprRaw)
		if err != nil {
			return nil, fmt.Errorf("parse `expressions[%d]` jq expression: %w", idx, err)
		}

		compiledQuery, err := gojq.Compile(query)
		if err != nil {
			return nil, fmt.Errorf("compile `expressions[%d]` jq expression: %w", idx, err)
		}

		jq.Exprs = append(jq.Exprs, compiledQuery)
	}

	jq.Path, err = params.String("path", "")
	if err != nil {
		return nil, err
	}

	if jq.Path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	return jq, err
}

// Name implements Factory.
func (f JqFactory) Name() string {
	return "jq"
}

// Jq is a filter that applies a jq expression to a file in the repository.
type Jq struct {
	Exprs       []*gojq.Code
	ExprsString string
	Path        string
}

// Do implements [Filter].
// It reads a file in a cloned repository and queries it via a jq expression.
// It returns true if the query returns at least one node.
func (jq *Jq) Do(ctx context.Context) (bool, error) {
	checkoutPath, ok := ctx.Value(sbcontext.CheckoutPath{}).(string)
	if !ok {
		return false, errors.New("context does not contain the checkout path")
	}

	f, err := os.Open(filepath.Join(checkoutPath, jq.Path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("open file in checkout: %w", err)
	}

	defer f.Close()
	var data interface{}
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&data)
	if err != nil {
		sbcontext.Log(ctx).Warnw("Failed to decode file for jq filter", zap.Error(err))
		return false, nil
	}

	for _, expr := range jq.Exprs {
		iter := expr.Run(data)
		match := checkJqMatch(iter)
		if !match {
			return false, nil
		}
	}

	return true, nil
}

// String implements Filter.
func (f *Jq) String() string {
	return fmt.Sprintf("jq(expressions=[%s], path=%s)", f.ExprsString, f.Path)
}

func checkJqMatch(iter gojq.Iter) bool {
	for {
		valueRaw, hasNext := iter.Next()
		if !hasNext {
			// No more data.
			break
		}

		if valueRaw == nil {
			// Query hasn't matched anything.
			// Continue in case the query returns more than one node.
			continue
		}

		if _, isErr := valueRaw.(error); isErr {
			// gojq can return a "halt" error.
			// Consider any error as not matching.
			return false
		}

		value, isBool := valueRaw.(bool)
		if isBool && !value {
			// A query can return a bool.
			// Example: '.hello == "world"'
			return false
		}

		return true
	}

	return false
}
