package api

import (
	"context"
	"errors"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
)

func (a *APIServer) DeleteRunV1(_ context.Context, req openapi.DeleteRunV1RequestObject) (openapi.DeleteRunV1ResponseObject, error) {
	err := a.WorkerService.DeleteRun(req.RunId)
	var clientErr sberror.Client
	if errors.As(err, &clientErr) {
		if clientErr.ErrorID() == sberror.ClientIDRunCannotDelete {
			return openapi.DeleteRunV1400JSONResponse(clientErr.ToApiError()), nil
		}

		return openapi.DeleteRunV1404JSONResponse(clientErr.ToApiError()), nil
	}

	if err != nil {
		return nil, err
	}

	return openapi.DeleteRunV1200JSONResponse{}, nil
}

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
