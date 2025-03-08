package integration_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func genGithubWebhookSignature(secret []byte, content []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(content)
	sum := mac.Sum(nil)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(sum))
}

func TestServer_WebhookGithub_Success(t *testing.T) {
	event := github.PushEvent{
		Ref: ptr.To("refs/tags/v3.14.1"),
		Repo: &github.PushEventRepository{
			ID: ptr.To(int64(100)),
		},
	}
	eventBytes, err := json.Marshal(event)
	require.NoError(t, err)

	tc := testCase{
		name: `When a task triggers on a GitHub webhook then it schedules a new run`,
		tasks: []schema.Task{
			{
				Name: "unittest",
				Trigger: &schema.TaskTrigger{
					Webhook: &schema.TaskTriggerWebhook{
						Github: []schema.GithubTrigger{
							{
								Event: ptr.To("push"),
								RunData: schema.GithubTriggerRunData{
									"tag":    `.ref | match("refs\/tags\/(.+)") | .captures[0].string`,
									"repoID": `.repository.id`,
								},
							},
						},
					},
				},
			},
		},
		apiCalls: []apiCall{
			// Send the webhook request
			{
				method: "POST",
				path:   "/webhooks/github",
				requestHeaders: map[string]string{
					github.EventTypeHeader:       "push",
					github.SHA256SignatureHeader: genGithubWebhookSignature([]byte("secret"), eventBytes),
				},
				requestBody: event,
				statusCode:  http.StatusOK,
			},
			// Check for the newly scheduled run.
			{
				sleep:      5 * time.Millisecond, // Need to sleep because write of webhook happens in goroutine
				method:     "GET",
				path:       "/api/v1/worker/work",
				statusCode: http.StatusOK,
				responseBody: openapi.GetWorkV1Response{
					RunID: 1,
					RunData: ptr.To(map[string]string{
						"tag":    "v3.14.1",
						"repoID": "100",
					}),
					Task: openapi.WorkTaskV1{
						Hash: "4ebc85491822e40239d27d97c01baf7c302441c369aa479fa401ab3ac0f9857b",
						Name: "unittest",
					},
				},
			},
		},
	}

	executeTestCase(t, tc)
}

func TestServer_WebhookGithub_NoTrigger(t *testing.T) {
	tc := testCase{
		name: `When task does not trigger on a GitHub webhook then it does not schedule a run`,
		tasks: []schema.Task{
			{
				Name: "unittest",
			},
		},
		apiCalls: []apiCall{
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
	}

	executeTestCase(t, tc)
}

func TestServer_WebhookGithub_WithDelay(t *testing.T) {
	tc := testCase{
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
	}

	executeTestCase(t, tc)
}
