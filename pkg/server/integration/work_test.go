package integration_test

import (
	"fmt"
	"net/http"
	"testing"

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
								ScheduleAfter: testDate(1, 0, 4),
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
								FinishedAt:    ptr.To(testDate(0, 0, 3)),
								Id:            1,
								Reason:        openapi.New,
								ScheduleAfter: testDate(0, 0, 0),
								StartedAt:     ptr.To(testDate(0, 0, 2)),
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

func TestServer_API_GetWorkV1(t *testing.T) {
	testCases := []testCase{
		{
			name: `Given two tasks
							When a worker requests work
							Then it returns the task with the oldest schedule timestamp`,
			tasks: []schema.Task{
				{Name: "unittest 1"},
				{Name: "unittest 2"},
			},
			apiCalls: []apiCall{
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Tasks: []openapi.GetWorkV1Task{
							{Hash: "ab5a03b44faf542081c9b54eab3ce7c10731b917ebca511b28b7723258ad49b2", Name: "unittest 1"},
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
