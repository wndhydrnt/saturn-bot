package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
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

func TestServer_API_ListTaskRecentTaskResultsV1(t *testing.T) {
	testCases := []testCase{
		{
			name:  `Returns the latest results per repository for a task`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				// Schedule the first run.
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						TaskName: defaultTask.Name,
					},
					statusCode: http.StatusOK,
					responseBody: openapi.ScheduleRunV1Response{
						RunID: 2,
					},
				},
				// And report the result of the first run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID: 2,
						Task: openapi.WorkTaskV1{
							Name: defaultTask.Name,
						},
						TaskResults: []openapi.ReportWorkV1TaskResult{
							{
								PullRequestState: ptr.To(openapi.TaskResultStatusV1Open),
								PullRequestUrl:   ptr.To("http://git.local/unit/test/pr/1"),
								RepositoryName:   "git.local/unit/test",
								Result:           11, // processor.ResultPrOpen
							},
						},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// Schedule a second run.
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						TaskName: defaultTask.Name,
					},
					statusCode: http.StatusOK,
					responseBody: openapi.ScheduleRunV1Response{
						RunID: 3,
					},
				},
				// Read the second run.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 3,
						Task:  openapi.WorkTaskV1{Hash: defaultTaskHash, Name: defaultTask.Name},
					},
				},
				// Report the result of the second run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID: 3,
						Task: openapi.WorkTaskV1{
							Name: defaultTask.Name,
						},
						TaskResults: []openapi.ReportWorkV1TaskResult{
							{
								PullRequestState: ptr.To(openapi.TaskResultStatusV1Merged),
								PullRequestUrl:   ptr.To("http://git.local/unit/test/pr/1"),
								RepositoryName:   "git.local/unit/test",
								Result:           10, // processor.ResultPrMerged
							},
						},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// List the latest results.
				{
					method:     "GET",
					path:       fmt.Sprintf("/api/v1/tasks/%s/results", defaultTask.Name),
					statusCode: http.StatusOK,
					responseBody: openapi.ListTaskRecentTaskResultsV1Response{
						Page: openapi.Page{
							CurrentPage:  1,
							ItemsPerPage: 20,
							TotalItems:   1,
							TotalPages:   1,
						},
						TaskResults: []openapi.TaskResultV1{
							{
								PullRequestUrl: ptr.To("http://git.local/unit/test/pr/1"),
								RepositoryName: "git.local/unit/test",
								RunId:          3,
								Status:         "merged",
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
					responseBody: openapi.Error{Error: sberror.ClientIDTaskNotFound, Message: "unknown task: unknown"},
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
