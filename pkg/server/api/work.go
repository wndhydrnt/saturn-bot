package api

import (
	"context"
	"errors"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"go.uber.org/zap"
)

func (a *APIServer) GetWorkV1(ctx context.Context, _ openapi.GetWorkV1RequestObject) (openapi.GetWorkV1ResponseObject, error) {
	resp := openapi.GetWorkV1200JSONResponse{}
	run, task, err := a.WorkerService.NextRun()
	if err != nil {
		if errors.Is(err, service.ErrNoRun) {
			return resp, nil
		}

		log.Log().Errorw("Failed to get next run", zap.Error(err))
		return resp, ErrInternal
	}

	if len(run.RepositoryNames) > 0 {
		resp.Repositories = ptr.To([]string(run.RepositoryNames))
	}

	resp.RunID = int(run.ID) // #nosec G115 -- no info by gosec on how to fix this
	resp.Tasks = []openapi.GetWorkV1Task{
		{Hash: task.Checksum(), Name: task.Task.Name},
	}

	return resp, nil
}

func (a *APIServer) ListRunsV1(ctx context.Context, request openapi.ListRunsV1RequestObject) (openapi.ListRunsV1ResponseObject, error) {
	listOpts := toListOptions(request.Params.ListOptions)
	queryOpts := service.ListRunsOptions{}
	if request.Params.Task != nil {
		queryOpts.TaskName = ptr.From(request.Params.Task)
	}

	runs, totalCount, err := a.WorkerService.ListRuns(queryOpts, listOpts)
	if err != nil {
		log.Log().Errorw("Failed to list runs of task", zap.Error(err))
		return nil, ErrInternal
	}

	result := make([]openapi.RunV1, len(runs))
	for idx, run := range runs {
		result[idx] = mapRun(run)
	}

	resp := openapi.ListRunsV1200JSONResponse{
		Page: openapi.Page{
			Next:  listOpts.Next(int(totalCount)),
			Total: int(totalCount),
		},
		Result: result,
	}
	return resp, nil
}

func (a *APIServer) ReportWorkV1(_ context.Context, request openapi.ReportWorkV1RequestObject) (openapi.ReportWorkV1ResponseObject, error) {
	resp := openapi.ReportWorkV1201JSONResponse{}
	err := a.WorkerService.ReportRun(*request.Body)
	if err != nil {
		return resp, ErrInternal
	}

	resp.Result = "ok"
	return resp, nil
}

func (a *APIServer) ScheduleRunV1(_ context.Context, req openapi.ScheduleRunV1RequestObject) (openapi.ScheduleRunV1ResponseObject, error) {
	var schedulerAfter time.Time
	if req.Body.ScheduleAfter == nil {
		schedulerAfter = a.Clock.Now()
	} else {
		schedulerAfter = *req.Body.ScheduleAfter
	}

	var repositoryNames []string
	if req.Body.RepositoryNames != nil {
		repositoryNames = *req.Body.RepositoryNames
	}

	runID, err := a.WorkerService.ScheduleRun(db.RunReasonManual, repositoryNames, schedulerAfter, req.Body.TaskName, nil)
	if err != nil {
		return nil, ErrInternal
	}

	return openapi.ScheduleRunV1200JSONResponse{
		RunID: int(runID), // #nosec G115 -- no info by gosec on how to fix this
	}, nil
}

func mapRun(r db.Run) openapi.RunV1 {
	run := openapi.RunV1{
		FinishedAt:    r.FinishedAt,
		Id:            r.ID,
		Reason:        mapRunReason(r.Reason),
		ScheduleAfter: r.ScheduleAfter,
		StartedAt:     r.StartedAt,
		Status:        mapRunStatus(r.Status),
		Task:          r.TaskName,
	}
	if len(r.RepositoryNames) > 0 {
		run.Repositories = ptr.To([]string(r.RepositoryNames))
	}

	return run
}

func mapRunReason(r db.RunReason) openapi.RunV1Reason {
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

func mapRunStatus(s db.RunStatus) openapi.RunV1Status {
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
