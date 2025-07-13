package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"go.uber.org/zap"
)

var (
	errUnknownApiKey = errors.New("unknown api key")
)

// APIServer provides the implementation of the OpenAPI endpoints.
type APIServer struct {
	Clock         clock.Clock
	TaskService   *service.TaskService
	WorkerService *service.WorkerService
}

// Stop gracefully stops the API server.
func (a *APIServer) Stop(ctx context.Context, checkInterval time.Duration) error {
	a.WorkerService.Shutdown()

	runCountInitial, err := a.WorkerService.CountRunningRuns()
	if err != nil {
		return err
	}

	if runCountInitial == 0 {
		return nil
	}

	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ticker.C:
			runCount, err := a.WorkerService.CountRunningRuns()
			if err != nil {
				ticker.Stop()
				return err
			}

			if runCount == 0 {
				ticker.Stop()
				return nil
			}

			log.Log().Warnf("Waiting for %d runs to finish before shutdown", runCount)

		case <-ctx.Done():
			markErr := a.WorkerService.MarkActiveRunsAsFailed("Run failed to report before shutdown")
			return fmt.Errorf("shutdown worker service: %w", errors.Join(ctx.Err(), markErr))
		}
	}
}

// NewAPIServerOptions are passed to [RegisterAPIServer].
type NewAPIServerOptions struct {
	ApiKey        string
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
	middlewares := []openapi.StrictMiddlewareFunc{newApiKeyMiddleware(options.ApiKey)}
	return openapi.HandlerWithOptions(
		openapi.NewStrictHandlerWithOptions(apiServer, middlewares, handlerOpts),
		openapi.ChiServerOptions{
			BaseRouter: options.Router,
		}), apiServer
}

func handleHttpError(w http.ResponseWriter, _ *http.Request, err error) {
	apiError := openapi.Error{}
	var statusCode int
	if errors.Is(err, errUnknownApiKey) {
		log.Log().Errorw("API key validation failed", zap.Error(err))
		statusCode = http.StatusUnauthorized
		apiError.Errors = append(apiError.Errors, openapi.ErrorDetail{
			Error:   sberror.ClientUnknownApiKey,
			Message: "unknown api key",
		})
	} else {
		log.Log().Errorw("Internal Server Error", zap.Error(err))
		statusCode = http.StatusInternalServerError
		apiError.Errors = append(apiError.Errors, openapi.ErrorDetail{
			Error:   sberror.ServerIDDefault,
			Message: "internal server error",
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	encErr := enc.Encode(apiError)
	if encErr != nil {
		log.Log().Errorw("Encode HTTP error", zap.Error(encErr))
	}
}

func newApiKeyMiddleware(key string) openapi.StrictMiddlewareFunc {
	return func(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
			if r.Header.Get(openapi.HeaderApiKey) != key {
				return nil, errUnknownApiKey
			}

			return f(ctx, w, r, request)
		}
	}
}
