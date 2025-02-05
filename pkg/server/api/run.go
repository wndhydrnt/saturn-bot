package api

import (
	"context"
	"errors"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
)

// GetRunV1 implements openapi.ServerInterface.
func (a *APIServer) GetRunV1(_ context.Context, req openapi.GetRunV1RequestObject) (openapi.GetRunV1ResponseObject, error) {
	run, err := a.WorkerService.GetRun(req.RunId)
	var clientErr sberror.Client
	if errors.As(err, &clientErr) {
		return openapi.GetRunV1404JSONResponse(clientErr.ToApiError()), nil
	}

	if err != nil {
		return nil, err
	}

	return openapi.GetRunV1200JSONResponse{
		Run: mapRun(run),
	}, nil
}
