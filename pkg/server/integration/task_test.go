package integration_test

import (
	"net/http"
	"testing"

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
