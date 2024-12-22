package api

import (
	"context"
	"errors"

	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
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

func (a *APIServer) ListTaskResultsV1(ctx context.Context, request openapi.ListTaskResultsV1RequestObject) (openapi.ListTaskResultsV1ResponseObject, error) {
	opts := service.ListTaskResultsOptions{}
	if request.Params.RunId != nil {
		opts.RunId = ptr.From(request.Params.RunId)
	}

	if request.Params.Status != nil {
		for _, apiStatus := range ptr.From(request.Params.Status) {
			opts.Status = append(opts.Status, db.TaskResultStatus(apiStatus))
		}
	}

	listOpts := toListOptions(request.Params.ListOptions)
	taskResults, err := a.WorkerService.ListTaskResults(opts, &listOpts)
	if err != nil {
		return nil, err
	}

	resp := openapi.ListTaskResultsV1200JSONResponse{
		Page: openapi.Page{
			PreviousPage: listOpts.Previous(),
			CurrentPage:  listOpts.Page,
			NextPage:     listOpts.Next(),
			ItemsPerPage: listOpts.Limit,
			TotalItems:   listOpts.TotalItems(),
			TotalPages:   listOpts.TotalPages(),
		},
	}
	for _, tr := range taskResults {
		resp.TaskResults = append(resp.TaskResults, mapTaskResultFromDbToApi(tr))
	}

	return resp, nil
}

func mapTaskResultFromDbToApi(db db.TaskResult) openapi.TaskResultV1 {
	api := openapi.TaskResultV1{
		RepositoryName: db.RepositoryName,
		RunId:          int(db.RunID),
		Status:         openapi.TaskResultStatusV1(db.Status),
	}
	if db.Error != nil {
		api.Error = db.Error
	}

	return api
}
