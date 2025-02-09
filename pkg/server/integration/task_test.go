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
								State:          openapi.TaskResultStateV1Open,
								PullRequestUrl: ptr.To("http://git.local/unit/test/pr/1"),
								RepositoryName: "git.local/unit/test",
								Result:         11, // processor.ResultPrOpen
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
								State:          openapi.TaskResultStateV1Merged,
								PullRequestUrl: ptr.To("http://git.local/unit/test/pr/1"),
								RepositoryName: "git.local/unit/test",
								Result:         10, // processor.ResultPrMerged
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
	taskWithInputs := schema.Task{
		Name: "unittest-inputs",
		Inputs: []schema.Input{
			{
				Default:     ptr.To("Hello"),
				Description: ptr.To("How to greet."),
				Name:        "greeting",
				Options:     []string{"Hello", "Hallo"},
				Validation:  ptr.To("^Hello|Hallo$"),
			},
		},
	}

	testCases := []testCase{
		{
			name:  `When it receives a request to get one task and the task does exist then it returns the task`,
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
					method:     "GET",
					path:       "/api/v1/tasks/unknown",
					statusCode: http.StatusNotFound,
					responseBody: openapi.Error{
						Errors: []openapi.ErrorDetail{
							{Error: sberror.ClientIDTaskNotFound, Message: "unknown task"},
						},
					},
				},
			},
		},

		{
			name:  `task with inputs`,
			tasks: []schema.Task{taskWithInputs},
			apiCalls: []apiCall{
				{
					method:     "GET",
					path:       fmt.Sprintf("/api/v1/tasks/%s", taskWithInputs.Name),
					statusCode: http.StatusOK,
					responseBody: openapi.GetTaskV1Response{
						Name:    "unittest-inputs",
						Hash:    "a3dc7adc92a9139d193c0b4b622e9587038bfc9449763c3328979f934dc9fcd8",
						Content: "aW5wdXRzOgogIC0gZGVmYXVsdDogSGVsbG8KICAgIGRlc2NyaXB0aW9uOiBIb3cgdG8gZ3JlZXQuCiAgICBuYW1lOiBncmVldGluZwogICAgb3B0aW9uczoKICAgICAgLSBIZWxsbwogICAgICAtIEhhbGxvCiAgICB2YWxpZGF0aW9uOiBeSGVsbG98SGFsbG8kCm5hbWU6IHVuaXR0ZXN0LWlucHV0cwo=",
						Inputs: ptr.To([]openapi.TaskV1Input{
							{
								Default:     ptr.To("Hello"),
								Description: ptr.To("How to greet."),
								Name:        "greeting",
								Options:     ptr.To([]string{"Hello", "Hallo"}),
								Validation:  ptr.To("^Hello|Hallo$"),
							},
						}),
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
