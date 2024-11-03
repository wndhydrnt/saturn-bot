package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/server/task"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"github.com/wndhydrnt/saturn-bot/pkg/version"
	"github.com/wndhydrnt/saturn-bot/pkg/worker/client"
	"go.uber.org/zap"
)

var (
	ctx       = context.Background()
	ErrNoExec = errors.New("no execution")
)

func handleApiError(err error) error {
	var apiError *client.GenericOpenAPIError
	if errors.As(err, &apiError) {
		return fmt.Errorf(
			"%s - %s",
			apiError.Error(),
			string(bytes.TrimSpace(apiError.Body())),
		)
	}

	return err
}

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
	client client.WorkerAPI
}

func (a *APIExecutionSource) Next() (Execution, error) {
	req := a.client.GetWorkV1(ctx)
	resp, _, err := a.client.GetWorkV1Execute(req)
	if err != nil {
		metricServerRequestsFailed.WithLabelValues(metricLabelOpGetWorkV1).Inc()
		return Execution{}, fmt.Errorf("api request to get work: %w", handleApiError(err))
	}

	if len(resp.Tasks) == 0 {
		return Execution{}, ErrNoExec
	}

	return Execution(*resp), nil
}

func (a *APIExecutionSource) Report(result Result) error {
	payload := client.ReportWorkV1Request{
		RunID:       result.Execution.RunID,
		TaskResults: mapRunResultsToTaskResults(result.TaskResults),
	}
	if result.RunError != nil {
		payload.Error = client.PtrString(result.RunError.Error())
	}

	log.Log().Debugf("Reporting run %d", result.Execution.RunID)
	req := a.client.
		ReportWorkV1(ctx).
		ReportWorkV1Request(payload)
	_, _, err := a.client.ReportWorkV1Execute(req)
	if err != nil {
		metricServerRequestsFailed.WithLabelValues(metricLabelOpReportWorkV1).Inc()
		return fmt.Errorf("send execution result to API: %w", handleApiError(err))
	}

	return nil
}

type Worker struct {
	Exec               ExecutionSource
	ParallelExecutions int

	opts       options.Opts
	resultChan chan Result
	tasks      []schema.ReadResult
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

	tasks, err := task.Load(taskPaths)
	if err != nil {
		return nil, err
	}

	apiClient := client.NewAPIClient(&client.Configuration{
		Servers: []client.ServerConfiguration{
			{URL: opts.Config.WorkerServerAPIBaseURL},
		},
	})
	return &Worker{
		Exec:  &APIExecutionSource{client: apiClient.WorkerAPI},
		opts:  opts,
		tasks: tasks,
	}, nil
}

func (w *Worker) Start() {
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
	var repositoryNames []string
	if exec.Repository != nil {
		repositoryNames = append(repositoryNames, *exec.Repository)
	}

	var taskPaths []string
	for _, taskReq := range exec.Tasks {
		t, err := w.findTaskByName(taskReq.Name, taskReq.Hash)
		if err != nil {
			result <- Result{
				RunError:  err,
				Execution: exec,
			}
			return
		}
		taskPaths = append(taskPaths, t.Path)
	}
	results, err := command.ExecuteRun(w.opts, repositoryNames, taskPaths)
	result <- Result{
		RunError:    err,
		Execution:   exec,
		TaskResults: results,
	}
}

func (w *Worker) findTaskByName(name string, hash string) (schema.ReadResult, error) {
	for _, t := range w.tasks {
		if t.Task.Name == name {
			if t.Sha256 == hash {
				return t, nil
			} else {
				return schema.ReadResult{}, fmt.Errorf("hash of task '%s' does not match - got '%s' want '%s'", name, hash, t.Hash)
			}
		}
	}

	return schema.ReadResult{}, fmt.Errorf("task '%s' not found", name)
}

func mapRunResultsToTaskResults(runResults []command.RunResult) []client.ReportWorkV1TaskResult {
	var results []client.ReportWorkV1TaskResult
	for _, rr := range runResults {
		if rr.Result == processor.ResultNoMatch {
			continue
		}

		result := client.ReportWorkV1TaskResult{
			RepositoryName: rr.RepositoryName,
			Result:         client.PtrInt32(int32(rr.Result)), // #nosec G115 -- no info by gosec on how to fix this
			TaskName:       rr.TaskName,
		}
		if rr.Error != nil {
			result.Error = client.PtrString(rr.Error.Error())
		}

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

	hs := &httpServer{}
	hs.handle("/healthz", http.HandlerFunc(healthHandler))
	hs.handle("/metrics", promhttp.Handler())
	go hs.start()

	log.Log().Infof("Worker started %s", version.String())
	sig := <-sigs
	log.Log().Infof("Caught signal %s - shutting down", sig.String())
	hs.stop()
	<-s.Stop()
	log.Log().Info("Worker stopped")
	return nil
}
