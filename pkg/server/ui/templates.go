package ui

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"go.uber.org/zap"
)

//go:embed templates/*.html
var templateFS embed.FS

var templateFuncs = template.FuncMap{
	"markdown":                   renderMarkdown,
	"pathEscape":                 url.PathEscape,
	"renderUrl":                  renderUrl,
	"runStatusToCssClass":        mapRunStatusToCssClass,
	"taskResultStatusToCssClass": mapTaskResultStatusToCssClass,
	"timeSub":                    timeSub,
}

var templateRoot = template.Must(template.New("").Funcs(templateFuncs).Funcs(sprig.FuncMap()).ParseFS(templateFS, "templates/base.html"))

type pagination struct {
	// Page information.
	Page openapi.Page
	// URL of the page.
	// Used to render links to previous/next pages.
	URL *url.URL
}

func mapRunStatusToCssClass(status openapi.RunStatusV1) string {
	switch status {
	case openapi.Failed:
		return "is-danger"
	case openapi.Finished:
		return "is-success"
	case openapi.Pending:
		return "is-info"
	case openapi.Running:
		return "is-primary"
	}

	return "is-warning"
}

func mapTaskResultStatusToCssClass(status openapi.TaskResultStateV1) string {
	switch status {
	case openapi.TaskResultStateV1Closed:
		return "is-warning"
	case openapi.TaskResultStateV1Error:
		return "is-danger"
	case openapi.TaskResultStateV1Merged, openapi.TaskResultStateV1Pushed:
		return "is-success"
	default:
		return "is-info"
	}
}

// renderUrl takes a [url.URL] and returns its string representation.
//
// Only the path and the query parameters are returned.
//
// params are an optional list of key/value pairs that are added as query parameters.
// Any existing query parameters are preserved.
func renderUrl(u *url.URL, params ...any) string {
	idx := 0
	urlValues := u.Query()
	for idx < len(params) {
		key, isString := params[idx].(string)
		if !isString {
			idx = idx + 2
			continue
		}

		switch v := params[idx+1].(type) {
		case string:
			urlValues.Set(key, v)
		case int:
			urlValues.Set(key, strconv.Itoa(v))
		}

		idx = idx + 2
	}

	if len(urlValues) == 0 {
		return u.Path
	}

	return u.Path + "?" + urlValues.Encode()
}

func renderMarkdown(input string) template.HTML {
	// Escape any HTML already present in the input to prevent attacks.
	// For example, **bold** works but <b>bold</b> does not.
	escaped := template.HTMLEscapeString(input)
	extensions := parser.CommonExtensions
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(escaped))
	if len(doc.GetChildren()) < 1 {
		// Return the escaped string if, for some reason,
		// the parser didn't find a node.
		return template.HTML(escaped) //nolint:gosec // input gets escaped above
	}

	// The parser always wraps the content in a paragraph (<p>).
	// That behavior isn't desired here.
	// Get rid of the paragraph by extracting its child nodes.
	doc.SetChildren(doc.GetChildren()[0].GetChildren())
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return template.HTML(markdown.Render(doc, renderer)) //nolint:gosec // input gets escaped above
}

func renderTemplate(data any, w http.ResponseWriter, names ...string) {
	var namesWithPrefix []string
	for _, n := range names {
		namesWithPrefix = append(namesWithPrefix, "templates/"+n)
	}

	tpl, err := template.Must(templateRoot.Clone()).ParseFS(templateFS, namesWithPrefix...)
	if err != nil {
		log.Log().Errorw("Parse templates", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tpl.ExecuteTemplate(w, names[len(names)-1], data)
	if err != nil {
		log.Log().Errorw("Execute template", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// timeSub returns the duration that has passed between start and end
// by subtracting start from end.
//
// It returns the duration in seconds, without nanoseconds.
func timeSub(start, end time.Time) string {
	return strconv.FormatFloat(end.Sub(start).Seconds(), 'f', 0, 64)
}
