package integration_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v68/github"
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
						RunID: 1,
						Task: openapi.WorkTaskV1{
							Hash: "41f8becac4691a4b990e6a92fd58810a17c20866abb07726eee0d588e703bfa5",
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
					path:       "/api/v1/runs",
					statusCode: http.StatusOK,
					responseBody: openapi.ListRunsV1Response{
						Page: openapi.Page{CurrentPage: 1, ItemsPerPage: 20, TotalItems: 1, TotalPages: 1},
						Result: []openapi.RunV1{
							{
								Id:            1,
								Reason:        openapi.Webhook,
								ScheduleAfter: testDate(1, 0, 5, 1),
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
