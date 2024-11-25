package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
)

// APIServer provides the implementation of the OpenAPI endpoints.
type APIServer struct {
	Clock         clock.Clock
	TaskService   *service.TaskService
	WorkerService *service.WorkerService
}

// NewAPIServerOptions are passed to [RegisterAPIServer].
type NewAPIServerOptions struct {
	Router        chi.Router
	TaskService   *service.TaskService
	WorkerService *service.WorkerService
}

// RegisterAPIServer registers the OpenAPI implementation with the router.
func RegisterAPIServer(options *NewAPIServerOptions) http.Handler {
	apiServer := &APIServer{
		TaskService:   options.TaskService,
		WorkerService: options.WorkerService,
	}

	return openapi.HandlerWithOptions(
		openapi.NewStrictHandlerWithOptions(apiServer, nil, openapi.StrictHTTPServerOptions{}),
		openapi.ChiServerOptions{
			BaseRouter: options.Router,
		})
}
