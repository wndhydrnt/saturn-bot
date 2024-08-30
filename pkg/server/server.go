package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"github.com/wndhydrnt/saturn-bot/pkg/server/task"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
	"go.uber.org/zap"
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

	database, err := db.New(opts.Config.ServerDatabaseLog, true, dbPath)
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
	err = api.RegisterOpenAPIDefinitionRoute(opts.Config.ServerBaseUrl, router)
	if err != nil {
		return fmt.Errorf("failed to register OpenAPI definition route: %w", err)
	}

	s.httpServer = &http.Server{
		ReadHeaderTimeout: 10 * time.Millisecond,
		Addr:              opts.Config.ServerAddr,
	}
	s.httpServer.Handler = router
	go func(server *http.Server) {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Log().Errorw("HTTP server failed - exiting", zap.Error(err))
			os.Exit(2)
		}
	}(s.httpServer)
	return nil
}

func (s *Server) Stop() error {
	if s.httpServer != nil {
		log.Log().Debug("Shutting down HTTP server")
		ctx := context.Background()
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(1*time.Minute))
		err := s.httpServer.Shutdown(ctx)
		cancel()
		if err != nil {
			return fmt.Errorf("shutdown of http server failed: %w", err)
		}

		log.Log().Debug("Shutdown of HTTP server finished")
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

	log.Log().Infof("Server started %s", version.String())
	sig := <-sigs
	log.Log().Infof("Caught signal %s - shutting down", sig.String())
	err = s.Stop()
	if err == nil {
		log.Log().Info("Server stopped")
	} else {
		log.Log().Errorw("Server failed during stop", zap.Error(err))
	}
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
