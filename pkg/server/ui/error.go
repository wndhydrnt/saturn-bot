package ui

import (
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type DataError struct {
	ID      int
	Message string
}

func renderApiError(err openapi.Error, w http.ResponseWriter) {
	renderTemplate("error.html", DataError{ID: err.Error, Message: err.Message}, w)
}

func renderError(err error, w http.ResponseWriter) {
	renderTemplate("error.html", DataError{Message: err.Error()}, w)
}
