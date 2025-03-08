package ui

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataResultsIndexFilters struct {
	TaskResultStatusCurrent string
	TaskResultStatusList    []openapi.TaskResultStateV1
}

type dataResultsIndex struct {
	DisplayRunLink bool
	Filters        dataResultsIndexFilters
	Pagination     pagination
	TaskName       string
	TaskResults    []openapi.TaskResultV1
}

// ResultsIndex renders the list of results of the latest run of a task.
func (u *Ui) ResultsIndex(w http.ResponseWriter, r *http.Request) {
	statusParam := r.URL.Query().Get("status")
	data := dataResultsIndex{
		DisplayRunLink: true,
		Filters: dataResultsIndexFilters{
			TaskResultStatusCurrent: statusParam,
			TaskResultStatusList:    taskResultStatusOptions,
		},
	}

	name := chi.URLParam(r, "name")
	listTaskResultsReq := openapi.ListTaskRecentTaskResultsV1RequestObject{
		Task: name,
		Params: openapi.ListTaskRecentTaskResultsV1Params{
			ListOptions: &openapi.ListOptions{
				Limit: parseIntParam(r, "limit", 10),
				Page:  parseIntParam(r, "page", 1),
			},
		},
	}

	if statusParam != "" {
		listTaskResultsReq.Params.Status = ptr.To([]openapi.TaskResultStateV1{openapi.TaskResultStateV1(statusParam)})
	}

	listTaskResultsResp, err := u.API.ListTaskRecentTaskResultsV1(r.Context(), listTaskResultsReq)
	if err != nil {
		renderError(err, w)
		return
	}

	switch resp := listTaskResultsResp.(type) {
	case openapi.ListTaskRecentTaskResultsV1200JSONResponse:
		data.Pagination = pagination{
			Page: resp.Page,
			URL:  r.URL,
		}
		data.TaskName = name
		data.TaskResults = resp.TaskResults
		renderTemplate(data, w, "results_table.html", "results_index.html")
	case openapi.ListTaskRecentTaskResultsV1404JSONResponse:
		renderApiError(openapi.Error(resp), w, http.StatusNotFound, "")
	case openapi.ListTaskRecentTaskResultsV1500JSONResponse:
		renderApiError(openapi.Error(resp), w, http.StatusInternalServerError, "")
	}
}
