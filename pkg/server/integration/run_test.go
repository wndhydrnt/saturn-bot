package integration_test

import (
	"net/http"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func Test_API_GetRunV1(t *testing.T) {
	testCases := []testCase{
		{
			name:  `When the run exists then it returns the run`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				{
					method:     "GET",
					path:       "/api/v1/runs/1",
					statusCode: http.StatusOK,
					responseBody: openapi.GetRunV1Response{
						Run: openapi.RunV1{
							Id:            1,
							Reason:        openapi.Cron,
							ScheduleAfter: testDate(1, 6, 3, 0),
							Status:        openapi.Pending,
							Task:          defaultTask.Name,
						},
					},
				},
			},
		},

		{
			name:  `When the run does not exist then it is not found`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				{
					method:     "GET",
					path:       "/api/v1/runs/100",
					statusCode: http.StatusNotFound,
					responseBody: openapi.Error{
						Error:   1002,
						Message: "unknown run with ID 100",
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
