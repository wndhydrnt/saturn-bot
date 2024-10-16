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
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
)

// XpathFactory creates Xpath filters.
type XpathFactory struct{}

// Create implements Factory.
func (f XpathFactory) Create(params params.Params) (Filter, error) {
	expressionsRaw, err := params.StringSlice("expressions", []string{})
	if err != nil {
		return nil, err
	}

	if len(expressionsRaw) == 0 {
		return nil, fmt.Errorf("required parameter `expressions` not set")
	}

	xp := &Xpath{}
	for idx, exprRaw := range expressionsRaw {
		expr, err := xpath.Compile(exprRaw)
		if err != nil {
			return nil, fmt.Errorf("compile `expressions[%d]` XPath expression: %w", idx, err)
		}

		xp.Exprs = append(xp.Exprs, expr)
	}

	xp.Path, err = params.String("path", "")
	if err != nil {
		return nil, err
	}

	if xp.Path == "" {
		return nil, fmt.Errorf("required parameter `path` not set")
	}

	return xp, nil
}

// Name implements Factory.
func (f XpathFactory) Name() string {
	return "xpath"
}

// Xpath filters repositories by applying an XPath expression to a file in the repository.
type Xpath struct {
	Exprs []*xpath.Expr
	Path  string
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
		log.Log().Warnf("Failed to parse XML file %s: %s", x.Path, err)
		return false, nil
	}

	for _, expr := range x.Exprs {
		nodes := xmlquery.QuerySelectorAll(doc, expr)
		if len(nodes) == 0 {
			return false, nil
		}
	}

	return true, nil
}

// String implements Filter.
func (x *Xpath) String() string {
	var expressions []string
	for _, expr := range x.Exprs {
		expressions = append(expressions, expr.String())
	}
	return fmt.Sprintf("xpath(expressions=%s,path=%s)", expressions, x.Path)
}
