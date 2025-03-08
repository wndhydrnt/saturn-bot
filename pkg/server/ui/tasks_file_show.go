package ui

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataTasksFileShow struct {
	Content  string
	TaskName string
}

// TasksFileShow renders the content of the file of a task.
func (u *Ui) TasksFileShow(w http.ResponseWriter, r *http.Request) {
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
		data := dataTasksFileShow{Content: v.Content, TaskName: v.Name}
		renderTemplate(data, w, "tasks_file_show.html")

	case openapi.GetTaskV1404JSONResponse:
		renderApiError(openapi.Error(v), w, http.StatusNotFound, "")

	case openapi.GetTaskV1500JSONResponse:
		renderApiError(openapi.Error(v), w, http.StatusInternalServerError, "")
	}
}
