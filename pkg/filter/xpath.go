package filter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
)

type XpathFactory struct{}

func (f XpathFactory) Create(params params.Params) (Filter, error) {
	expr, err := params.String("expression", "")
	if err != nil {
		return nil, err
	}

	if expr == "" {
		return nil, fmt.Errorf("required parameter `expression` not set")
	}

	// Compile expression here to return early if it is invalid.
	_, err = xpath.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid XPath expression: %w", err)
	}

	path, err := params.String("path", "")
	if err != nil {
		return nil, err
	}

	if path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	return &Xpath{Expr: expr, Path: path}, nil
}

func (f XpathFactory) Name() string {
	return "xpath"
}

type Xpath struct {
	Expr string
	Path string
}

func (x *Xpath) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(sbcontext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context does not contain a repository")
	}

	content, err := repo.GetFile(x.Path)
	if err != nil {
		if errors.Is(err, host.ErrFileNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("download file from repository: %w", err)
	}

	doc, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return false, fmt.Errorf("parse XML document: %w", err)
	}

	// Note: QueryAll uses an internal cache to prevent parsing the same expression multiple times.
	result, err := xmlquery.QueryAll(doc, x.Expr)
	if err != nil {
		return false, fmt.Errorf("query XML document: %w", err)
	}

	return len(result) > 0, nil
}

func (x *Xpath) String() string {
	return fmt.Sprintf("xpath(expression=%s,path=%s)", x.Expr, x.Path)
}
