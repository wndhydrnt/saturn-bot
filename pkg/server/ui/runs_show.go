package ui

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataRunsShow struct {
	DisplayRunLink bool
	Filters        dataResultsIndexFilters
	Pagination     pagination
	Run            openapi.RunV1
	TaskResults    []openapi.TaskResultV1
}

// RunsShow renders the details and results of a run.
func (u *Ui) RunsShow(w http.ResponseWriter, r *http.Request) {
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
	data := dataRunsShow{
		DisplayRunLink: false,
		Filters: dataResultsIndexFilters{
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
	renderTemplate(data, w, "results_table.html", "runs_show.html")
}
