package ui

import (
	"context"
	"net/http"

	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type dataIndex struct {
	RecentRuns   openapi.ListRunsV1200JSONResponse
	UpcomingRuns openapi.ListRunsV1200JSONResponse
}

// GetHome renders the homepage.
func (u *Ui) GetHome(w http.ResponseWriter, r *http.Request) {
	reqUpcoming := openapi.ListRunsV1RequestObject{
		Params: openapi.ListRunsV1Params{
			ListOptions: &openapi.ListOptions{
				Limit: 5,
			},
			Status: ptr.To([]openapi.RunStatusV1{openapi.Pending}),
		},
	}
	respUpcoming, err := u.API.ListRunsV1(context.Background(), reqUpcoming)
	if err != nil {
		renderError(err, w)
		return
	}

	tplData := dataIndex{}
	switch payload := respUpcoming.(type) {
	case openapi.ListRunsV1200JSONResponse:
		tplData.UpcomingRuns = payload
	}

	reqRecent := openapi.ListRunsV1RequestObject{
		Params: openapi.ListRunsV1Params{
			ListOptions: &openapi.ListOptions{
				Limit: 5,
			},
			Status: ptr.To([]openapi.RunStatusV1{openapi.Finished, openapi.Failed}),
		},
	}
	respRecent, err := u.API.ListRunsV1(context.Background(), reqRecent)
	if err != nil {
		renderError(err, w)
		return
	}

	switch payload := respRecent.(type) {
	case openapi.ListRunsV1200JSONResponse:
		tplData.RecentRuns = payload
	}

	renderTemplate("home.html", tplData, w)
}
