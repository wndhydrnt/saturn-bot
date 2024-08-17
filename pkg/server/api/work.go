package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"go.uber.org/zap"
)

type WorkHandler struct {
	WorkerService *service.WorkerService
}

func (wh *WorkHandler) GetWorkV1(_ context.Context) (openapi.ImplResponse, error) {
	run, task, err := wh.WorkerService.NextRun()
	if err != nil {
		if errors.Is(err, service.ErrNoRun) {
			body := openapi.GetWorkV1Response{}
			return openapi.Response(http.StatusOK, body), nil
		}

		log.Log().Errorw("Failed to get next run", zap.Error(err))
		return openapi.Response(http.StatusInternalServerError, serverError), nil
	}

	body := openapi.GetWorkV1Response{
		RunID: int32(run.ID),
		Tasks: []openapi.GetWorkV1Task{
			{Hash: task.Hash, Name: task.TaskName},
		},
	}
	if run.RepositoryName != nil {
		body.Repository = *run.RepositoryName
	}

	return openapi.Response(http.StatusOK, body), nil
}

func (wh *WorkHandler) ReportWorkV1(_ context.Context, req openapi.ReportWorkV1Request) (openapi.ImplResponse, error) {
	err := wh.WorkerService.ReportRun(req)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, serverError), nil
	}

	body := openapi.ReportWorkV1Response{
		Result: "ok",
	}
	return openapi.ImplResponse{Code: http.StatusCreated, Body: body}, nil
}

func (wh *WorkHandler) ScheduleRunV1(_ context.Context, req openapi.ScheduleRunV1Request) (openapi.ImplResponse, error) {
	runID, err := wh.WorkerService.ScheduleRun(db.RunReasonManual, req.RepositoryName, req.ScheduleAfter, req.TaskName, nil)
	if err != nil {
		return openapi.Response(http.StatusInternalServerError, serverError), nil
	}

	body := openapi.ScheduleRunV1Response{RunID: int32(runID)}
	return openapi.ImplResponse{Code: http.StatusCreated, Body: body}, nil
}
