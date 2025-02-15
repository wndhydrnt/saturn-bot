package integration_test

import (
	"net/http"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
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
						Errors: []openapi.ErrorDetail{
							{Error: 1002, Message: "unknown run"},
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

func Test_API_DeleteRunV1(t *testing.T) {
	testCases := []testCase{
		{
			name:  `deletes a run if it has been scheduled manually`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				// Schedule a new run.
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
				{
					method:       "DELETE",
					path:         "/api/v1/runs/2",
					statusCode:   http.StatusOK,
					responseBody: openapi.DeleteRunV1Response{},
				},
			},
		},

		{
			name:  `does not delete a run if it has not been scheduled manually`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				{
					method:     "DELETE",
					path:       "/api/v1/runs/1",
					statusCode: http.StatusBadRequest,
					responseBody: openapi.DeleteRunV1400JSONResponse{
						Errors: []openapi.ErrorDetail{
							{Error: 1003, Message: "cannot delete run"},
						},
					},
				},
			},
		},

		{
			name:  `does not delete a run if the run does not exist`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				{
					method:     "DELETE",
					path:       "/api/v1/runs/100",
					statusCode: http.StatusNotFound,
					responseBody: openapi.DeleteRunV1400JSONResponse{
						Errors: []openapi.ErrorDetail{
							{Error: 1002, Message: "unknown run"},
						},
					},
				},
			},
		},

		{
			name: `does not delete a run if it has not been scheduled manually but is not pending`,
			tasks: []schema.Task{
				{
					Name: "with-inputs",
					Inputs: []schema.Input{
						{Name: "unit"},
					},
				},
			},
			apiCalls: []apiCall{
				// Schedule a new run.
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						RunData:  ptr.To(map[string]string{"unit": "test"}),
						TaskName: "with-inputs",
					},
					statusCode: http.StatusOK,
					responseBody: openapi.ScheduleRunV1Response{
						RunID: 1,
					},
				},
				// Process the run.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunData: ptr.To(map[string]string{"unit": "test"}),
						RunID:   1,
						Task: openapi.WorkTaskV1{
							Hash: "2b77e497f5d91796abf103724538734cf8ae737ef0a0b134c7b75ebe26e4e2b8",
							Name: "with-inputs",
						},
					},
				},
				{
					method:     "DELETE",
					path:       "/api/v1/runs/1",
					statusCode: http.StatusBadRequest,
					responseBody: openapi.DeleteRunV1400JSONResponse{
						Errors: []openapi.ErrorDetail{
							{Error: 1003, Message: "cannot delete run"},
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
