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

var (
	runStatusOptions        = []string{string(openapi.Failed), string(openapi.Finished), string(openapi.Pending), string(openapi.Running)}
	taskResultStatusOptions = []openapi.TaskResultStateV1{openapi.TaskResultStateV1Closed, openapi.TaskResultStateV1Error, openapi.TaskResultStateV1Merged, openapi.TaskResultStateV1Open}
)

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

	renderTemplate(tplData, w, "run-list.html")
}

type dataListTaskResultsOfRun struct {
	DisplayRunLink bool
	Filters        dataTaskResultsFilters
	Pagination     pagination
	Run            openapi.RunV1
	TaskResults    []openapi.TaskResultV1
}

// GetRun renders the details and results of a run.
func (u *Ui) GetRun(w http.ResponseWriter, r *http.Request) {
	runId, err := strconv.Atoi(chi.URLParam(r, "runId"))
	if err != nil {
		renderError(fmt.Errorf("convert parameter runId to int: %w", err), w)
		return
	}

	getRunReq := openapi.GetRunV1RequestObject{
		RunId: runId,
	}
	getRunResp, err := u.API.GetRunV1(r.Context(), getRunReq)
	if err != nil {
		renderError(err, w)
		return
	}

	statusParam := r.URL.Query().Get("status")
	data := dataListTaskResultsOfRun{
		DisplayRunLink: false,
		Filters: dataTaskResultsFilters{
			TaskResultStatusCurrent: statusParam,
			TaskResultStatusList:    taskResultStatusOptions,
		},
	}
	switch getRunObj := getRunResp.(type) {
	case openapi.GetRunV1200JSONResponse:
		data.Run = getRunObj.Run
	case openapi.GetRunV1404JSONResponse:
		renderApiError(openapi.Error(getRunObj), w, http.StatusNotFound, "")
		return
	}

	listTaskResultsReq := openapi.ListTaskResultsV1RequestObject{
		Params: openapi.ListTaskResultsV1Params{
			RunId: ptr.To(runId),
			ListOptions: &openapi.ListOptions{
				Limit: parseIntParam(r, "limit", 10),
				Page:  parseIntParam(r, "page", 1),
			},
		},
	}

	if statusParam != "" {
		listTaskResultsReq.Params.Status = ptr.To([]openapi.TaskResultStateV1{openapi.TaskResultStateV1(statusParam)})
	}

	listTaskResultsResp, err := u.API.ListTaskResultsV1(r.Context(), listTaskResultsReq)
	if err != nil {
		renderError(err, w)
		return
	}

	listTaskResultsObj := listTaskResultsResp.(openapi.ListTaskResultsV1200JSONResponse)
	data.Pagination = pagination{
		Page: listTaskResultsObj.Page,
		URL:  r.URL,
	}
	data.TaskResults = listTaskResultsObj.TaskResults
	renderTemplate(data, w, "task-results-table.html", "run-get.html")
}
