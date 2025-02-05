package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/processor"
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
				// Process the run.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 2,
						Task: openapi.WorkTaskV1{
							Hash: defaultTaskHash,
							Name: defaultTask.Name,
						},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID: 2,
						Task: openapi.WorkTaskV1{
							Name: defaultTask.Name,
						},
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
						Page: openapi.Page{CurrentPage: 1, NextPage: 2, ItemsPerPage: 1, TotalItems: 2, TotalPages: 2},
						Result: []openapi.RunV1{
							{
								Id:            1,
								Reason:        openapi.Cron,
								ScheduleAfter: testDate(1, 6, 3, 0),
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
						Page: openapi.Page{PreviousPage: 1, CurrentPage: 2, NextPage: 0, ItemsPerPage: 1, TotalItems: 2, TotalPages: 2},
						Result: []openapi.RunV1{
							{
								FinishedAt:    ptr.To(testDate(1, 0, 0, 4)),
								Id:            2,
								Reason:        openapi.Manual,
								ScheduleAfter: testDate(1, 0, 0, 1),
								StartedAt:     ptr.To(testDate(1, 0, 0, 3)),
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
				// Schedule a new run for the first task.
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						TaskName: "unittest 1",
					},
					statusCode: http.StatusOK,
					responseBody: openapi.ScheduleRunV1Response{
						RunID: 1,
					},
				},
				// Schedule a new run for the second task.
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						TaskName: "unittest 2",
					},
					statusCode: http.StatusOK,
					responseBody: openapi.ScheduleRunV1Response{
						RunID: 2,
					},
				},
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "5ac498db72aa17c5f0c213781e3b18a9330db1bdf934010e4489621c7d9ec422",
							Name: "unittest 1",
						},
					},
				},
			},
		},

		{
			name: `Given a task that expects an input
							And a run of that task is scheduled manually
							When a worker requests work
							Then it returns the task`,
			tasks: []schema.Task{
				{Name: "unittest", Inputs: []schema.Input{{Name: "example"}}},
			},
			apiCalls: []apiCall{
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						TaskName: "unittest",
						RunData:  ptr.To(map[string]string{"example": "data"}),
					},
					statusCode: http.StatusOK,
					responseBody: openapi.ScheduleRunV1Response{
						RunID: 1,
					},
				},
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID:   1,
						RunData: ptr.To(map[string]string{"example": "data"}),
						Task: openapi.WorkTaskV1{
							Hash: "b8113667c6c0dd4ab59129f042cbb6ca63ee905d8e32a7f684fc94b54dd68613",
							Name: "unittest",
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

func TestServer_API_ScheduleRunV1(t *testing.T) {
	testCases := []testCase{
		{
			name: `When scheduling a run for an unknown task
							Then it returns the info that the task does not exist`,
			tasks: []schema.Task{
				{Name: "unittest"},
			},
			apiCalls: []apiCall{
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						TaskName: "other",
					},
					statusCode: http.StatusBadRequest,
					responseBody: openapi.Error{
						Errors: []openapi.ErrorDetail{
							{Error: 1000, Message: "unknown task"},
						},
					},
				},
			},
		},

		{
			name: `Given a task that requires an input
							When scheduling a run for that task without providing values for those inputs
							Then it returns the info that required inputs are missing`,
			tasks: []schema.Task{
				{
					Name: "unittest",
					Inputs: []schema.Input{
						{Name: "greeting"},
					},
				},
			},
			apiCalls: []apiCall{
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						TaskName: "unittest",
					},
					statusCode: http.StatusBadRequest,
					responseBody: openapi.Error{
						Errors: []openapi.ErrorDetail{
							{Error: 1001, Message: "missing required input", Detail: ptr.To("missing value for input 'greeting'")},
						},
					},
				},
			},
		},

		{
			name: `Given a task
							When scheduling a run that sets all fields possible
							Then it schedules the run`,
			tasks: []schema.Task{
				{
					Name: "unittest",
					Inputs: []schema.Input{
						{Name: "greeting"},
					},
				},
			},
			apiCalls: []apiCall{
				// Schedule a new run
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						Assignees:       ptr.To([]string{"ellie"}),
						RepositoryNames: ptr.To([]string{"git.local/unit/test"}),
						Reviewers:       ptr.To([]string{"abby"}),
						RunData:         ptr.To(map[string]string{"greeting": "Hello"}),
						ScheduleAfter:   ptr.To(testDate(1, 6, 0, 0)),
						TaskName:        "unittest",
					},
					statusCode: http.StatusOK,
					responseBody: openapi.ScheduleRunV1Response{
						RunID: 1,
					},
				},
				// Check that the task got scheduled
				{
					method:     "GET",
					path:       "/api/v1/worker/runs",
					statusCode: http.StatusOK,
					responseBody: openapi.ListRunsV1Response{
						Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 1, TotalPages: 1},
						Result: []openapi.RunV1{
							{
								Id:           1,
								Reason:       openapi.Manual,
								Repositories: ptr.To([]string{"git.local/unit/test"}),
								RunData: ptr.To(map[string]string{
									"greeting":     "Hello",
									"sb.assignees": "ellie",
									"sb.reviewers": "abby",
								}),
								ScheduleAfter: testDate(1, 6, 0, 0),
								Status:        openapi.Pending,
								Task:          "unittest",
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

func TestServer_API_ReportWorkV1(t *testing.T) {
	testCases := []testCase{
		{
			name: `When a task defines a cron trigger then it schedules the next run according to the cron schedule`,
			tasks: []schema.Task{
				{
					Name: "unittest",
					Trigger: &schema.TaskTrigger{
						Cron: ptr.To("0 0 * * *"),
					},
				},
			},
			apiCalls: []apiCall{
				// Read the run that gets scheduled at the start of the server.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "84e3a25fc87b09b007d80d8aeacc2a5c4f3ba9a3f35590174148954486c56f16",
							Name: "unittest",
						},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID: 1,
						Task:  openapi.WorkTaskV1{Name: "unittest", Hash: "abc"},
						TaskResults: []openapi.ReportWorkV1TaskResult{
							{RepositoryName: "git.local/unit/test", Result: int(processor.ResultNoChanges)},
						},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// List the runs of the task.
				{
					method:     "GET",
					path:       "/api/v1/worker/runs",
					statusCode: http.StatusOK,
					responseBody: openapi.ListRunsV1Response{
						Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 2, TotalPages: 1},
						Result: []openapi.RunV1{
							{
								Id:            2,
								Reason:        openapi.Cron,
								ScheduleAfter: testDate(2, 0, 0, 0),
								Status:        openapi.Pending,
								Task:          defaultTask.Name,
							},
							{
								FinishedAt:    ptr.To(testDate(1, 0, 0, 3)),
								Id:            1,
								Reason:        openapi.Cron,
								ScheduleAfter: testDate(1, 0, 0, 0),
								StartedAt:     ptr.To(testDate(1, 0, 0, 2)),
								Status:        openapi.Finished,
								Task:          defaultTask.Name,
							},
						},
					},
				},
			},
		},

		{
			name: `When a result reports an open pr then it schedules a next run in one day`,
			tasks: []schema.Task{
				{Name: "unittest"},
			},
			apiCalls: []apiCall{
				// Schedule a new run for the first task.
				{
					method: "POST",
					path:   "/api/v1/runs",
					requestBody: openapi.ScheduleRunV1Request{
						TaskName: "unittest",
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
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "7d4262799e93d4fb6abc2f299a1846921256fc7aa64d80f87d2ad579e5c31306",
							Name: "unittest",
						},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "7d4262799e93d4fb6abc2f299a1846921256fc7aa64d80f87d2ad579e5c31306",
							Name: "unittest",
						},
						TaskResults: []openapi.ReportWorkV1TaskResult{
							{RepositoryName: "git.local/unit/test", Result: int(processor.ResultPrCreated)},
						},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// List the runs of the task.
				{
					method:     "GET",
					path:       "/api/v1/worker/runs",
					statusCode: http.StatusOK,
					responseBody: openapi.ListRunsV1Response{
						Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 2, TotalPages: 1},
						Result: []openapi.RunV1{
							{
								Id:            2,
								Reason:        openapi.Manual,
								ScheduleAfter: testDate(2, 0, 0, 1),
								Status:        openapi.Pending,
								Task:          defaultTask.Name,
							},
							{
								FinishedAt:    ptr.To(testDate(1, 0, 0, 4)),
								Id:            1,
								Reason:        openapi.Manual,
								ScheduleAfter: testDate(1, 0, 0, 1),
								StartedAt:     ptr.To(testDate(1, 0, 0, 3)),
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

func TestServer_API_ReportWorkV1_PrStatusChange(t *testing.T) {
	task := schema.Task{Name: "unittest"}
	taskHash := "7d4262799e93d4fb6abc2f299a1846921256fc7aa64d80f87d2ad579e5c31306"

	tc := testCase{
		name:  `When the outcome of a result changes then it creates a new task result entry`,
		tasks: []schema.Task{task},
		apiCalls: []apiCall{
			// Schedule the first run.
			{
				method: "POST",
				path:   "/api/v1/runs",
				requestBody: openapi.ScheduleRunV1Request{
					TaskName: task.Name,
				},
				statusCode: http.StatusOK,
				responseBody: openapi.ScheduleRunV1Response{
					RunID: 1,
				},
			},
			// Process the first run.
			{
				method:     "GET",
				path:       "/api/v1/worker/work",
				statusCode: http.StatusOK,
				responseBody: openapi.GetWorkV1Response{
					RunID: 1,
					Task: openapi.WorkTaskV1{
						Hash: taskHash,
						Name: task.Name,
					},
				},
			},
			// And report the result of the first run.
			{
				method: "POST",
				path:   "/api/v1/worker/work",
				requestBody: openapi.ReportWorkV1Request{
					RunID: 1,
					Task: openapi.WorkTaskV1{
						Hash: taskHash,
						Name: task.Name,
					},
					TaskResults: []openapi.ReportWorkV1TaskResult{
						{
							RepositoryName:   "git.local/unit/test",
							Result:           int(processor.ResultPrCreated),
							PullRequestUrl:   ptr.To("http://git.local/unit/test/pr/1"),
							PullRequestState: ptr.To(openapi.TaskResultStatusV1Open),
						},
					},
				},
				statusCode: http.StatusCreated,
				responseBody: openapi.ReportWorkV1Response{
					Result: "ok",
				},
			},
			// Schedule the second run.
			{
				method: "POST",
				path:   "/api/v1/runs",
				requestBody: openapi.ScheduleRunV1Request{
					TaskName: task.Name,
				},
				statusCode: http.StatusOK,
				responseBody: openapi.ScheduleRunV1Response{
					RunID: 2,
				},
			},
			// Process the second run.
			{
				method:     "GET",
				path:       "/api/v1/worker/work",
				statusCode: http.StatusOK,
				responseBody: openapi.GetWorkV1Response{
					RunID: 2,
					Task: openapi.WorkTaskV1{
						Hash: taskHash,
						Name: task.Name,
					},
				},
			},
			// And report the result of the second run.
			{
				method: "POST",
				path:   "/api/v1/worker/work",
				requestBody: openapi.ReportWorkV1Request{
					RunID: 2,
					Task: openapi.WorkTaskV1{
						Hash: taskHash,
						Name: task.Name,
					},
					TaskResults: []openapi.ReportWorkV1TaskResult{
						{
							RepositoryName:   "git.local/unit/test",
							Result:           int(processor.ResultPrMerged),
							PullRequestUrl:   ptr.To("http://git.local/unit/test/pr/1"),
							PullRequestState: ptr.To(openapi.TaskResultStatusV1Merged),
						},
					},
				},
				statusCode: http.StatusCreated,
				responseBody: openapi.ReportWorkV1Response{
					Result: "ok",
				},
			},
			{
				method:     "GET",
				path:       fmt.Sprintf("/api/v1/tasks/%s/results", task.Name),
				statusCode: http.StatusOK,
				responseBody: openapi.ListTaskRecentTaskResultsV1Response{
					Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 1, TotalPages: 1},
					TaskResults: []openapi.TaskResultV1{
						{
							PullRequestUrl: ptr.To("http://git.local/unit/test/pr/1"),
							RepositoryName: "git.local/unit/test",
							RunId:          2,
							Status:         openapi.TaskResultStatusV1Merged,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, tc)
}

func TestServer_API_ReportWorkV1_NoPrStatusChange(t *testing.T) {
	task := schema.Task{Name: "unittest"}
	taskHash := "7d4262799e93d4fb6abc2f299a1846921256fc7aa64d80f87d2ad579e5c31306"

	tc := testCase{
		name:  `When the outcome of a result does not change then it does not create a new task result entry`,
		tasks: []schema.Task{task},
		apiCalls: []apiCall{
			// Schedule the first run.
			{
				method: "POST",
				path:   "/api/v1/runs",
				requestBody: openapi.ScheduleRunV1Request{
					TaskName: task.Name,
				},
				statusCode: http.StatusOK,
				responseBody: openapi.ScheduleRunV1Response{
					RunID: 1,
				},
			},
			// Process the first run.
			{
				method:     "GET",
				path:       "/api/v1/worker/work",
				statusCode: http.StatusOK,
				responseBody: openapi.GetWorkV1Response{
					RunID: 1,
					Task: openapi.WorkTaskV1{
						Hash: taskHash,
						Name: task.Name,
					},
				},
			},
			// And report the result of the first run.
			{
				method: "POST",
				path:   "/api/v1/worker/work",
				requestBody: openapi.ReportWorkV1Request{
					RunID: 1,
					Task: openapi.WorkTaskV1{
						Hash: taskHash,
						Name: task.Name,
					},
					TaskResults: []openapi.ReportWorkV1TaskResult{
						{
							RepositoryName:   "git.local/unit/test",
							Result:           int(processor.ResultPrCreated),
							PullRequestUrl:   ptr.To("http://git.local/unit/test/pr/1"),
							PullRequestState: ptr.To(openapi.TaskResultStatusV1Open),
						},
					},
				},
				statusCode: http.StatusCreated,
				responseBody: openapi.ReportWorkV1Response{
					Result: "ok",
				},
			},
			// Schedule the second run.
			{
				method: "POST",
				path:   "/api/v1/runs",
				requestBody: openapi.ScheduleRunV1Request{
					TaskName: task.Name,
				},
				statusCode: http.StatusOK,
				responseBody: openapi.ScheduleRunV1Response{
					RunID: 2,
				},
			},
			// Process the second run.
			{
				method:     "GET",
				path:       "/api/v1/worker/work",
				statusCode: http.StatusOK,
				responseBody: openapi.GetWorkV1Response{
					RunID: 2,
					Task: openapi.WorkTaskV1{
						Hash: taskHash,
						Name: task.Name,
					},
				},
			},
			// And report the result of the second run.
			{
				method: "POST",
				path:   "/api/v1/worker/work",
				requestBody: openapi.ReportWorkV1Request{
					RunID: 2,
					Task: openapi.WorkTaskV1{
						Hash: taskHash,
						Name: task.Name,
					},
					TaskResults: []openapi.ReportWorkV1TaskResult{
						{
							RepositoryName:   "git.local/unit/test",
							Result:           int(processor.ResultPrOpen),
							PullRequestUrl:   ptr.To("http://git.local/unit/test/pr/1"),
							PullRequestState: ptr.To(openapi.TaskResultStatusV1Open),
						},
					},
				},
				statusCode: http.StatusCreated,
				responseBody: openapi.ReportWorkV1Response{
					Result: "ok",
				},
			},
			{
				method:     "GET",
				path:       "/api/v1/tasks/unittest/results",
				statusCode: http.StatusOK,
				responseBody: openapi.ListTaskRecentTaskResultsV1Response{
					Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 1, TotalPages: 1},
					TaskResults: []openapi.TaskResultV1{
						{
							PullRequestUrl: ptr.To("http://git.local/unit/test/pr/1"),
							RepositoryName: "git.local/unit/test",
							RunId:          1,
							Status:         openapi.TaskResultStatusV1Open,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, tc)
}

func TestServer_API_ReportWorkV1_AutoMergeOpenPrSchedule1Hour(t *testing.T) {
	tc := testCase{
		name: `When a result reports an open pr then it schedules a next run in one hour`,
		tasks: []schema.Task{
			{Name: "unittest", AutoMerge: true},
		},
		apiCalls: []apiCall{
			// Schedule a new run for the first task.
			{
				method: "POST",
				path:   "/api/v1/runs",
				requestBody: openapi.ScheduleRunV1Request{
					TaskName: "unittest",
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
					RunID: 1,
					Task: openapi.WorkTaskV1{
						Hash: "7ec41bb59c284620c38d2d01a8a26c96947b3ed96f28acf6051df054d56ae844",
						Name: "unittest",
					},
				},
			},
			// And report the result of the run.
			{
				method: "POST",
				path:   "/api/v1/worker/work",
				requestBody: openapi.ReportWorkV1Request{
					RunID: 1,
					Task: openapi.WorkTaskV1{
						Hash: "7ec41bb59c284620c38d2d01a8a26c96947b3ed96f28acf6051df054d56ae844",
						Name: "unittest",
					},
					TaskResults: []openapi.ReportWorkV1TaskResult{
						{RepositoryName: "git.local/unit/test", Result: int(processor.ResultPrCreated)},
					},
				},
				statusCode: http.StatusCreated,
				responseBody: openapi.ReportWorkV1Response{
					Result: "ok",
				},
			},
			// List the runs of the task.
			{
				method:     "GET",
				path:       "/api/v1/worker/runs",
				statusCode: http.StatusOK,
				responseBody: openapi.ListRunsV1Response{
					Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 2, TotalPages: 1},
					Result: []openapi.RunV1{
						{
							Id:            2,
							Reason:        openapi.Manual,
							ScheduleAfter: testDate(1, 1, 0, 1),
							Status:        openapi.Pending,
							Task:          defaultTask.Name,
						},
						{
							FinishedAt:    ptr.To(testDate(1, 0, 0, 4)),
							Id:            1,
							Reason:        openapi.Manual,
							ScheduleAfter: testDate(1, 0, 0, 1),
							StartedAt:     ptr.To(testDate(1, 0, 0, 3)),
							Status:        openapi.Finished,
							Task:          defaultTask.Name,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, tc)
}
