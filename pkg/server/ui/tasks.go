package ui

import (
	"fmt"
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataListTasks struct {
	Tasks []string
}

func (u *Ui) ListTasks(w http.ResponseWriter, r *http.Request) {
	listTasksResp, err := u.API.ListTasksV1(r.Context(), openapi.ListTasksV1RequestObject{})
	if err != nil {
		renderError(fmt.Errorf("list tasks: %w", err), w)
		return
	}

	taskList := listTasksResp.(openapi.ListTasksV1200JSONResponse)
	var data dataListTasks
	data.Tasks = taskList.Tasks
	renderTemplate("task-list.html", data, w)
}
