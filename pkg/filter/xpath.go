package filter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
)

// XpathFactory creates Xpath filters.
type XpathFactory struct{}

func (f XpathFactory) CreatePreClone(_ CreateOptions, params params.Params) (Filter, error) {
	return nil, nil
}

// Create implements Factory.
func (f XpathFactory) CreatePostClone(_ CreateOptions, params params.Params) (Filter, error) {
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
// It reads a XML document in a cloned repository and queries it via an XPath expression.
// It returns true if the XPath expression returns at least one node.
func (x *Xpath) Do(ctx context.Context) (bool, error) {
	checkoutPath, ok := ctx.Value(sbcontext.CheckoutPath{}).(string)
	if !ok {
		return false, errors.New("context does not contain the checkout path")
	}

	f, err := os.Open(filepath.Join(checkoutPath, x.Path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("open file in checkout: %w", err)
	}

	defer f.Close()
	doc, err := xmlquery.Parse(f)
	if err != nil {
		sbcontext.Log(ctx).Warnf("Failed to parse XML file %s: %s", x.Path, err)
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
