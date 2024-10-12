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

// XpathFactory creates Xpath filters.
type XpathFactory struct{}

// Create implements Factory.
func (f XpathFactory) Create(params params.Params) (Filter, error) {
	exprRaw, err := params.String("expression", "")
	if err != nil {
		return nil, err
	}

	if exprRaw == "" {
		return nil, fmt.Errorf("required parameter `expression` not set")
	}

	// Compile expression here to return early if it is invalid.
	expr, err := xpath.Compile(exprRaw)
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

// Name implements Factory.
func (f XpathFactory) Name() string {
	return "xpath"
}

// Xpath filters repositories by applying an XPath expression to a file in the repository.
type Xpath struct {
	Expr *xpath.Expr
	Path string
}

// Do implements Filter.
// It downloads a XML document from a repository and queries it via an XPath expression.
// It returns true if the XPath expression returns at least one node.
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

	nodes := xmlquery.QuerySelectorAll(doc, x.Expr)
	return len(nodes) > 0, nil
}

// String implements Filter.
func (x *Xpath) String() string {
	return fmt.Sprintf("xpath(expression=%s,path=%s)", x.Expr, x.Path)
}
