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

func renderApiError(err openapi.Error, w http.ResponseWriter) {
	renderTemplate("error.html", DataError{ID: err.Error, Message: err.Message}, w)
}

func renderError(err error, w http.ResponseWriter) {
	log.Log().Errorw("Rendering of UI failed", zap.Error(err))
	renderTemplate("error.html", DataError{Message: err.Error()}, w)
}
