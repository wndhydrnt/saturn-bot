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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
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
		metricServerRequestsFailed.WithLabelValues(metricLabelOpGetWorkV1).Inc()
		return Execution{}, fmt.Errorf("api request to get work: %w", err)
	}

	if len(resp.JSON200.Tasks) == 0 {
		return Execution{}, ErrNoExec
	}

	return Execution(*resp.JSON200), nil
}

func (a *APIExecutionSource) Report(result Result) error {
	payload := client.ReportWorkV1Request{
		RunID:       result.Execution.RunID,
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

	apiClient, err := client.NewClientWithResponses(opts.Config.WorkerServerAPIBaseURL)
	if err != nil {
		return nil, fmt.Errorf("create openapi client: %w", err)
	}

	return &Worker{
		Exec:  &APIExecutionSource{client: apiClient},
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
			Result:         int(rr.Result),
			TaskName:       rr.TaskName,
		}
		if rr.Error != nil {
			result.Error = ptr.To(rr.Error.Error())
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
