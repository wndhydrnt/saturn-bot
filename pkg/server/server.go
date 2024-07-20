package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api"
	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api/openapi"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Start(taskPaths []string) error {
	s.httpServer = &http.Server{
		Addr: ":3000",
	}

	taskService, err := api.NewTaskService(taskPaths)
	if err != nil {
		return fmt.Errorf("load tasks on server start: %w", err)
	}

	taskCtrl := openapi.NewTaskAPIController(taskService)
	workerCtrl := openapi.NewWorkerAPIController(&api.WorkerService{})
	router := newRouter(taskCtrl, workerCtrl)
	s.httpServer.Handler = router
	go func(server *http.Server) {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed - exiting", "err", err)
			os.Exit(2)
		}
	}(s.httpServer)
	return nil
}

func (s *Server) Stop() error {
	if s.httpServer != nil {
		slog.Debug("Shutting down HTTP server")
		ctx := context.Background()
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Minute))
		err := s.httpServer.Shutdown(ctx)
		cancel()
		if err != nil {
			return fmt.Errorf("shutdown of http server failed: %w", err)
		}

		slog.Debug("Shutdown of HTTP server finished")
		return nil
	}

	return nil
}

func Run(taskPaths []string) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := &Server{}
	s.Start(taskPaths)
	slog.Info("Server started")
	sig := <-sigs
	slog.Info("Shutting down", "signal", sig.String())
	s.Stop()
	slog.Info("Server stopped")
	return nil
}

// newRouter copies openapi.newRouter.
// This is done to configure middlewares of chi.Router.
func newRouter(routers ...openapi.Router) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Compress(5))
	for _, api := range routers {
		for _, route := range api.Routes() {
			var handler http.Handler = route.HandlerFunc
			router.Method(route.Method, route.Pattern, handler)
		}
	}

	return router
}
