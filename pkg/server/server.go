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

	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr: ":3000",
	}

	workerCtrl := api.NewWorkerAPIController(&api.WorkerAPIService{})
	router := api.NewRouter(workerCtrl)
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

func Run() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := &Server{}
	s.Start()
	slog.Info("Server started")
	sig := <-sigs
	slog.Info("Shutting down", "signal", sig.String())
	s.Stop()
	slog.Info("Server stopped")
	return nil
}
