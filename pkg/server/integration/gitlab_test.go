package integration_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v59/github"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func TestServer_WebhookGitlab(t *testing.T) {
	testCases := []testCase{
		{
			name: `When a task triggers on a GitLab webhook then it schedules a new run`,
			tasks: []schema.Task{
				{
					Name: "unittest",
					Trigger: &schema.TaskTrigger{
						Webhook: &schema.TaskTriggerWebhook{
							Gitlab: []schema.GitlabTrigger{
								{Event: ptr.To("Push Hook"), Filters: []string{`.repository.path_with_namespace == "unit/test"`}},
							},
						},
					},
				},
			},
			apiCalls: []apiCall{
				// Drain the run queue.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "35e6e13d4ec91723e96cf0e93998043d042d98a637fe56f09bee1cd8558ea950",
							Name: "unittest",
						},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID:       1,
						Task:        openapi.WorkTaskV1{Name: "unittest"},
						TaskResults: []openapi.ReportWorkV1TaskResult{},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// Send the webhook request
				{
					method: "POST",
					path:   "/webhooks/gitlab",
					requestHeaders: map[string]string{
						"X-Gitlab-Event": "Push Hook",
						"X-Gitlab-Token": "secret",
					},
					requestBody: gitlab.PushEvent{
						Repository: &gitlab.Repository{PathWithNamespace: "unit/test"},
					},
					statusCode: http.StatusOK,
				},
				// Check for the newly scheduled run.
				{
					sleep:      5 * time.Millisecond, // Need to sleep because write of webhook happens in goroutine
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 2,
						Task: openapi.WorkTaskV1{
							Hash: "35e6e13d4ec91723e96cf0e93998043d042d98a637fe56f09bee1cd8558ea950",
							Name: "unittest",
						},
					},
				},
			},
		},

		{
			name: `When a filter does not match the content of the webhook then it does not schedule a new run`,
			tasks: []schema.Task{
				{
					Name: "unittest",
					Trigger: &schema.TaskTrigger{
						Webhook: &schema.TaskTriggerWebhook{
							Gitlab: []schema.GitlabTrigger{
								{Event: ptr.To("Push Hook"), Filters: []string{`.repository.path_with_namespace == "unit/test"`}},
							},
						},
					},
				},
			},
			apiCalls: []apiCall{
				// Drain the run queue.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "35e6e13d4ec91723e96cf0e93998043d042d98a637fe56f09bee1cd8558ea950",
							Name: "unittest",
						},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID:       1,
						Task:        openapi.WorkTaskV1{Name: "unittest"},
						TaskResults: []openapi.ReportWorkV1TaskResult{},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// Send the webhook request
				{
					method: "POST",
					path:   "/webhooks/gitlab",
					requestHeaders: map[string]string{
						"X-Gitlab-Event": "Push Hook",
						"X-Gitlab-Token": "secret",
					},
					requestBody: gitlab.PushEvent{
						Repository: &gitlab.Repository{PathWithNamespace: "unit/other"},
					},
					statusCode: http.StatusOK,
				},
				// Check that no new run has been scheduled.
				{
					sleep:        5 * time.Millisecond, // Need to sleep because write of webhook happens in goroutine
					method:       "GET",
					path:         "/api/v1/worker/work",
					statusCode:   http.StatusOK,
					responseBody: openapi.GetWorkV1Response{},
				},
			},
		},

		{
			name: `When task does not trigger on a GitLab webhook then it does not schedule a run`,
			tasks: []schema.Task{
				{
					Name: "unittest",
				},
			},
			apiCalls: []apiCall{
				// Drain the run queue.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "e42a6e186f31b860f22f07ed468b99c6dc75318542fc9ac8383358fae1b5ab8b",
							Name: "unittest",
						},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID:       1,
						Task:        openapi.WorkTaskV1{Name: "unittest"},
						TaskResults: []openapi.ReportWorkV1TaskResult{},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// Send the webhook request
				{
					method: "POST",
					path:   "/webhooks/gitlab",
					requestHeaders: map[string]string{
						"X-Gitlab-Event": "Push Hook",
						"X-Gitlab-Token": "secret",
					},
					requestBody: github.PushEvent{},
					statusCode:  http.StatusOK,
				},
				// Check that no new run has been scheduled.
				{
					sleep:        5 * time.Millisecond, // Need to sleep because write of webhook happens in goroutine
					method:       "GET",
					path:         "/api/v1/worker/work",
					statusCode:   http.StatusOK,
					responseBody: openapi.GetWorkV1Response{},
				},
			},
		},

		{
			name: `When the task defines a GitLab webhook with a delay then it schedules the run with a delay`,
			tasks: []schema.Task{
				{
					Name: "unittest",
					Trigger: &schema.TaskTrigger{
						Webhook: &schema.TaskTriggerWebhook{
							Delay: 300,
							Gitlab: []schema.GitlabTrigger{
								{Event: ptr.To("Push Hook"), Filters: []string{`.repository.path_with_namespace == "unit/test"`}},
							},
						},
					},
				},
			},
			apiCalls: []apiCall{
				// Drain the run queue.
				{
					method:     "GET",
					path:       "/api/v1/worker/work",
					statusCode: http.StatusOK,
					responseBody: openapi.GetWorkV1Response{
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "ef99cc7f5c98b01042d78394fa938bd6746c82f10033868e7302daf586ba33a2",
							Name: "unittest",
						},
					},
				},
				// And report the result of the run.
				{
					method: "POST",
					path:   "/api/v1/worker/work",
					requestBody: openapi.ReportWorkV1Request{
						RunID:       1,
						Task:        openapi.WorkTaskV1{Name: "unittest"},
						TaskResults: []openapi.ReportWorkV1TaskResult{},
					},
					statusCode: http.StatusCreated,
					responseBody: openapi.ReportWorkV1Response{
						Result: "ok",
					},
				},
				// Send the webhook request
				{
					method: "POST",
					path:   "/webhooks/gitlab",
					requestHeaders: map[string]string{
						"X-Gitlab-Event": "Push Hook",
						"X-Gitlab-Token": "secret",
					},
					requestBody: gitlab.PushEvent{
						Repository: &gitlab.Repository{PathWithNamespace: "unit/test"},
					},
					statusCode: http.StatusOK,
				},
				// Check that the new run is delayed by ~5 minutes.
				{
					sleep:      5 * time.Millisecond, // Need to sleep because write of webhook happens in goroutine
					method:     "GET",
					path:       "/api/v1/worker/runs",
					statusCode: http.StatusOK,
					responseBody: openapi.ListRunsV1Response{
						Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 2, TotalPages: 1},
						Result: []openapi.RunV1{
							{
								Id:            2,
								Reason:        openapi.Webhook,
								ScheduleAfter: testDate(1, 0, 5, 5),
								Status:        openapi.Pending,
								Task:          "unittest",
							},
							{
								FinishedAt:    ptr.To(testDate(1, 0, 0, 3)),
								Id:            1,
								Reason:        openapi.New,
								ScheduleAfter: testDate(1, 0, 0, 0),
								StartedAt:     ptr.To(testDate(1, 0, 0, 2)),
								Status:        openapi.Finished,
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
