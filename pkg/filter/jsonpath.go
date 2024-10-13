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

func (f JsonPathFactory) Create(params params.Params) (Filter, error) {
	exprRaw, err := params.String("expression", "")
	if err != nil {
		return nil, err
	}

	if exprRaw == "" {
		return nil, fmt.Errorf("required parameter `expression` not set")
	}

	expr, err := jp.ParseString(exprRaw)
	if err != nil {
		return nil, fmt.Errorf("parse JSONPath expression: %w", err)
	}

	path, err := params.String("path", "")
	if err != nil {
		return nil, err
	}

	if path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	return &JsonPath{
		Expr: expr,
		Path: path,
	}, nil
}

func (f JsonPathFactory) Name() string {
	return "jsonpath"
}

type JsonPath struct {
	Expr jp.Expr
	Path string
}

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

	result := f.Expr.Get(data)
	return len(result) > 0, nil
}

func (jp *JsonPath) String() string {
	return fmt.Sprintf("jsonpath(expression=%s,path=%s)", jp.Expr, jp.Path)
}
