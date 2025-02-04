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

// ListTasks renders the list of all tasks known to saturn-bot.
func (u *Ui) ListTasks(w http.ResponseWriter, r *http.Request) {
	listTasksResp, err := u.API.ListTasksV1(r.Context(), openapi.ListTasksV1RequestObject{})
	if err != nil {
		renderError(fmt.Errorf("list tasks: %w", err), w)
		return
	}

	taskList := listTasksResp.(openapi.ListTasksV1200JSONResponse)
	var data dataListTasks
	data.Tasks = taskList.Tasks
	renderTemplate(data, w, "task-list.html")
}

type dataGetTaskFile struct {
	Content  string
	TaskName string
}

// GetTaskFile renders the content of the file of a task.
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
		renderTemplate(data, w, "task-get-file.html")

	case openapi.GetTaskV1404JSONResponse:
		renderApiError(openapi.Error(v), w, http.StatusNotFound, "")

	case openapi.GetTaskV1500JSONResponse:
		renderApiError(openapi.Error(v), w, http.StatusInternalServerError, "")
	}
}

type dataTaskResultsFilters struct {
	TaskResultStatusCurrent string
	TaskResultStatusList    []openapi.TaskResultStatusV1
}

type dataGetTaskResults struct {
	DisplayRunLink bool
	Filters        dataTaskResultsFilters
	Pagination     pagination
	TaskName       string
	TaskResults    []openapi.TaskResultV1
}

// GetTaskFile renders the list of results of the latest run of a task.
func (u *Ui) GetTaskResults(w http.ResponseWriter, r *http.Request) {
	statusParam := r.URL.Query().Get("status")
	data := dataGetTaskResults{
		DisplayRunLink: true,
		Filters: dataTaskResultsFilters{
			TaskResultStatusCurrent: statusParam,
			TaskResultStatusList:    taskResultStatusOptions,
		},
	}

	name := chi.URLParam(r, "name")
	listTaskResultsReq := openapi.ListTaskRecentTaskResultsV1RequestObject{
		Task: name,
		Params: openapi.ListTaskRecentTaskResultsV1Params{
			ListOptions: &openapi.ListOptions{
				Limit: parseIntParam(r, "limit", 10),
				Page:  parseIntParam(r, "page", 1),
			},
		},
	}

	if statusParam != "" {
		listTaskResultsReq.Params.Status = ptr.To([]openapi.TaskResultStatusV1{openapi.TaskResultStatusV1(statusParam)})
	}

	listTaskResultsResp, err := u.API.ListTaskRecentTaskResultsV1(r.Context(), listTaskResultsReq)
	if err != nil {
		renderError(err, w)
		return
	}

	switch resp := listTaskResultsResp.(type) {
	case openapi.ListTaskRecentTaskResultsV1200JSONResponse:
		data.Pagination = pagination{
			Page: resp.Page,
			URL:  r.URL,
		}
		data.TaskName = name
		data.TaskResults = resp.TaskResults
		renderTemplate(data, w, "task-results-table.html", "task-get-results.html")
	case openapi.ListTaskRecentTaskResultsV1404JSONResponse:
		renderApiError(openapi.Error(resp), w, http.StatusNotFound, "")
	case openapi.ListTaskRecentTaskResultsV1500JSONResponse:
		renderApiError(openapi.Error(resp), w, http.StatusInternalServerError, "")
	}
}

type dataNewRun struct {
	Inputs   []openapi.TaskV1Input
	TaskName string
}

// NewRun returns a form to schedule a new run for a task.
// The task is identifier by the path parameter name.
func (u *Ui) NewRun(w http.ResponseWriter, r *http.Request) {
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
		data := dataNewRun{TaskName: v.Name}
		if v.Inputs != nil {
			data.Inputs = ptr.From(v.Inputs)
		}

		renderTemplate(data, w, "task-run-new.html")

	case openapi.GetTaskV1404JSONResponse:
		renderApiError(openapi.Error(v), w, http.StatusNotFound, "")

	case openapi.GetTaskV1500JSONResponse:
		renderApiError(openapi.Error(v), w, http.StatusInternalServerError, "")
	}
}

// CreateRun schedules a new run for a task.
// The task is identifier by the path parameter name.
// It redirects to the detail page of the run on success.
func (u *Ui) CreateRun(w http.ResponseWriter, r *http.Request) {
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
		runData := map[string]string{}
		if v.Inputs != nil {
			for _, input := range *v.Inputs {
				value := r.PostFormValue(input.Name)
				runData[input.Name] = value
			}
		}

		scheduleRunV1Request := openapi.ScheduleRunV1Request{
			TaskName: v.Name,
			RunData:  ptr.To(runData),
		}
		scheduleRunResp, err := u.API.ScheduleRunV1(r.Context(), openapi.ScheduleRunV1RequestObject{
			Body: ptr.To(scheduleRunV1Request),
		})
		if err != nil {
			renderError(err, w)
			return
		}

		switch resp := scheduleRunResp.(type) {
		case openapi.ScheduleRunV1200JSONResponse:
			http.Redirect(w, r, fmt.Sprintf("/ui/runs/%d", resp.RunID), http.StatusFound)
			return
		case openapi.ScheduleRunV1400JSONResponse:
			renderApiError(openapi.Error(resp), w, http.StatusBadRequest, "")
			return
		}

	case openapi.GetTaskV1404JSONResponse:
		renderApiError(openapi.Error(v), w, http.StatusNotFound, "")

	case openapi.GetTaskV1500JSONResponse:
		renderApiError(openapi.Error(v), w, http.StatusInternalServerError, "")
	}
}
