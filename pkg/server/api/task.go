package api

import (
	"context"
	"errors"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
)

// GetTaskV1 implements [openapi.ServerInterface].
func (a *APIServer) GetTaskV1(_ context.Context, request openapi.GetTaskV1RequestObject) (openapi.GetTaskV1ResponseObject, error) {
	t, err := a.TaskService.GetTask(request.Task)
	var clientErr sberror.Client
	if errors.As(err, &clientErr) {
		return openapi.GetTaskV1404JSONResponse{
			Error:   clientErr.ErrorID(),
			Message: clientErr.Error(),
		}, nil
	}

	content, err := a.TaskService.EncodeTaskBase64(t.Name)
	if err != nil {
		return nil, err
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
