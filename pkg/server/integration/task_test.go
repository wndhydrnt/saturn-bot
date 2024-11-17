package integration_test

import (
	"net/http"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func TestServer_TaskAPI(t *testing.T) {
	testCases := []testCase{
		{
			name:  `When it receives a request to list tasks then it returns a list of tasks`,
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executeTestCase(t, tc)
		})
	}
}
