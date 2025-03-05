package api

import (
	"context"
	"errors"
	"strings"
	"time"

	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"go.uber.org/zap"
)

// DeleteRunV1 implements [github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi.ServerInterface].
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

// ListRunsV1 implements [github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi.ServerInterface].
func (a *APIServer) ListRunsV1(ctx context.Context, request openapi.ListRunsV1RequestObject) (openapi.ListRunsV1ResponseObject, error) {
	listOpts := toListOptions(request.Params.ListOptions)
	queryOpts := service.ListRunsOptions{}
	if request.Params.Status != nil {
		for _, apiStatus := range ptr.From(request.Params.Status) {
			dbStatus := mapRunStatusFromApiToDb(apiStatus)
			queryOpts.Status = append(queryOpts.Status, dbStatus)
		}
	}

	if request.Params.Task != nil {
		queryOpts.TaskName = ptr.From(request.Params.Task)
	}

	runs, err := a.WorkerService.ListRuns(queryOpts, &listOpts)
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
			PreviousPage: listOpts.Previous(),
			CurrentPage:  listOpts.Page,
			NextPage:     listOpts.Next(),
			ItemsPerPage: listOpts.Limit,
			TotalItems:   listOpts.TotalItems(),
			TotalPages:   listOpts.TotalPages(),
		},
		Result: result,
	}
	return resp, nil
}

// ScheduleRunV1 implements [github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi.ServerInterface].
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

	var runData map[string]string
	if req.Body.RunData == nil {
		runData = map[string]string{}
	} else {
		runData = *req.Body.RunData
	}

	if req.Body.Assignees != nil {
		runData[sbcontext.RunDataKeyAssignees] = strings.Join(ptr.From(req.Body.Assignees), ",")
	}

	if req.Body.Reviewers != nil {
		runData[sbcontext.RunDataKeyReviewers] = strings.Join(ptr.From(req.Body.Reviewers), ",")
	}

	runID, err := a.WorkerService.ScheduleRun(db.RunReasonManual, repositoryNames, schedulerAfter, req.Body.TaskName, runData, nil)
	if err != nil {
		var clientErr sberror.Client
		if errors.As(err, &clientErr) {
			return openapi.ScheduleRunV1400JSONResponse(clientErr.ToApiError()), nil
		}

		return nil, ErrInternal
	}

	return openapi.ScheduleRunV1200JSONResponse{
		RunID: int(runID), // #nosec G115 -- no info by gosec on how to fix this
	}, nil
}
