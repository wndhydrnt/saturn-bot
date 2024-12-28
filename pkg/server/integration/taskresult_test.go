package integration_test

import (
	"net/http"
	"testing"

	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func Test_API_ListTaskResultsV1(t *testing.T) {
	testCases := []testCase{
		{
			name:  `When task results exist then it lists them`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				// Read the run that gets scheduled at the start of the server.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Task:  openapi.WorkTaskV1{Hash: defaultTaskHash, Name: defaultTask.Name},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Name: defaultTask.Name,
						},
						TaskResults: []openapi.ReportWorkV1TaskResult{
							{
								PullRequestUrl: ptr.To("https://git.local/unittest/one/pr/1"),
								RepositoryName: "git.local/unittest/one",
								Result:         int(processor.ResultPrOpen),
							},
							{
								PullRequestUrl: ptr.To("https://git.local/unittest/two/pr/1"),
								RepositoryName: "git.local/unittest/two",
								Result:         int(processor.ResultPrClosed),
							},
							{
								PullRequestUrl: ptr.To("https://git.local/unittest/three/pr/1"),
								RepositoryName: "git.local/unittest/three",
								Result:         int(processor.ResultPrMerged),
							},
							{
								RepositoryName: "git.local/unittest/four",
								Result:         int(processor.ResultUnknown), Error: ptr.To("error"),
							},
						},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// And verify that the task results have been properly stored.
				{
					method:     "GET",
					path:       "/api/v1/taskResults",
					statusCode: http.StatusOK,
					responseBody: openapi.ListTaskResultsV1Response{
						Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 4, TotalPages: 1},
						TaskResults: []openapi.TaskResultV1{
							{
								RepositoryName: "git.local/unittest/four",
								RunId:          1,
								Error:          ptr.To("error"),
								Status:         openapi.TaskResultStatusV1Error,
							},
							{
								PullRequestUrl: ptr.To("https://git.local/unittest/three/pr/1"),
								RepositoryName: "git.local/unittest/three",
								RunId:          1,
								Status:         openapi.TaskResultStatusV1Merged,
							},
							{
								PullRequestUrl: ptr.To("https://git.local/unittest/two/pr/1"),
								RepositoryName: "git.local/unittest/two",
								RunId:          1,
								Status:         openapi.TaskResultStatusV1Closed,
							},
							{
								PullRequestUrl: ptr.To("https://git.local/unittest/one/pr/1"),
								RepositoryName: "git.local/unittest/one",
								RunId:          1,
								Status:         openapi.TaskResultStatusV1Open,
							},
						},
					},
				},
				// And verify that filtering by status works.
				{
					method:     "GET",
					path:       "/api/v1/taskResults",
					query:      "status=merged",
					statusCode: http.StatusOK,
					responseBody: openapi.ListTaskResultsV1Response{
						Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 1, TotalPages: 1},
						TaskResults: []openapi.TaskResultV1{
							{
								PullRequestUrl: ptr.To("https://git.local/unittest/three/pr/1"),
								RepositoryName: "git.local/unittest/three",
								RunId:          1,
								Status:         openapi.TaskResultStatusV1Merged,
							},
						},
					},
				},
			},
		},

		{
			name:  `When the run does not exist then it returns an empty list`,
			tasks: []schema.Task{defaultTask},
			apiCalls: []apiCall{
				{
					method: "GET",
					path:   "/api/v1/taskResults",
					requestBody: openapi.ListTaskResultsV1RequestObject{
						Params: openapi.ListTaskResultsV1Params{RunId: ptr.To(12)},
					},
					statusCode: http.StatusOK,
					responseBody: openapi.ListTaskResultsV1Response{
						Page:        openapi.Page{CurrentPage: 1, ItemsPerPage: 20},
						TaskResults: []openapi.TaskResultV1{},
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
