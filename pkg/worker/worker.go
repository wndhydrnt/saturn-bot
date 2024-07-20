package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/worker/client"
)

var (
	ctx       = context.Background()
	ErrNoExec = errors.New("no execution")
)

type Execution struct {
	ID string
}

type ExecutionResult struct {
	Payload string
}

type ExecutionSource interface {
	Next() (Execution, error)
	Report(ExecutionResult) error
}

type DummyExecutionSource struct{}

func (d *DummyExecutionSource) Next() (Execution, error) {
	if rand.Intn(2) == 1 {
		return Execution{}, ErrNoExec
	}

	return Execution{ID: genID(8)}, nil
}

func (d *DummyExecutionSource) Report(result ExecutionResult) error {
	slog.Info("Work finished", "executionID", result.Payload)
	return nil
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

	return Execution{ID: fmt.Sprintf("%d", resp.ExecutionID)}, nil
}

func (a *APIExecutionSource) Report(result ExecutionResult) error {
	req := a.client.ReportWorkV1(ctx)
	req = req.ReportWorkV1Request(client.ReportWorkV1Request{
		ExecutionID: 123,
		TaskResults: []client.ReportWorkV1RequestTaskResultsInner{
			{Name: "foobar", Result: int32(processor.ResultNoChanges)},
		},
	})
	_, _, err := a.client.ReportWorkV1Execute(req)
	if err != nil {
		return fmt.Errorf("send execution result to API: %w", err)
	}

	return nil
}

type Worker struct {
	Exec               ExecutionSource
	ParallelExecutions int

	resultChan chan ExecutionResult
	stopped    bool
	stopChan   chan chan struct{}
}

func (w *Worker) Start() {
	w.resultChan = make(chan ExecutionResult, 1)
	w.stopChan = make(chan chan struct{})
	t := time.NewTicker(5 * time.Second)
	executionCounter := 0
	for {
		select {
		case <-t.C:
			if w.stopped {
				continue
			}

			slog.Info("Parallel executions", "count", executionCounter)
			if executionCounter >= w.ParallelExecutions {
				slog.Debug("Max number of parallel executions reached")
				continue
			}

			exec, err := w.Exec.Next()
			if err != nil {
				if errors.Is(err, ErrNoExec) {
					slog.Info("No new executions")
				} else {
					slog.Error("Failed to get next execution", "err", err)
				}

				continue
			}

			// Process in a Go routine
			go func(exec Execution, result chan ExecutionResult) {
				wait := time.Duration(rand.Intn(30)) * time.Second
				slog.Info("Worker going to sleep", "duration", wait, "executionID", exec.ID)
				time.Sleep(wait)
				result <- ExecutionResult{Payload: exec.ID}
			}(exec, w.resultChan)
			executionCounter += 1

		case result := <-w.resultChan:
			err := w.Exec.Report(result)
			if err != nil {
				slog.Error("Failed to report execution", "executionID", result.Payload, "err", err)
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
					slog.Info("Waiting for workers to finish", "count", executionCounter)
					time.Sleep(1 * time.Second)
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

func Run() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	apiClient := client.NewAPIClient(&client.Configuration{
		Servers: []client.ServerConfiguration{
			{URL: "http://localhost:3000"},
		},
	})
	s := &Worker{
		Exec:               &APIExecutionSource{client: apiClient.WorkerAPI},
		ParallelExecutions: 4,
	}
	go s.Start()
	slog.Info("Worker started")
	sig := <-sigs
	slog.Info("Shutting down", "signal", sig.String())
	<-s.Stop()
	slog.Info("Worker stopped")
	return nil
}
