package ui

import (
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"go.uber.org/zap"
)

type DataError struct {
	ID      int
	Message string
}

func renderApiError(err openapi.Error, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	renderTemplate(DataError{ID: err.Error, Message: err.Message}, w, "error.html")
}

func renderError(err error, w http.ResponseWriter) {
	log.Log().Errorw("Rendering of UI failed", zap.Error(err))
	w.WriteHeader(http.StatusInternalServerError)
	renderTemplate(DataError{Message: err.Error()}, w, "error.html")
}
