package ui

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"go.uber.org/zap"
)

//go:embed templates/*.html
var templateFS embed.FS

var templateFuncs = template.FuncMap{
	"runStatusToCssClass": mapRunStatusToCssClass,
}

var templateRoot = template.Must(template.New("").Funcs(templateFuncs).Funcs(sprig.FuncMap()).ParseFS(templateFS, "templates/base.html"))

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

func renderTemplate(name string, data any, w http.ResponseWriter) {
	tpl, err := template.Must(templateRoot.Clone()).ParseFS(templateFS, "templates/"+name)
	if err != nil {
		log.Log().Errorw("Parse templates", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tpl.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Log().Errorw("Execute template", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
