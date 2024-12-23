package ui

import (
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataIndex struct {
	RecentRuns openapi.ListRunsV1200JSONResponse
	Tasks      []string
}

// GetHome renders the homepage.
func (u *Ui) GetHome(w http.ResponseWriter, r *http.Request) {
	tasksResp, err := u.API.ListTasksV1(r.Context(), openapi.ListTasksV1RequestObject{})
	if err != nil {
		renderError(err, w)
		return
	}

	tplData := dataIndex{}
	tasksObj := tasksResp.(openapi.ListTasksV1200JSONResponse)
	if len(tasksObj.Tasks) > 5 {
		tplData.Tasks = tasksObj.Tasks[0:5]
	} else {
		tplData.Tasks = tasksObj.Tasks
	}

	reqRecent := openapi.ListRunsV1RequestObject{
		Params: openapi.ListRunsV1Params{
			ListOptions: &openapi.ListOptions{
				Limit: 5,
			},
			Status: ptr.To([]openapi.RunStatusV1{openapi.Finished, openapi.Failed}),
		},
	}
	recentRunsResp, err := u.API.ListRunsV1(r.Context(), reqRecent)
	if err != nil {
		renderError(err, w)
		return
	}

	switch recentRunsObj := recentRunsResp.(type) {
	case openapi.ListRunsV1200JSONResponse:
		tplData.RecentRuns = recentRunsObj
	}

	renderTemplate(tplData, w, "home.html")
}