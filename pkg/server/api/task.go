package api

import (
	"context"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"go.uber.org/zap"
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

func (a *APIServer) ListTaskRunsV1(_ context.Context, request openapi.ListTaskRunsV1RequestObject) (openapi.ListTaskRunsV1ResponseObject, error) {
	listOpts := toListOptions(request.Params.ListOptions)
	runs, totalCount, err := a.WorkerService.ListRunsOfTask(request.Task, listOpts)
	if err != nil {
		log.Log().Errorw("Failed to list runs of task", zap.Error(err))
		return nil, ErrInternal
	}

	result := make([]openapi.TaskRunV1, len(runs))
	for idx, run := range runs {
		result[idx] = mapRun(run)
	}

	resp := openapi.ListTaskRunsV1200JSONResponse{
		Page: openapi.Page{
			Next: listOpts.Next(int(totalCount)),
		},
		Result: result,
	}
	return resp, nil
}

func mapRun(r db.Run) openapi.TaskRunV1 {
	return openapi.TaskRunV1{
		Id:            int(r.ID),
		Reason:        mapRunReason(r.Reason),
		ScheduleAfter: r.ScheduleAfter,
		Status:        mapRunStatus(r.Status),
	}
}

func mapRunReason(r db.RunReason) openapi.TaskRunV1Reason {
	switch r {
	case db.RunReasonChanged:
		return openapi.Changed
	case db.RunReasonManual:
		return openapi.Manual
	case db.RunReasonNew:
		return openapi.New
	case db.RunReasonWebhook:
		return openapi.Webhook
	default:
		return openapi.Next
	}
}

func mapRunStatus(s db.RunStatus) openapi.TaskRunV1Status {
	switch s {
	case db.RunStatusFailed:
		return openapi.Failed
	case db.RunStatusFinished:
		return openapi.Finished
	case db.RunStatusRunning:
		return openapi.Running
	default:
		return openapi.Pending
	}
}
