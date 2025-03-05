package api

import (
	"context"
	"errors"

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

	if len(run.RunData) > 0 {
		resp.RunData = ptr.To(map[string]string(run.RunData))
	}

	resp.RunID = int(run.ID) // #nosec G115 -- no info by gosec on how to fix this
	resp.Task = openapi.WorkTaskV1{Hash: task.Checksum(), Name: task.Task.Name}
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

func mapRun(r db.Run) openapi.RunV1 {
	run := openapi.RunV1{
		Error:         r.Error,
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

	if len(r.RunData) > 0 {
		run.RunData = ptr.To(map[string]string(r.RunData))
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
	case db.RunReasonCron:
		return openapi.Cron
	default:
		return openapi.Next
	}
}

func mapRunStatus(s db.RunStatus) openapi.RunStatusV1 {
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

func mapRunStatusFromApiToDb(rs openapi.RunStatusV1) db.RunStatus {
	switch rs {
	case openapi.Failed:
		return db.RunStatusFailed
	case openapi.Finished:
		return db.RunStatusFinished
	case openapi.Pending:
		return db.RunStatusPending
	default:
		return db.RunStatusRunning
	}
}
