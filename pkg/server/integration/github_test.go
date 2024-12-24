package integration_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v59/github"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func TestServer_WebhookGithub(t *testing.T) {
	testCases := []testCase{
		{
			name: `When a task triggers on a GitHub webhook then it schedules a new run`,
			tasks: []schema.Task{
				{
					Name: "unittest",
					Trigger: &schema.TaskTrigger{
						Webhook: &schema.TaskTriggerWebhook{
							Github: []schema.GithubTrigger{
								{Event: ptr.To("push")},
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
							Hash: "9215d69a1ab043bb5a1e2150fc0dbb73888aea1340d8327d6f1467f6ac0211ee",
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
					path:   "/webhooks/github",
					requestHeaders: map[string]string{
						github.EventTypeHeader:       "push",
						github.SHA256SignatureHeader: genGithubWebhookSignature([]byte("secret"), []byte("{}")),
					},
					requestBody: github.PushEvent{},
					statusCode:  http.StatusOK,
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
							Hash: "9215d69a1ab043bb5a1e2150fc0dbb73888aea1340d8327d6f1467f6ac0211ee",
							Name: "unittest",
						},
					},
				},
			},
		},

		{
			name: `When task does not trigger on a GitHub webhook then it does not schedule a run`,
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
					path:   "/webhooks/github",
					requestHeaders: map[string]string{
						github.EventTypeHeader:       "push",
						github.SHA256SignatureHeader: genGithubWebhookSignature([]byte("secret"), []byte("{}")),
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
			name: `When the task defines a GitHub webhook with a delay then it schedules the run with a delay`,
			tasks: []schema.Task{
				{
					Name: "unittest",
					Trigger: &schema.TaskTrigger{
						Webhook: &schema.TaskTriggerWebhook{
							Delay: 300,
							Github: []schema.GithubTrigger{
								{Event: ptr.To("push")},
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
							Hash: "bfd1a13eb4a7c5220318e80ee34fa089163a8e338bb980c3cedc11f0c259df9d",
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
					path:   "/webhooks/github",
					requestHeaders: map[string]string{
						github.EventTypeHeader:       "push",
						github.SHA256SignatureHeader: genGithubWebhookSignature([]byte("secret"), []byte("{}")),
					},
					requestBody: github.PushEvent{},
					statusCode:  http.StatusOK,
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

func genGithubWebhookSignature(secret []byte, content []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(content)
	sum := mac.Sum(nil)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(sum))
}
