package ui

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	Content string
	Name    string
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
		data := dataGetTaskFile{Content: v.Content, Name: v.Name}
		renderTemplate("task-get-file.html", data, w)

	case openapi.GetTaskV1404JSONResponse:
		renderApiError(openapi.Error(v), w)

	case openapi.GetTaskV1500JSONResponse:
		renderApiError(openapi.Error(v), w)
	}
}
