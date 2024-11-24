package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func TestServer_API_ListRunsV1(t *testing.T) {
	testCases := []testCase{
		{
			name: `Given two runs of a task
								When a request with limit=1&page=1 is sent
								And a request with limit=1&page=2 is sent
								Then it returns the latest run
								And it returns the previous run`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				// Read the run that gets scheduled at the start of the server.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Tasks: []openapi.GetWorkV1Task{
							{Hash: defaultTaskHash, Name: defaultTask.Name},
						},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID:       1,
						TaskResults: []openapi.ReportWorkV1TaskResult{},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// List the runs of the task. Limit to one result to test pagination.
				{
					method:     "GET",
					path:       "/api/v1/worker/runs",
					query:      fmt.Sprintf("limit=1&page=1&task=%s", defaultTask.Name),
					statusCode: http.StatusOK,
					responseBody: openapi.ListRunsV1Response{
						Page: openapi.Page{Next: 2},
						Result: []openapi.RunV1{
							{
								Id:            2,
								Reason:        openapi.Next,
								ScheduleAfter: time.Date(2000, 1, 1, 1, 0, 41, 0, time.UTC),
								Status:        openapi.Pending,
								Task:          defaultTask.Name,
							},
						},
					},
				},
				// List the next page of runs.
				{
					method:     "GET",
					path:       "/api/v1/worker/runs",
					query:      fmt.Sprintf("limit=1&page=2&task=%s", defaultTask.Name),
					statusCode: http.StatusOK,
					responseBody: openapi.ListRunsV1Response{
						Page: openapi.Page{Next: 0},
						Result: []openapi.RunV1{
							{
								FinishedAt:    ptr.To(time.Date(2000, 1, 1, 0, 0, 40, 0, time.UTC)),
								Id:            1,
								Reason:        openapi.New,
								ScheduleAfter: time.Date(2000, 1, 1, 0, 0, 37, 0, time.UTC),
								StartedAt:     ptr.To(time.Date(2000, 1, 1, 0, 0, 39, 0, time.UTC)),
								Status:        openapi.Finished,
								Task:          defaultTask.Name,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executeTestCase(t, tc)
		})
	}
}