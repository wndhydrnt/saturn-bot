package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wndhydrnt/saturn-bot/pkg/client"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
	"go.uber.org/zap"
)

var (
	ctx       = context.Background()
	ErrNoExec = errors.New("no execution")
)

type Execution client.GetWorkV1Response

type Result struct {
	RunError    error
	Execution   Execution
	TaskResults []command.RunResult
}

type ExecutionSource interface {
	Next() (Execution, error)
	Report(Result) error
}

type APIExecutionSource struct {
	client client.ClientWithResponsesInterface
}

func (a *APIExecutionSource) Next() (Execution, error) {
	resp, err := a.client.GetWorkV1WithResponse(ctx)
	if err != nil {
		return Execution{}, fmt.Errorf("api request to get work: %w", err)
	}

	if resp.JSON200 != nil {
		if resp.JSON200.RunID == 0 {
			return Execution{}, ErrNoExec
		}

		return Execution(*resp.JSON200), nil
	}

	if resp.JSON401 != nil {
		var errs []error
		for _, apiErr := range resp.JSON401.Errors {
			errs = append(errs, fmt.Errorf("%d: %s", apiErr.Error, apiErr.Message))
		}

		return Execution{}, errors.Join(errs...)
	}

	return Execution{}, fmt.Errorf("server returned an unexpected response: %s", resp.Status())
}

func (a *APIExecutionSource) Report(result Result) error {
	payload := client.ReportWorkV1Request{
		RunID:       result.Execution.RunID,
		Task:        result.Execution.Task,
		TaskResults: mapRunResultsToTaskResults(result.TaskResults),
	}
	if result.RunError != nil {
		payload.Error = ptr.To(result.RunError.Error())
	}

	log.Log().Debugf("Reporting run %d", result.Execution.RunID)
	_, err := a.client.ReportWorkV1WithResponse(ctx, payload)
	if err != nil {
		metricServerRequestsFailed.WithLabelValues(metricLabelOpReportWorkV1).Inc()
		return fmt.Errorf("send execution result to API: %w", err)
	}

	return nil
}

type Worker struct {
	Exec               ExecutionSource
	ParallelExecutions int

	httpServer *http.Server
	opts       options.Opts
	resultChan chan Result
	tasks      []*task.Task
	stopped    bool
	stopChan   chan chan struct{}
}

func NewWorker(configPath string, taskPaths []string) (*Worker, error) {
	cfg, err := config.Read(configPath)
	if err != nil {
		return nil, err
	}

	opts, err := options.ToOptions(cfg)
	if err != nil {
		return nil, err
	}

	err = options.Initialize(&opts)
	if err != nil {
		return nil, err
	}

	reg := task.NewRegistry(options.Opts{
		ActionFactories: opts.ActionFactories,
		Config:          opts.Config,
		Clock:           opts.Clock,
		FilterFactories: opts.FilterFactories,
		Hosts:           opts.Hosts,
		IsCi:            opts.IsCi,
		// This registry holds tasks to perform validation.
		// No need to start plugins here.
		SkipPlugins: true,
	})
	err = reg.ReadAll(taskPaths)
	if err != nil {
		return nil, err
	}

	apiClient, err := client.NewCustomClientWithResponses(client.CustomClientWithResponsesOptions{
		ApiKey:  opts.Config.ServerApiKey,
		BaseUrl: opts.Config.WorkerServerAPIBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create openapi client: %w", err)
	}

	router := chi.NewRouter()
	router.Get("/healthz", http.HandlerFunc(healthHandler))
	router.Handle("GET /metrics", promhttp.Handler())
	if opts.Config.GoProfiling {
		router.Mount("/debug", middleware.Profiler())
	}

	worker := &Worker{
		Exec:       &APIExecutionSource{client: apiClient},
		httpServer: newHttpServer("", router),
		opts:       opts,
		tasks:      reg.GetTasks(),
	}

	router.Handle("GET /info", infoHandler(worker))
	return worker, nil
}

func (w *Worker) Start() {
	w.startHttpServer()
	w.resultChan = make(chan Result, 1)
	w.stopChan = make(chan chan struct{})
	t := time.NewTicker(w.opts.WorkerLoopInterval())
	parallelExecutions := w.opts.Config.WorkerParallelExecutions
	metricRunsMax.Set(float64(parallelExecutions))
	executionCounter := 0
	for {
		select {
		case <-t.C:
			if w.stopped {
				continue
			}

			if executionCounter > 0 {
				log.Log().Debugf("Parallel executions %d", executionCounter)
			}

			if executionCounter >= parallelExecutions {
				log.Log().Debug("Max number of parallel executions reached")
				continue
			}

			exec, err := w.Exec.Next()
			if err != nil {
				if !errors.Is(err, ErrNoExec) {
					metricServerRequestsFailed.WithLabelValues(metricLabelOpGetWorkV1).Inc()
					log.Log().Errorw("Failed to get next execution", zap.Error(err))
				}

				continue
			}

			log.Log().Debugf("Processing run %d", exec.RunID)
			// Process in a Go routine
			go w.executeRun(exec, w.resultChan)
			executionCounter += 1
			metricRuns.Inc()

		case result := <-w.resultChan:
			log.Log().Debugf("Received result of run %d", result.Execution.RunID)
			if result.RunError != nil {
				metricRunsFailed.Inc()
				log.Log().Errorw("Run failed", zap.Error(fmt.Errorf("ID %d: %w", result.Execution.RunID, result.RunError)))
			}

			err := w.Exec.Report(result)
			if err != nil {
				log.Log().Errorw("Failed to report run", zap.Error(fmt.Errorf("ID %d: %w", result.Execution.RunID, err)))
			}
			executionCounter -= 1
			metricRuns.Dec()

		case wait := <-w.stopChan:
			w.stopped = true
			t.Stop()
			go func() {
				for {
					if executionCounter == 0 {
						w.stopHttpServer()
						close(wait)
						return
					}
					waitDuration := 10 * time.Second
					log.Log().Infof("Waiting %s for %d workers to finish", waitDuration, executionCounter)
					time.Sleep(waitDuration)
				}
			}()
		}
	}
}

func (w *Worker) Stop() chan struct{} {
	waitChan := make(chan struct{})
	w.stopChan <- waitChan
	return waitChan
}

func (w *Worker) executeRun(exec Execution, result chan Result) {
	t, err := w.findTaskByName(exec.Task.Name, exec.Task.Hash)
	if err != nil {
		result <- Result{
			RunError:  err,
			Execution: exec,
		}
		return
	}

	taskPaths := []string{t.Path()}

	var repositories []string
	if exec.Repositories != nil {
		repositories = ptr.From(exec.Repositories)
	}

	var runData map[string]string
	if exec.RunData == nil {
		runData = map[string]string{}
	} else {
		runData = ptr.From(exec.RunData)
	}

	results, err := command.ExecuteRun(w.opts, repositories, taskPaths, runData)
	result <- Result{
		RunError:    err,
		Execution:   exec,
		TaskResults: results,
	}
}

func (w *Worker) findTaskByName(name string, hash string) (*task.Task, error) {
	for _, t := range w.tasks {
		if t.Task.Name == name {
			if t.Checksum() == hash {
				return t, nil
			} else {
				return nil, fmt.Errorf("hash of task '%s' does not match - got '%s' want '%s'", name, hash, t.Checksum())
			}
		}
	}

	return nil, fmt.Errorf("task '%s' not found", name)
}

func (w *Worker) startHttpServer() {
	if w.httpServer == nil {
		return
	}

	go func(s *http.Server) {
		log.Log().Infof("HTTP server listening on %s", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Log().Errorw("HTTP server failed", zap.Error(err))
			}
		}
	}(w.httpServer)
}

