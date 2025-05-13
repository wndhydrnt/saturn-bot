package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataRunsRepositoryErrorShow struct {
	Error          string
	RepositoryName string
	RunId          int
}

func (u *Ui) RunsRepositoryErrorShow(w http.ResponseWriter, r *http.Request) {
	runId, err := strconv.Atoi(chi.URLParam(r, "runId"))
	if err != nil {
		renderError(fmt.Errorf("convert parameter runId to int: %w", err), w)
		return
	}

	repositoryNameRaw := chi.URLParam(r, "repositoryName")
	if repositoryNameRaw == "" {
		renderError(fmt.Errorf("cannot extract repositoryName path parameter"), w)
		return
	}

	repositoryName, err := url.PathUnescape(repositoryNameRaw)
	if err != nil {
		renderError(fmt.Errorf("unescape repository name: %w", err), w)
	}

	listTaskResultsResp, err := u.API.ListTaskResultsV1(r.Context(), openapi.ListTaskResultsV1RequestObject{
		Params: openapi.ListTaskResultsV1Params{
			RepositoryName: ptr.To(repositoryName),
			RunId:          ptr.To(runId),
		},
	})
	if err != nil {
		renderError(err, w)
		return
	}

	switch respObj := listTaskResultsResp.(type) {
	case openapi.ListTaskResultsV1200JSONResponse:
		if len(respObj.TaskResults) != 1 {
			renderError(fmt.Errorf("expected 1 task result, got %d", len(respObj.TaskResults)), w)
			return
		}

		if respObj.TaskResults[0].Error == nil {
			renderError(fmt.Errorf("no error for %s", repositoryName), w)
			return
		}

		data := dataRunsRepositoryErrorShow{
			Error:          ptr.From(respObj.TaskResults[0].Error),
			RepositoryName: respObj.TaskResults[0].RepositoryName,
			RunId:          respObj.TaskResults[0].RunId,
		}
		renderTemplate(data, w, "runs_repository_error_show.html")
	default:
		renderError(fmt.Errorf("expected ListTaskResultsV1200JSONResponse, got %T", respObj), w)
	}
}
