package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api"
	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"github.com/wndhydrnt/saturn-bot/pkg/server/task"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Start(opts options.Opts, taskPaths []string) error {
	tasks, err := task.Load(taskPaths)
	if err != nil {
		return fmt.Errorf("load tasks on server start: %w", err)
	}

	dbPath := opts.Config.ServerDatabasePath
	if dbPath == "" {
		dbPath = filepath.Join(opts.DataDir(), "db", "saturn-bot.db")
	}

	database, err := db.New(true, dbPath)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	taskService := service.NewTaskService(database, tasks)
	err = taskService.SyncDbTasks()
	if err != nil {
		return err
	}

	taskCtrl := openapi.NewTaskAPIController(&api.TaskHandler{TaskService: taskService})
	workerService := service.NewWorkerService(database, tasks)
	workerHandler := &api.WorkHandler{WorkerService: workerService}
	workerCtrl := openapi.NewWorkerAPIController(workerHandler)
	router := newRouter(opts, taskCtrl, workerCtrl)
	s.httpServer = &http.Server{
		Addr: opts.Config.ServerAddr,
	}
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

func Run(configPath string, taskPaths []string) error {
	cfg, err := config.Read(configPath)
	if err != nil {
		return err
	}

	opts, err := options.ToOptions(cfg)
	if err != nil {
		return err
	}

	err = options.Initialize(&opts)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := &Server{}
	err = s.Start(opts, taskPaths)
	if err != nil {
		return err
	}

	slog.Info("Server started")
	sig := <-sigs
	slog.Info("Shutting down", "signal", sig.String())
	s.Stop()
	slog.Info("Server stopped")
	return nil
}

// newRouter copies openapi.newRouter.
// This is done to configure middlewares of chi.Router.
func newRouter(opts options.Opts, routers ...openapi.Router) chi.Router {
	router := chi.NewRouter()
	if opts.Config.ServerCompress {
		router.Use(middleware.Compress(5))
	}

	if opts.Config.ServerAccessLog {
		router.Use(middleware.Logger)
	}

	for _, api := range routers {
		for _, route := range api.Routes() {
			var handler http.Handler = route.HandlerFunc
			router.Method(route.Method, route.Pattern, handler)
		}
	}

	return router
}
