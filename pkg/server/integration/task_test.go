package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func TestServer_API_ListTasksV1(t *testing.T) {
	testCases := []testCase{
		{
			name:  `When it receives a request to list tasks then it returns the list of known tasks`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				{
					method:     "GET",
					path:       "/api/v1/tasks",
					statusCode: http.StatusOK,
					responseBody: openapi.ListTasksV1Response{
						Tasks: []string{"unittest"},
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

func TestServer_API_GetTaskV1(t *testing.T) {
	testCases := []testCase{
		{
			name:  `When it receives a request to get one task and the task does exist then it returns the tasks`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				{
					method:     "GET",
					path:       "/api/v1/tasks/unittest",
					statusCode: http.StatusOK,
					responseBody: openapi.GetTaskV1Response{
						Name:    "unittest",
						Hash:    defaultTaskHash,
						Content: defaultTaskContentBase64,
					},
				},
			},
		},

		{
			name:  `When it receives a request to get one task and the task doesn't exist then it indicates that to the client`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				{
					method:       "GET",
					path:         "/api/v1/tasks/unknown",
					statusCode:   http.StatusNotFound,
					responseBody: openapi.Error{Error: "Not Found", Message: "Task unknown"},
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

func TestServer_API_ListTaskRunsV1(t *testing.T) {
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
					path:       fmt.Sprintf("/api/v1/tasks/%s/runs", defaultTask.Name),
					query:      "limit=1&page=1",
					statusCode: http.StatusOK,
					responseBody: openapi.ListTaskRunsV1Response{
						Page: openapi.Page{Next: 2},
						Result: []openapi.TaskRunV1{
							{Id: 2, Reason: openapi.Next, ScheduleAfter: time.Date(2000, 1, 1, 1, 0, 3, 0, time.UTC), Status: openapi.Pending},
						},
					},
				},
				// List the next page of runs.
				{
					method:     "GET",
					path:       fmt.Sprintf("/api/v1/tasks/%s/runs", defaultTask.Name),
					query:      "limit=1&page=2",
					statusCode: http.StatusOK,
					responseBody: openapi.ListTaskRunsV1Response{
						Page: openapi.Page{Next: 0},
						Result: []openapi.TaskRunV1{
							{Id: 1, Reason: openapi.New, ScheduleAfter: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), Status: openapi.Finished},
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
