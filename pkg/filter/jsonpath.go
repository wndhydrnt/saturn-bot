package filter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ohler55/ojg/jp"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"gopkg.in/yaml.v3"
)

// JsonPathFactory creates JsonPath filters.
type JsonPathFactory struct{}

// Create implements Factory.
func (f JsonPathFactory) Create(params params.Params) (Filter, error) {
	expressionsRaw, err := params.StringSlice("expressions", []string{})
	if err != nil {
		return nil, err
	}

	if len(expressionsRaw) == 0 {
		return nil, fmt.Errorf("required parameter `expressions` not set")
	}

	jpf := &JsonPath{}
	for idx, exprRaw := range expressionsRaw {
		expr, err := jp.ParseString(exprRaw)
		if err != nil {
			return nil, fmt.Errorf("parse `expressions[%d]` JSONPath expression: %w", idx, err)
		}

		jpf.Exprs = append(jpf.Exprs, expr)
	}

	jpf.Path, err = params.String("path", "")
	if err != nil {
		return nil, err
	}

	if jpf.Path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	return jpf, err
}

// Name implements Factory.
func (f JsonPathFactory) Name() string {
	return "jsonpath"
}

// JsonPath is a filter that applies a JSONPath expression to a file in the repository.
type JsonPath struct {
	Exprs []jp.Expr
	Path  string
}

// Do implements Filter.
// It downloads a file from a repository and queries it via a JSONPath expression.
// It returns true if the query returns at least one node.
func (f *JsonPath) Do(ctx context.Context) (bool, error) {
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
		return false, fmt.Errorf("decode file for JSONPath filter: %w", err)
	}

	for _, expr := range f.Exprs {
		result := expr.Get(data)
		if len(result) == 0 {
			return false, nil
		}
	}

	return true, nil
}

// String implements Filter.
func (f *JsonPath) String() string {
	var expressions []string
	for _, expr := range f.Exprs {
		expressions = append(expressions, expr.String())
	}
	return fmt.Sprintf("jsonpath(expressions=%s,path=%s)", expressions, f.Path)
}
