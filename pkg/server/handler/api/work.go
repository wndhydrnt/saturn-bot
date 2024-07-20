package api

import (
	"context"
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api/openapi"
)

type WorkerService struct{}

func (ws *WorkerService) GetWorkV1(_ context.Context) (openapi.ImplResponse, error) {
	body := openapi.GetWorkV1200Response{
		ExecutionID: 123,
		Repository:  "gitlab.com/wandhydrant/rcmt-test",
		Tasks:       []string{"a", "b"},
	}
	return openapi.Response(http.StatusOK, body), nil
}

func (ws *WorkerService) ReportWorkV1(_ context.Context, req openapi.ReportWorkV1Request) (openapi.ImplResponse, error) {
	body := openapi.ReportWorkV1201Response{
		Result: "ok",
	}
	return openapi.ImplResponse{Code: http.StatusCreated, Body: body}, nil
}
