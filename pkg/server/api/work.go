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
