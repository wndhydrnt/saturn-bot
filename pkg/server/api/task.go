package api

import (
	"context"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

// GetTaskV1 implements [openapi.ServerInterface].
func (a *APIServer) GetTaskV1(_ context.Context, request openapi.GetTaskV1RequestObject) (openapi.GetTaskV1ResponseObject, error) {
	t, content := a.TaskService.GetTask(request.Task)
	if t == nil {
		return openapi.GetTaskV1404JSONResponse{
			Error:   "Not Found",
			Message: "Task unknown",
		}, nil
	}

	return openapi.GetTaskV1200JSONResponse{
		Name:    t.Name,
		Hash:    t.Checksum(),
		Content: content,
	}, nil
}

// ListTasksV1 implements [openapi.ServerInterface].
func (th *APIServer) ListTasksV1(_ context.Context, request openapi.ListTasksV1RequestObject) (openapi.ListTasksV1ResponseObject, error) {
	resp := openapi.ListTasksV1200JSONResponse{
		Tasks: []string{},
	}
	for _, entry := range th.TaskService.ListTasks() {
		resp.Tasks = append(resp.Tasks, entry.Task.Name)
	}

	return resp, nil
}
