package ui

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type DataListRuns struct {
	Runs openapi.ListRunsV1200JSONResponse
}

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

	tplData := DataListRuns{}
	switch payload := resp.(type) {
	case openapi.ListRunsV1200JSONResponse:
		tplData.Runs = payload
	}

	renderTemplate("run-list.html", tplData, w)
}

type DataGetRun struct {
	Run openapi.RunV1
}

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

	tplData := DataGetRun{}
	switch payload := resp.(type) {
	case openapi.GetRunV1200JSONResponse:
		tplData.Run = payload.Run
	case openapi.GetRunV1404JSONResponse:
		renderApiError(openapi.Error(payload), w)
		return
	}

	renderTemplate("run-get.html", tplData, w)
}
