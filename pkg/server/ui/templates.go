package ui

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Masterminds/sprig/v3"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"go.uber.org/zap"
)

//go:embed templates/*.html
var templateFS embed.FS

var templateFuncs = template.FuncMap{
	"pathEscape":                 url.PathEscape,
	"renderUrl":                  renderUrl,
	"runStatusToCssClass":        mapRunStatusToCssClass,
	"taskResultStatusToCssClass": mapTaskResultStatusToCssClass,
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
