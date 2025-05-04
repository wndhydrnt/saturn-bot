package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

// GetTaskV1 implements [openapi.ServerInterface].
func (a *APIServer) GetTaskV1(_ context.Context, request openapi.GetTaskV1RequestObject) (openapi.GetTaskV1ResponseObject, error) {
	t, err := a.TaskService.GetTask(request.Task)
	var clientErr sberror.Client
	if errors.As(err, &clientErr) {
		return openapi.GetTaskV1404JSONResponse{
			Errors: []openapi.ErrorDetail{
				{Error: clientErr.ErrorID(), Message: clientErr.Error()},
			},
		}, nil
	}

	content, err := a.TaskService.EncodeTaskBase64(t.Name)
	if err != nil {
		return nil, err
	}

	resp := openapi.GetTaskV1200JSONResponse{
		Name:    t.Name,
		Hash:    t.Checksum(),
		Content: content,
	}

	if len(t.Inputs) > 0 {
		var inputs []openapi.TaskV1Input
		for _, tinput := range t.Inputs {
			inputs = append(inputs, mapTaskInputToApi(tinput))
		}

		resp.Inputs = ptr.To(inputs)
	}

	return resp, nil
}

// ListTasksV1 implements [openapi.ServerInterface].
func (th *APIServer) ListTasksV1(_ context.Context, request openapi.ListTasksV1RequestObject) (openapi.ListTasksV1ResponseObject, error) {
	resp := openapi.ListTasksV1200JSONResponse{
		Results: []openapi.ListTasksV1ResponseTask{},
	}
	listOpts := toListOptions(request.Params.ListOptions)
	tasks, err := th.TaskService.ListTasksFromDatabase(service.ListTasksFromDatabaseOptions{
		Active: request.Params.Active,
	}, &listOpts)
	if err != nil {
		return nil, fmt.Errorf("ListTasksV1: %w", err)
	}

	for _, entry := range tasks {
		resp.Results = append(resp.Results, openapi.ListTasksV1ResponseTask{
			Active:   entry.Active,
			Checksum: entry.Hash,
			Name:     entry.Name,
		})
	}

	resp.Page = openapi.Page{
		PreviousPage: listOpts.Previous(),
		CurrentPage:  listOpts.Page,
		NextPage:     listOpts.Next(),
		ItemsPerPage: listOpts.Limit,
		TotalItems:   listOpts.TotalItems(),
		TotalPages:   listOpts.TotalPages(),
	}
	return resp, nil
}

// ListTaskResultsV1 implements [openapi.ServerInterface].
func (a *APIServer) ListTaskResultsV1(ctx context.Context, request openapi.ListTaskResultsV1RequestObject) (openapi.ListTaskResultsV1ResponseObject, error) {
	opts := service.ListTaskResultsOptions{}
	if request.Params.RepositoryName != nil {
		opts.RepositoryName = ptr.From(request.Params.RepositoryName)
	}

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
		TaskResults: []openapi.TaskResultV1{},
	}
	for _, tr := range taskResults {
		resp.TaskResults = append(resp.TaskResults, mapTaskResultFromDbToApi(tr))
	}

	return resp, nil
}

// ListTaskRecentTaskResultsV1 lists recent run results of a task by repository.
func (a *APIServer) ListTaskRecentTaskResultsV1(ctx context.Context, request openapi.ListTaskRecentTaskResultsV1RequestObject) (openapi.ListTaskRecentTaskResultsV1ResponseObject, error) {
	opts := service.ListRecentTaskResultsByTaskOptions{
		TaskName: request.Task,
	}
	if request.Params.Status != nil {
		for _, apiStatus := range ptr.From(request.Params.Status) {
			opts.Status = append(opts.Status, db.TaskResultStatus(apiStatus))
		}
	}

	listOpts := toListOptions(request.Params.ListOptions)
	taskResults, err := a.TaskService.ListRecentTaskResultsByTask(opts, &listOpts)
	if err != nil {
		var clientErr sberror.Client
		if errors.As(err, &clientErr) {
			return openapi.ListTaskRecentTaskResultsV1404JSONResponse(clientErr.ToApiError()), nil
		}

		return nil, err
	}

	resp := openapi.ListTaskRecentTaskResultsV1200JSONResponse{
		Page: openapi.Page{
			PreviousPage: listOpts.Previous(),
			CurrentPage:  listOpts.Page,
			NextPage:     listOpts.Next(),
			ItemsPerPage: listOpts.Limit,
			TotalItems:   listOpts.TotalItems(),
			TotalPages:   listOpts.TotalPages(),
		},
		TaskResults: []openapi.TaskResultV1{},
	}
	for _, tr := range taskResults {
		resp.TaskResults = append(resp.TaskResults, mapTaskResultFromDbToApi(tr))
	}

	return resp, nil
}

func mapTaskResultFromDbToApi(db db.TaskResult) openapi.TaskResultV1 {
	api := openapi.TaskResultV1{
		RepositoryName: db.RepositoryName,
		RunId:          int(db.RunID), // #nosec G115 -- no info by gosec on how to fix this
		Status:         openapi.TaskResultStateV1(db.Status),
	}
	if db.Error != nil {
		api.Error = db.Error
	}

	if db.PullRequestUrl != nil {
		api.PullRequestUrl = db.PullRequestUrl
	}

	return api
}

func mapTaskInputToApi(i schema.Input) openapi.TaskV1Input {
	a := openapi.TaskV1Input{
		Default:     i.Default,
		Description: i.Description,
		Name:        i.Name,
		Validation:  i.Validation,
	}

	if len(i.Options) > 0 {
		a.Options = ptr.To(i.Options)
	}

	return a
}
