package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/metrics"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"go.uber.org/zap"
)

var (
	ErrNoHostsConfigured = errors.New("no hosts configured")
)

type RunResult struct {
	Error          error
	PullRequest    *host.PullRequest
	RepositoryName string
	Result         processor.Result
	TaskName       string
}

type Run struct {
	Clock            clock.Clock
	DryRun           bool
	Hosts            []host.Host
	Processor        processor.RepositoryTaskProcessor
	PullRequestCache host.PullRequestCache
	PushGateway      *push.Pusher
	RepositoryLister host.RepositoryLister
	TaskRegistry     *task.Registry
}

func (r *Run) Run(repositoryNames, taskFiles []string, inputs map[string]string) ([]RunResult, error) {
	metrics.RunStart.SetToCurrentTime()
	defer func() {
		metrics.RunFinish.SetToCurrentTime()
		r.pushMetrics()
	}()

	if len(r.Hosts) == 0 {
		return nil, ErrNoHostsConfigured
	}

	err := r.TaskRegistry.ReadAll(taskFiles)
	if err != nil {
		return nil, err
	}

	defer r.TaskRegistry.Stop()
	tasks := r.TaskRegistry.GetTasks()
	if len(tasks) == 0 {
		log.Log().Warn("0 tasks loaded from files - stopping")
		return nil, nil
	}

	if err := host.UpdatePullRequestCache(r.Clock, r.Hosts, r.PullRequestCache); err != nil {
		return nil, fmt.Errorf("update pull request cache on start of the run: %w", err)
	}

	tasks = setInputs(tasks, inputs)
	repos := make(chan host.Repository)
	doneChan := make(chan error)
	if len(repositoryNames) > 0 {
		go discoverRepositoriesFromCLI(r.Hosts, repositoryNames, repos, doneChan)
	} else {
		go r.RepositoryLister.List(r.Hosts, repos, doneChan)
	}

	// Track the outcome of each task.
	taskSuccessTracker := initTaskSuccessTracker(tasks)
	defer recordTaskSuccessMetric(taskSuccessTracker)

	success := true
	var results []RunResult
	done := false
	for {
		select {
		case repo := <-repos:
			doFilter := len(repositoryNames) == 0
			processResults := r.Processor.Process(r.DryRun, repo, tasks, doFilter)
			for _, p := range processResults {
				if p.Error == nil && taskSuccessTracker[p.Task.Name] == nil {
					taskSuccessTracker[p.Task.Name] = ptr.To(float64(1))
				}

				if p.Error != nil {
					taskSuccessTracker[p.Task.Name] = ptr.To(float64(0))
					success = false
				}

				results = append(results, RunResult{
					Error:          p.Error,
					PullRequest:    p.PullRequest,
					RepositoryName: repo.FullName(),
					Result:         p.Result,
					TaskName:       p.Task.Name,
				})
			}
		case err := <-doneChan:
			if err != nil {
				return results, err
			}

			done = true
		}

		if done {
			break
		}
	}

	if !success {
		return results, fmt.Errorf("errors occurred, check previous log messages")
	}
	log.Log().Info("Run finished")
	return results, nil
}

func (r *Run) pushMetrics() {
	if r.PushGateway != nil {
		err := r.PushGateway.Push()
		if err != nil {
			log.Log().Warnw("Push to Prometheus PushGateway failed", zap.Error(err))
		}
	}
}

func ExecuteRun(opts options.Opts, repositoryNames, taskFiles []string, inputs map[string]string) ([]RunResult, error) {
	err := options.Initialize(&opts)
	if err != nil {
		return nil, fmt.Errorf("initialize options: %w", err)
	}

	taskRegistry := task.NewRegistry(opts)

	gitClient, err := git.New(opts)
	if err != nil {
		return nil, fmt.Errorf("new git client for run: %w", err)
	}

	dataCache, err := cache.New(filepath.Join(opts.DataDir, "cache.db"))
	if err != nil {
		return nil, err
	}

	repositoryCache := host.NewRepositoryCache(dataCache, clock.Default, filepath.Join(opts.DataDir, "cache"), opts.RepositoryCacheTtl)

	prCache := host.NewPullRequestCacheFromHosts(dataCache, opts.Hosts)
	e := &Run{
		Clock:  opts.Clock,
		DryRun: opts.Config.DryRun,
		Hosts:  opts.Hosts,
		Processor: &processor.Processor{
			DataDir:          opts.DataDir,
			Git:              gitClient,
			PullRequestCache: prCache,
		},
		PullRequestCache: prCache,
		PushGateway:      opts.PushGateway,
		RepositoryLister: repositoryCache,
		TaskRegistry:     taskRegistry,
	}
	return e.Run(repositoryNames, taskFiles, inputs)
}

func applyActionsInDirectory(actions []action.Action, ctx context.Context, dir string) error {
	return inDirectory(dir, func() error {
		for _, a := range actions {
			err := a.Apply(ctx)
			if err != nil {
				return fmt.Errorf("action %s failed: %w", a.String(), err)
			}
		}

		return nil
	})
}

func inDirectory(dir string, f func() error) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		return fmt.Errorf("change to work directory: %w", err)
	}

	funcErr := f()
	err = os.Chdir(currentDir)
	if err != nil {
		return fmt.Errorf("changing back to previous directory: %w", err)
	}

	return funcErr
}

// discoverRepositoriesFromCLI takes a list of repository names and turns them into repositories.
func discoverRepositoriesFromCLI(
	hosts []host.Host,
	repositoryNames []string,
	repoChan chan host.Repository,
	errChan chan error,
) {
	log.Log().Info("Discovering repositories from CLI")
	for _, repoName := range repositoryNames {
		repo, err := host.NewRepositoryFromName(hosts, repoName)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- repo
	}

	errChan <- nil
}

// setInputs sets inputs passed to Run().
// It filters out tasks when an expected input is missing.
func setInputs(tasks []*task.Task, inputs map[string]string) []*task.Task {
	var tasksWithInputs []*task.Task
	for _, t := range tasks {
		err := t.SetInputs(inputs)
		if err == nil {
			tasksWithInputs = append(tasksWithInputs, t)
		} else {
			log.Log().Warnf("Deactivating Task due to missing inputs: %s", err)
		}
	}

	return tasksWithInputs
}

// initTaskSuccessTracker initializes a map to track the outcome of each task.
// Initial value for each entry is always nil, which means that no value has been recorded.
// See also function recordTaskSuccessMetric.
func initTaskSuccessTracker(tasks []*task.Task) map[string]*float64 {
	data := make(map[string]*float64, len(tasks))
	for _, t := range tasks {
		data[t.Name] = nil
	}

	return data
}

// recordTaskSuccessMetric updates the metric [github.com/wndhydrnt/saturn-bot/pkg/metrics.RunTaskSuccess]
// with the outcome of each task.
func recordTaskSuccessMetric(tracker map[string]*float64) {
	for taskName, metricVal := range tracker {
		// Default to 0 if nil, which means failure.
		// Handles that case where a "global" error occurs before any task can be executed.
		// For example a failure to list the repositories from the host.
		val := ptr.FromDef(metricVal, 0.0)
		metrics.RunTaskSuccess.WithLabelValues(taskName).Set(val)
	}
}
