package api

import (
	"context"
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
)

type TaskHandler struct {
	TaskService *service.TaskService
}

// GetTaskV1 implements openapi.TaskAPIServicer
func (ts *TaskHandler) GetTaskV1(_ context.Context, taskName string) (openapi.ImplResponse, error) {
	t, content := ts.TaskService.GetTask(taskName)
	if t == nil {
		return openapi.Response(http.StatusNotFound, openapi.Error{Error: "Not Found", Message: "Task unknown"}), nil
	}

	body := openapi.GetTaskV1Response{
		Name:    t.Task.Name,
		Hash:    t.Sha256,
		Content: content,
	}
	return openapi.Response(http.StatusOK, body), nil
}

// ListTasksV1 implements openapi.TaskAPIServicer
func (th *TaskHandler) ListTasksV1(_ context.Context) (openapi.ImplResponse, error) {
	body := openapi.ListTasksV1Response{
		Tasks: []string{},
	}
	for _, entry := range th.TaskService.ListTasks() {
		body.Tasks = append(body.Tasks, entry.Task.Name)
	}

	return openapi.Response(http.StatusOK, body), nil
}
