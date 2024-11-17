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
			name: `Given a task that triggers on GitHub webhook event "push" when it receives a GitHub webhook then it creates a new work item for the task`,
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
						Tasks: []openapi.GetWorkV1Task{
							{Hash: "8a6affb94ff09af5491b02dbcb5dff22ff56108e2f7d2032a8dd7661245015f4", Name: "unittest"},
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
						Tasks: []openapi.GetWorkV1Task{
							{Hash: "8a6affb94ff09af5491b02dbcb5dff22ff56108e2f7d2032a8dd7661245015f4", Name: "unittest"},
						},
					},
				},
			},
		},

		{
			name: `Given a task that does not trigger on a GitHub webhook event when it receives a GitHub webhook then it does not create a new work item for the task`,
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
