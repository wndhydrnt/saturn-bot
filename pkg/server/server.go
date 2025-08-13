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
	sbdb "github.com/wndhydrnt/saturn-bot/pkg/db"
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
	apiServer             *api.APIServer
	httpServer            *http.Server
	shutdownCheckInterval time.Duration
	shutdownTimeout       time.Duration
}

func (s *Server) Start(opts options.Opts, taskPaths []string) error {
	s.shutdownCheckInterval = opts.ServerShutdownCheckInterval
	s.shutdownTimeout = opts.ServerShutdownTimeout
	if opts.Config.ServerApiKey == "" {
		return fmt.Errorf("required setting serverApiKey not configured - see https://saturn-bot.readthedocs.io/en/latest/reference/configuration/#serverapikey")
	}

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
		dbPath = filepath.Join(opts.DataDir, "db", "saturn-bot.db")
	}

	database, err := sbdb.New(opts.Config.ServerDatabaseLog, dbPath, sbdb.Migrate(db.Migrations()))
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}

	dbInfoService := service.NewDbInfo(database)
	taskService := service.NewTaskService(opts.Clock, database, taskRegistry)
	workerService := service.NewWorkerService(opts.Clock, database, taskService)
	syncService := service.NewSync(opts.Clock, database, taskService, workerService)
	if err := syncService.SyncTasksInDatabase(); err != nil {
		return err
	}

	metrics.Init(opts.PrometheusRegisterer, dbInfoService, taskService, workerService)

	router := newRouter(opts)
	webhookService, err := service.NewWebhookService(opts.Clock, taskRegistry, workerService)
	if err != nil {
		return fmt.Errorf("create webhook service: %w", err)
	}
	api.RegisterGithubWebhookHandler(router, []byte(opts.Config.ServerWebhookSecretGithub), webhookService)
	api.RegisterGitlabWebhookHandler(router, opts.Config.ServerWebhookSecretGitlab, webhookService)
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
		ApiKey:        opts.Config.ServerApiKey,
		Clock:         opts.Clock,
		Router:        router,
		TaskService:   taskService,
		WorkerService: workerService,
	})
	s.apiServer = apiServer
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

// Stop initiates a graceful shutdown of the server.
func (s *Server) Stop() error {
	apiErr := s.stopApiServer()
	httpErr := s.stopHttpServer()
	return errors.Join(apiErr, httpErr)
}

func (s *Server) stopApiServer() error {
	if s.apiServer == nil {
		return nil
	}

	log.Log().Debug("Shutting down API server")
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	checkInterval := s.shutdownCheckInterval
	if checkInterval == 0 {
		checkInterval = 1 * time.Second
	}
	err := s.apiServer.Stop(ctx, checkInterval)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Log().Errorf("API server shutdown timed out")
		} else {
			return err
		}
	} else {
		log.Log().Debug("Shutdown of API server finished")
	}

	return nil
}

func (s *Server) stopHttpServer() error {
	if s.httpServer == nil {
		return nil
	}

	log.Log().Debug("Shutting down HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("shutdown of http server failed: %w", err)
	}

	log.Log().Debug("Shutdown of HTTP server finished")
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
