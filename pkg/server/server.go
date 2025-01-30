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
	chiprometheus "github.com/toshi0607/chi-prometheus"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/server/metrics"
	"github.com/wndhydrnt/saturn-bot/pkg/server/service"
	"github.com/wndhydrnt/saturn-bot/pkg/server/ui"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Start(opts options.Opts, taskPaths []string) error {
	metrics.Init(opts.PrometheusRegisterer)
	taskRegistry := task.NewRegistry(options.Opts{
		ActionFactories: opts.ActionFactories,
		FilterFactories: opts.FilterFactories,
		SkipPlugins:     true,
	})
	err := taskRegistry.ReadAll(taskPaths)
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

	taskService := service.NewTaskService(opts.Clock, database, taskRegistry)
	workerService := service.NewWorkerService(opts.Clock, database, taskService)
	syncService := service.NewSync(opts.Clock, database, taskService, workerService)
	if err := syncService.SyncTasksInDatabase(); err != nil {
		return err
	}

	router := newRouter(opts)
	webhookService, err := service.NewWebhookService(opts.Clock, taskRegistry, workerService)
	if err != nil {
		return fmt.Errorf("create webhook service: %w", err)
	}
	api.RegisterGithubWebhookHandler(router, []byte(opts.Config.ServerGithubWebhookSecret), webhookService)
	api.RegisterGitlabWebhookHandler(router, opts.Config.ServerGitlabWebhookSecret, webhookService)
	err = api.RegisterOpenAPIDefinitionRoute(opts.Config.ServerBaseUrl, router)
	if err != nil {
		return fmt.Errorf("failed to register OpenAPI definition route: %w", err)
	}
	api.RegisterHealthRoute(router)
	metrics.RegisterPrometheusRoute(metrics.RegisterPrometheusRouteOpts{
		PrometheusGatherer:   opts.PrometheusGatherer,
		PrometheusRegisterer: opts.PrometheusRegisterer,
		Router:               router,
	})

	if opts.Config.GoProfiling {
		router.Mount("/debug", middleware.Profiler())
	}

	handler, apiServer := api.RegisterAPIServer(&api.NewAPIServerOptions{
		Clock:         opts.Clock,
		Router:        router,
		TaskService:   taskService,
		WorkerService: workerService,
	})
	if opts.Config.ServerServeUi {
		log.Log().Info("Registering UI routes")
		ui.RegisterUiRoutes(router, apiServer)
	}

	s.httpServer = &http.Server{
		ReadHeaderTimeout: 10 * time.Millisecond,
		Addr:              opts.Config.ServerAddr,
	}
	s.httpServer.Handler = handler
	go func(server *http.Server) {
		log.Log().Infof("HTTP server listening on %s", opts.Config.ServerAddr)
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
func newRouter(opts options.Opts) chi.Router {
	router := chi.NewRouter()
	// if opts.Config.ServerCompress {
	// 	router.Use(middleware.Compress(5))
	// }

	if opts.Config.ServerAccessLog {
		router.Use(middleware.Logger)
	}

	pm := chiprometheus.New("saturn-bot")
	opts.PrometheusRegisterer.MustRegister(pm.Collectors()...)
	router.Use(pm.Handler)
	return router
}
