package ui

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataListRuns struct {
	Runs       []openapi.RunV1
	Pagination pagination
	Filters    dataListRunsFilters
}

type dataListRunsFilters struct {
	RunStatusList    []string
	RunStatusCurrent string
	TaskNames        []string
	TaskNameCurrent  string
}

var runStatusOptions = []string{string(openapi.Failed), string(openapi.Finished), string(openapi.Pending), string(openapi.Running)}

// ListRun renders the list of known runs.
func (u *Ui) ListRuns(w http.ResponseWriter, r *http.Request) {
	queryStatus := r.URL.Query().Get("status")
	queryTask := r.URL.Query().Get("task")
	tplData := dataListRuns{
		Filters: dataListRunsFilters{
			RunStatusList:    runStatusOptions,
			RunStatusCurrent: queryStatus,
			TaskNameCurrent:  queryTask,
		},
	}

	listTasksResp, err := u.API.ListTasksV1(r.Context(), openapi.ListTasksV1RequestObject{})
	if err != nil {
		renderError(fmt.Errorf("list tasks: %w", err), w)
		return
	}

	taskList := listTasksResp.(openapi.ListTasksV1200JSONResponse)
	tplData.Filters.TaskNames = taskList.Tasks

	limit := parseIntParam(r, "limit", 10)
	page := parseIntParam(r, "page", 1)
	req := openapi.ListRunsV1RequestObject{
		Params: openapi.ListRunsV1Params{
			ListOptions: &openapi.ListOptions{
				Limit: limit,
				Page:  page,
			},
		},
	}
	if queryStatus != "" {
		req.Params.Status = ptr.To([]openapi.RunStatusV1{openapi.RunStatusV1(queryStatus)})
	}
	if queryTask != "" {
		req.Params.Task = ptr.To(queryTask)
	}

	listRunsResp, err := u.API.ListRunsV1(context.Background(), req)
	if err != nil {
		renderError(err, w)
		return
	}

	switch payload := listRunsResp.(type) {
	case openapi.ListRunsV1200JSONResponse:
		tplData.Pagination = pagination{
			Page: payload.Page,
			URL:  r.URL,
		}
		tplData.Runs = payload.Result
	}

	renderTemplate("run-list.html", tplData, w)
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

	var tplData openapi.GetRunV1200JSONResponse
	switch payload := resp.(type) {
	case openapi.GetRunV1200JSONResponse:
		tplData = payload
	case openapi.GetRunV1404JSONResponse:
		renderApiError(openapi.Error(payload), w)
		return
	}

	renderTemplate("run-get.html", tplData, w)
}
