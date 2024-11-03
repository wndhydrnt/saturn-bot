package filter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"gopkg.in/yaml.v3"
)

// JqFactory creates jq filters.
type JqFactory struct{}

// Create implements Factory.
func (f JqFactory) Create(params params.Params) (Filter, error) {
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

// Do implements Filter.
// It downloads a file from a repository and queries it via a jq expression.
// It returns true if the query returns at least one node.
func (f *Jq) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(sbcontext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context does not contain a repository")
	}

	content, err := repo.GetFile(f.Path)
	if err != nil {
		if errors.Is(err, host.ErrFileNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("download file from repository: %w", err)
	}

	var data interface{}
	dec := yaml.NewDecoder(strings.NewReader(content))
	err = dec.Decode(&data)
	if err != nil {
		return false, fmt.Errorf("decode file for jq filter: %w", err)
	}

	for _, expr := range f.Exprs {
		iter := expr.Run(data)
		for {
			valueRaw, hasNext := iter.Next()
			if !hasNext {
				// No more data.
				break
			}

			if valueRaw == nil {
				// Query hasn't matched anything.
				return false, nil
			}

			if _, isErr := valueRaw.(error); isErr {
				// gojq can return a "halt" error.
				// Consider any error as not matching.
				return false, nil
			}

			value, isBool := valueRaw.(bool)
			if isBool && !value {
				// A query can return a bool.
				// Example: '.hello == "world"'
				return false, nil
			}
		}
	}

	return true, nil
}

// String implements Filter.
func (f *Jq) String() string {
	return fmt.Sprintf("jq(expressions=[%s], path=%s)", f.ExprsString, f.Path)
}
