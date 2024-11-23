package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-github/v59/github"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"go.uber.org/zap"
)

// GithubWebhookHandler handles webhooks received by GitLab.
type GithubWebhookHandler struct {
	SecretKey      []byte
	WebhookService *service.WebhookService
}

// HandleWebhook parses and validates a webhook sent by GitHub.
// If the webhook is valid, it sends the webhook on for processing.
func (h *GithubWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, h.SecretKey)
	if err != nil {
		log.Log().Errorw("Failed to validate GitHub webhook", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	whType := github.WebHookType(r)
	whID := github.DeliveryID(r)
	log.Log().Debugf("Received GitHub webhook %s of type %s", whID, whType)
	var content any
	err = json.Unmarshal(payload, &content)
	if err != nil {
		log.Log().Errorw("Failed to unmarshal GitHub webhook", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Note: GitHub expects a response within 10 seconds.
	log.Log().Debugf("Enqueuing GitHub webhook %s", whID)
	err = h.WebhookService.EnqueueGithub(&service.WebhookEnqueueInput{
		Event:   whType,
		ID:      whID,
		Payload: content,
	})
	if err != nil {
		log.Log().Errorw("Failed to enqueue GitHub webhook", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterGithubWebhookHandler registers the handler with a [github.com/go-chi/chi/v5.Router].
func RegisterGithubWebhookHandler(router chi.Router, secretKey []byte, ws *service.WebhookService) {
	h := &GithubWebhookHandler{SecretKey: secretKey, WebhookService: ws}
	router.Post("/webhooks/github", h.HandleWebhook)
}
