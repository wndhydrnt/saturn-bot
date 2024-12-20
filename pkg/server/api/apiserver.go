package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"go.uber.org/zap"
)

// APIServer provides the implementation of the OpenAPI endpoints.
type APIServer struct {
	Clock         clock.Clock
	TaskService   *service.TaskService
	WorkerService *service.WorkerService
}

// NewAPIServerOptions are passed to [RegisterAPIServer].
type NewAPIServerOptions struct {
	Clock         clock.Clock
	Router        chi.Router
	TaskService   *service.TaskService
	WorkerService *service.WorkerService
}

// RegisterAPIServer registers the OpenAPI implementation with the router.
func RegisterAPIServer(options *NewAPIServerOptions) (http.Handler, *APIServer) {
	var c clock.Clock
	if options.Clock == nil {
		c = clock.Default
	} else {
		c = options.Clock
	}

	apiServer := &APIServer{
		Clock:         c,
		TaskService:   options.TaskService,
		WorkerService: options.WorkerService,
	}

	handlerOpts := openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  handleHttpError,
		ResponseErrorHandlerFunc: handleHttpError,
	}
	return openapi.HandlerWithOptions(
		openapi.NewStrictHandlerWithOptions(apiServer, nil, handlerOpts),
		openapi.ChiServerOptions{
			BaseRouter: options.Router,
		}), apiServer
}

func handleHttpError(w http.ResponseWriter, _ *http.Request, err error) {
	log.Log().Errorw("Internal Server Error", zap.Error(err))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	apiError := openapi.Error{
		Error:   sberror.ServerIDDefault,
		Message: err.Error(),
	}
	enc := json.NewEncoder(w)
	encErr := enc.Encode(apiError)
	if encErr != nil {
		log.Log().Errorw("Encode HTTP error", zap.Error(encErr))
	}
}
