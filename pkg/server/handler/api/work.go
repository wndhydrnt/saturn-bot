package api

import (
	"context"
	"math/rand"
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/task"
)

type WorkerService struct {
	tasks []task.Task
}

func NewWorkerService(tasks []task.Task) *WorkerService {
	return &WorkerService{tasks: tasks}
}

func (ws *WorkerService) GetWorkV1(_ context.Context) (openapi.ImplResponse, error) {
	if len(ws.tasks) == 0 {
		return openapi.Response(http.StatusInternalServerError, serverError), nil
	}

	body := openapi.GetWorkV1Response{
		RunID:      genIDInt(6),
		Repository: "gitlab.com/wandhydrant/rcmt-test",
		Tasks: []openapi.GetWorkV1Task{
			{Hash: ws.tasks[0].Hash, Name: ws.tasks[0].TaskName},
		},
	}
	return openapi.Response(http.StatusOK, body), nil
}

func (ws *WorkerService) ReportWorkV1(_ context.Context, req openapi.ReportWorkV1Request) (openapi.ImplResponse, error) {
	body := openapi.ReportWorkV1201Response{
		Result: "ok",
	}
	return openapi.ImplResponse{Code: http.StatusCreated, Body: body}, nil
}

func genIDInt(n int32) int32 {
	return rand.Int31n(n * 1000)
}
