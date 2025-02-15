package ui

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

// RunsDestroy triggers the deletion of a run.
func (u *Ui) RunsDestroy(w http.ResponseWriter, r *http.Request) {
	runId, err := strconv.Atoi(chi.URLParam(r, "runId"))
	if err != nil {
		renderError(fmt.Errorf("convert parameter runId to int: %w", err), w)
		return
	}

	deleteRunReq := openapi.DeleteRunV1RequestObject{
		RunId: runId,
	}
	deleteRunResp, err := u.API.DeleteRunV1(r.Context(), deleteRunReq)
	if err != nil {
		renderError(err, w)
		return
	}

	switch deleteRunObj := deleteRunResp.(type) {
	case openapi.DeleteRunV1200JSONResponse:
		http.Redirect(w, r, "/ui/runs", http.StatusFound)
		return
	case openapi.DeleteRunV1400JSONResponse:
		renderApiError(openapi.Error(deleteRunObj), w, http.StatusNotFound, "")
		return
	case openapi.DeleteRunV1404JSONResponse:
		renderApiError(openapi.Error(deleteRunObj), w, http.StatusNotFound, "")
		return
	}
}
