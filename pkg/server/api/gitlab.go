package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
)

const (
	gitlabWebhookEventIDHeader = "X-Gitlab-Event-UUID"
	gitlabWebhookTokenHeader   = "X-Gitlab-Token"
)

type GitlabWebhookHandler struct {
	SecretToken    string
	WebhookService *service.WebhookService
}

func (gh *GitlabWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	defer DiscardRequest(r)
	eventToken := r.Header.Get(gitlabWebhookTokenHeader)
	if eventToken != gh.SecretToken {
		log.Log().Debug("GitLab webhook received request with wrong token")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		log.Log().Debug("GitLab webhook called with wrong HTTP method")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventType := gitlab.HookEventType(r)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Log().Warn("Failed to read payload of GitLab webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var event map[string]any
	err = json.Unmarshal(payload, &event)
	if err != nil {
		log.Log().Warn("Failed to parse GitLab webhook")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = gh.WebhookService.EnqueueGitlab(&service.WebhookEnqueueInput{
		Event:   string(eventType),
		ID:      r.Header.Get(gitlabWebhookEventIDHeader),
		Payload: event,
	})
	if err != nil {
		log.Log().Errorw("Failed to enqueue webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterGitlabWebhookHandler registers the handler with a [github.com/go-chi/chi/v5.Router].
func RegisterGitlabWebhookHandler(router chi.Router, token string, ws *service.WebhookService) {
	h := &GitlabWebhookHandler{SecretToken: token, WebhookService: ws}
	router.Post("/webhooks/gitlab", h.HandleWebhook)
}
