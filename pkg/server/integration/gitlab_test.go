package integration_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v59/github"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"github.com/xanzy/go-gitlab"
)

func TestServer_WebhookGitlab(t *testing.T) {
	testCases := []testCase{
		{
			name: `Given a task that triggers on GitLab webhook event "push"
							When it receives a GitLab webhook
							Then it creates a new work item for the task`,
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
						Tasks: []openapi.GetWorkV1Task{
							{Hash: "35e6e13d4ec91723e96cf0e93998043d042d98a637fe56f09bee1cd8558ea950", Name: "unittest"},
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
						Tasks: []openapi.GetWorkV1Task{
							{Hash: "35e6e13d4ec91723e96cf0e93998043d042d98a637fe56f09bee1cd8558ea950", Name: "unittest"},
						},
					},
				},
			},
		},

		{
			name: `Given a task that triggers on GitLab webhook event "push" and specifies a filter
							When it receives a GitLab webhook that does not match the filter
							Then it does not create a new work item for the task`,
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
						Tasks: []openapi.GetWorkV1Task{
							{Hash: "35e6e13d4ec91723e96cf0e93998043d042d98a637fe56f09bee1cd8558ea950", Name: "unittest"},
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
			name: `Given a task that does not trigger on a GitLab webhook event
							When it receives a GitLab webhook
							Then it does not create a new work item for the task`,
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
						Tasks: []openapi.GetWorkV1Task{
							{Hash: "e42a6e186f31b860f22f07ed468b99c6dc75318542fc9ac8383358fae1b5ab8b", Name: "unittest"},
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
			name: `Given a task that triggers on a GitLab webhook event
							And that defines a delay for the trigger
							When it receives a GitLab webhook
							Then schedules the run with a delay`,
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
						Tasks: []openapi.GetWorkV1Task{
							{Hash: "ef99cc7f5c98b01042d78394fa938bd6746c82f10033868e7302daf586ba33a2", Name: "unittest"},
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
						Page: openapi.Page{Total: 2},
						Result: []openapi.RunV1{
							{
								Id:            2,
								Reason:        openapi.Webhook,
								ScheduleAfter: testDate(0, 5, 5),
								Status:        openapi.Pending,
								Task:          "unittest",
							},
							{
								FinishedAt:    ptr.To(testDate(0, 0, 3)),
								Id:            1,
								Reason:        openapi.New,
								ScheduleAfter: testDate(0, 0, 0),
								StartedAt:     ptr.To(testDate(0, 0, 2)),
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
