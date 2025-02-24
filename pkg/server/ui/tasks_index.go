package ui

import (
	"fmt"
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataListTasks struct {
	Filters struct {
		ActiveCurrent string
	}
	Pagination pagination
	Tasks      []openapi.ListTasksV1ResponseTask
}

// TasksIndex renders the list of all tasks known to saturn-bot.
func (u *Ui) TasksIndex(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r, "limit", 10)
	page := parseIntParam(r, "page", 1)
	req := openapi.ListTasksV1RequestObject{
		Params: openapi.ListTasksV1Params{
			Active: ptr.To(true),
			ListOptions: &openapi.ListOptions{
				Limit: limit,
				Page:  page,
			},
		},
	}
	queryActive := r.URL.Query().Get("active")
	if queryActive == "" || queryActive == "true" {
		req.Params.Active = ptr.To(true)
	}

	if queryActive == "false" {
		req.Params.Active = ptr.To(false)
	}

	listTasksResp, err := u.API.ListTasksV1(r.Context(), req)
	if err != nil {
		renderError(fmt.Errorf("list tasks: %w", err), w)
		return
	}

	taskList := listTasksResp.(openapi.ListTasksV1200JSONResponse)
	var data dataListTasks
	data.Filters = struct{ ActiveCurrent string }{queryActive}
	data.Tasks = taskList.Results
	data.Pagination = pagination{
		Page: taskList.Page,
		URL:  r.URL,
	}
	renderTemplate(data, w, "tasks_index.html")
}
