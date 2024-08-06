package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/server/task"
	"github.com/wndhydrnt/saturn-bot/pkg/worker/client"
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
	client client.WorkerAPI
}

func (a *APIExecutionSource) Next() (Execution, error) {
	req := a.client.GetWorkV1(ctx)
	resp, _, err := a.client.GetWorkV1Execute(req)
	if err != nil {
		return Execution{}, fmt.Errorf("api request to get work: %w", err)
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

	req := a.client.
		ReportWorkV1(ctx).
		ReportWorkV1Request(payload)
	_, _, err := a.client.ReportWorkV1Execute(req)
	if err != nil {
		return fmt.Errorf("send execution result to API: %w", err)
	}

	return nil
}

type Worker struct {
	Exec               ExecutionSource
	ParallelExecutions int

	opts       options.Opts
	resultChan chan Result
	tasks      []task.Task
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
	executionCounter := 0
	for {
		select {
		case <-t.C:
			if w.stopped {
				continue
			}

			slog.Debug("Parallel executions", "count", executionCounter)
			if executionCounter >= parallelExecutions {
				slog.Debug("Max number of parallel executions reached")
				continue
			}

			exec, err := w.Exec.Next()
			if err != nil {
				if errors.Is(err, ErrNoExec) {
					slog.Info("No new executions")
				} else {
					slog.Error("Failed to get next execution", "error", err)
				}

				continue
			}

			// Process in a Go routine
			go w.executeRun(exec, w.resultChan)
			executionCounter += 1

		case result := <-w.resultChan:
			if result.RunError != nil {
				slog.Error("Run failed", "runID", result.Execution.RunID, "error", result.RunError)
			}

			err := w.Exec.Report(result)
			if err != nil {
				slog.Error("Failed to report run", "runID", result.Execution.RunID, "error", err)
			}
			executionCounter -= 1

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
					slog.Info(fmt.Sprintf("Waiting %s for %d workers to finish", waitDuration, executionCounter))
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
		taskPaths = append(taskPaths, t.TaskPath)
	}
	results, err := command.ExecuteRun(map[string]string{}, w.opts, repositoryNames, taskPaths)
	result <- Result{
		RunError:    err,
		Execution:   exec,
		TaskResults: results,
	}
}

func (w *Worker) findTaskByName(name string, hash string) (task.Task, error) {
	for _, t := range w.tasks {
		if t.TaskName == name {
			if t.Hash == hash {
				return t, nil
			} else {
				return task.Task{}, fmt.Errorf("hash of task '%s' does not match - got '%s' want '%s'", name, hash, t.Hash)
			}
		}
	}

	return task.Task{}, fmt.Errorf("task '%s' not found", name)
}

func mapRunResultsToTaskResults(runResults []command.RunResult) []client.ReportWorkV1TaskResult {
	var results []client.ReportWorkV1TaskResult
	for _, rr := range runResults {
		if rr.Result == processor.ResultNoMatch {
			continue
		}

		result := client.ReportWorkV1TaskResult{
			RepositoryName: rr.RepositoryName,
			Result:         int32(rr.Result),
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

	go s.Start()
	slog.Info("Worker started")
	sig := <-sigs
	slog.Info("Shutting down", "signal", sig.String())
	<-s.Stop()
	slog.Info("Worker stopped")
	return nil
}
