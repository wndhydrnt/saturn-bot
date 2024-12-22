package ui

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataListTasks struct {
	Tasks []string
}

func (u *Ui) ListTasks(w http.ResponseWriter, r *http.Request) {
	listTasksResp, err := u.API.ListTasksV1(r.Context(), openapi.ListTasksV1RequestObject{})
	if err != nil {
		renderError(fmt.Errorf("list tasks: %w", err), w)
		return
	}

	taskList := listTasksResp.(openapi.ListTasksV1200JSONResponse)
	var data dataListTasks
	data.Tasks = taskList.Tasks
	renderTemplate("task-list.html", data, w)
}

type dataGetTaskFile struct {
	Content  string
	TaskName string
}

func (u *Ui) GetTaskFile(w http.ResponseWriter, r *http.Request) {
	reqOpts := openapi.GetTaskV1RequestObject{
		Task: chi.URLParam(r, "name"),
	}
	resp, err := u.API.GetTaskV1(r.Context(), reqOpts)
	if err != nil {
		renderError(fmt.Errorf("get task: %w", err), w)
		return
	}

	switch v := resp.(type) {
	case openapi.GetTaskV1200JSONResponse:
		data := dataGetTaskFile{Content: v.Content, TaskName: v.Name}
		renderTemplate("task-get-file.html", data, w)

	case openapi.GetTaskV1404JSONResponse:
		renderApiError(openapi.Error(v), w)

	case openapi.GetTaskV1500JSONResponse:
		renderApiError(openapi.Error(v), w)
	}
}

type dataGetTaskResults struct {
	Pagination  pagination
	TaskName    string
	TaskResults []openapi.TaskResultV1
}

func (u *Ui) GetTaskResults(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	listRunsReq := openapi.ListRunsV1RequestObject{
		Params: openapi.ListRunsV1Params{
			Task:   ptr.To(name),
			Status: ptr.To([]openapi.RunStatusV1{openapi.Failed, openapi.Finished}),
			ListOptions: &openapi.ListOptions{
				Limit: 1,
				Page:  1,
			},
		},
	}
	listRunsResp, err := u.API.ListRunsV1(r.Context(), listRunsReq)
	if err != nil {
		renderError(err, w)
		return
	}

	listRunsObj := listRunsResp.(openapi.ListRunsV1200JSONResponse)
	if len(listRunsObj.Result) == 0 {
		// No results (yet)
		data := dataGetTaskResults{
			TaskName: name,
		}
		renderTemplate("task-get-results.html", data, w)
		return
	}

	listTaskResultsReq := openapi.ListTaskResultsV1RequestObject{
		Params: openapi.ListTaskResultsV1Params{
			RunId: ptr.To(int(listRunsObj.Result[0].Id)),
			ListOptions: &openapi.ListOptions{
				Limit: parseIntParam(r, "limit", 10),
				Page:  parseIntParam(r, "page", 1),
			},
		},
	}

	listTaskResultsResp, err := u.API.ListTaskResultsV1(r.Context(), listTaskResultsReq)
	if err != nil {
		renderError(err, w)
		return
	}

	listTaskResultsObj := listTaskResultsResp.(openapi.ListTaskResultsV1200JSONResponse)
	data := dataGetTaskResults{
		Pagination: pagination{
			Page: listTaskResultsObj.Page,
			URL:  r.URL,
		},
		TaskName:    name,
		TaskResults: listTaskResultsObj.TaskResults,
	}
	renderTemplate("task-get-results.html", data, w)
}