func (w *Worker) stopHttpServer() {
	if w.httpServer == nil {
		return
	}

	log.Log().Info("Stopping HTTP server")
	err := w.httpServer.Shutdown(context.Background())
	if err != nil {
		log.Log().Errorw("Shutdown of HTTP server failed", zap.Error(err))
	}
}

func mapRunResultsToTaskResults(runResults []command.RunResult) []client.ReportWorkV1TaskResult {
	var results []client.ReportWorkV1TaskResult
	for _, rr := range runResults {
		// Always report if a pull request is available.
		// Do this to update state in the database of the server.
		if rr.PullRequest == nil && !shouldReport(rr.Result) {
			continue
		}

		result := client.ReportWorkV1TaskResult{
			RepositoryName: rr.RepositoryName,
			Result:         int(rr.Result),
			State:          client.TaskResultStateV1Unknown,
		}
		updateTaskResultFromRunResult(&result, rr)
		results = append(results, result)
	}

	return results
}

func Run(configPath string, taskPaths []string) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s, err := NewWorker(configPath, taskPaths)
	if err != nil {
		return fmt.Errorf("start worker: %w", err)
	}

	initMetrics()
	go s.Start()

	log.Log().Infof("Worker started %s", version.String())
	sig := <-sigs
	log.Log().Infof("Caught signal %s - shutting down", sig.String())
	<-s.Stop()
	log.Log().Info("Worker stopped")
	return nil
}

func shouldReport(result processor.Result) bool {
	switch result {
	case processor.ResultNoChanges:
		return false
	case processor.ResultNoMatch:
		return false
	case processor.ResultSkip:
		return false
	default:
		return true
	}
}

func mapPullRequestStateToTaskResultStatus(state host.PullRequestState) client.TaskResultStateV1 {
	switch state {
	case host.PullRequestStateClosed:
		return client.TaskResultStateV1Closed
	case host.PullRequestStateMerged:
		return client.TaskResultStateV1Merged
	case host.PullRequestStateOpen:
		return client.TaskResultStateV1Open
	default:
		return client.TaskResultStateV1Unknown
	}
}

func updateTaskResultFromRunResult(taskResult *client.ReportWorkV1TaskResult, runResult command.RunResult) {
	if runResult.Error != nil {
		taskResult.Error = ptr.To(runResult.Error.Error())
		taskResult.State = client.TaskResultStateV1Error
		return
	}

	if runResult.Result == processor.ResultPushedDefaultBranch {
		taskResult.State = client.TaskResultStateV1Pushed
		return
	}

	if runResult.PullRequest != nil {
		taskResult.PullRequestUrl = ptr.To(runResult.PullRequest.WebURL)
		taskResult.State = mapPullRequestStateToTaskResultStatus(runResult.PullRequest.State)
		return
	}
}
