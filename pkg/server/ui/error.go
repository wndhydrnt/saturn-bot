package ui

import (
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"go.uber.org/zap"
)

type dataError struct {
	Link    string
	Error   openapi.Error
	Message string
}

func renderApiError(err openapi.Error, w http.ResponseWriter, status int, backLink string) {
	w.WriteHeader(status)
	data := dataError{
		Link:    backLink,
		Error:   err,
		Message: "Request failed",
	}
	renderTemplate(data, w, "error.html")
}

func renderError(err error, w http.ResponseWriter) {
	log.Log().Errorw("Rendering of UI failed", zap.Error(err))
	w.WriteHeader(http.StatusInternalServerError)
	renderTemplate(dataError{Message: err.Error()}, w, "error.html")
}
