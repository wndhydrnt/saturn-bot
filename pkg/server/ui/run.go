package ui

import (
	"context"
	"fmt"
	"net/http"

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
	for _, t := range taskList.Results {
		tplData.Filters.TaskNames = append(tplData.Filters.TaskNames, t.Name)
	}

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
