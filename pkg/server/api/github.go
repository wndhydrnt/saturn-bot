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

type GithubWebhookHandler struct {
	SecretKey      []byte
	WebhookService *service.WebhookService
}

func (h *GithubWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, h.SecretKey)
	if err != nil {
		log.Log().Errorw("Failed to validate GitHub webhook", zap.Error(err))
		return
	}

	whType := github.WebHookType(r)
	whID := github.DeliveryID(r)
	log.Log().Debugf("Received GitHub webhook %s of type %s", whID, whType)
	var content any
	err = json.Unmarshal(payload, &content)
	if err != nil {
		log.Log().Errorw("Failed to unmarshal GitHub webhook", zap.Error(err))
		return
	}

	// Run in goroutine to not block the handler.
	go func() {
		log.Log().Debugf("Enqueuing GitHub webhook %s", whID)
		h.WebhookService.Enqueue(&service.EnqueueInput{
			Event:   whType,
			ID:      whID,
			Payload: content,
			Type:    service.GithubWebhookType,
		})
	}()
}

func RegisterGithubWebhookHandler(router chi.Router, secretKey []byte, ws *service.WebhookService) {
	h := &GithubWebhookHandler{SecretKey: secretKey, WebhookService: ws}
	router.Post("/webhooks/github", h.HandleWebhook)
}
