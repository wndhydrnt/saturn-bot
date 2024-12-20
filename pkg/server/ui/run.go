package ui

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataListRuns struct {
	Runs openapi.ListRunsV1200JSONResponse
}

// ListRun renders the list of known runs.
func (u *Ui) ListRuns(w http.ResponseWriter, r *http.Request) {
	req := openapi.ListRunsV1RequestObject{
		Params: openapi.ListRunsV1Params{
			ListOptions: &openapi.ListOptions{
				Limit: 10,
			},
		},
	}
	resp, err := u.API.ListRunsV1(context.Background(), req)
	if err != nil {
		renderError(err, w)
		return
	}

	tplData := dataListRuns{}
	switch payload := resp.(type) {
	case openapi.ListRunsV1200JSONResponse:
		tplData.Runs = payload
	}

	renderTemplate("run-list.html", tplData, w)
}

type dataGetRun struct {
	Run openapi.RunV1
}

// GetRun renders the detail page of a run.
func (u *Ui) GetRun(w http.ResponseWriter, r *http.Request) {
	runId, err := strconv.Atoi(chi.URLParam(r, "runId"))
	if err != nil {
		renderError(fmt.Errorf("convert parameter runId to int: %w", err), w)
		return
	}

	req := openapi.GetRunV1RequestObject{
		RunId: runId,
	}
	resp, err := u.API.GetRunV1(context.Background(), req)
	if err != nil {
		renderError(err, w)
		return
	}

	tplData := dataGetRun{}
	switch payload := resp.(type) {
	case openapi.GetRunV1200JSONResponse:
		tplData.Run = payload.Run
	case openapi.GetRunV1404JSONResponse:
		renderApiError(openapi.Error(payload), w)
		return
	}

	renderTemplate("run-get.html", tplData, w)
}
