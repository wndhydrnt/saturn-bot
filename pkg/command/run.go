package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
	sContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/metrics"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"go.uber.org/zap"
)

var (
	ErrNoHostsConfigured = errors.New("no hosts configured")
)

type RunResult struct {
	Error          error
	RepositoryName string
	Result         processor.Result
	TaskName       string
}

type Run struct {
	Cache        cache.Cache
	DryRun       bool
	Hosts        []host.Host
	Processor    processor.RepositoryTaskProcessor
	PushGateway  *push.Pusher
	TaskRegistry *task.Registry
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

	err := r.Cache.Read()
	if err != nil {
		return nil, err
	}

	var since *time.Time
	if r.Cache.GetLastExecutionAt() != 0 {
		ts := time.UnixMicro(r.Cache.GetLastExecutionAt())
		since = &ts
	}

	err = r.TaskRegistry.ReadAll(taskFiles)
	if err != nil {
		return nil, err
	}

	tasks := r.TaskRegistry.GetTasks()
	if len(tasks) == 0 {
		log.Log().Warn("0 tasks loaded from files - stopping")
		return nil, nil
	}

	tasks = setInputs(tasks, inputs)
	needsAllRepositories := hasUpdatedTasks(r.Cache.GetCachedTasks(), tasks)
	if needsAllRepositories {
		since = nil
	}

	repos := make(chan []host.Repository)
	errChan := make(chan error)
	var expectedFinishes int
	if len(repositoryNames) > 0 {
		expectedFinishes = discoverRepositoriesFromCLI(r.Hosts, repositoryNames, repos, errChan)
	} else {
		expectedFinishes = discoverRepositoriesFromHosts(r.Hosts, since, repos, errChan)
	}
	finishes := 0
	visitedRepositories := map[string]struct{}{}
	success := true
	var results []RunResult
	for {
		select {
		case repoList := <-repos:
			for _, repo := range repoList {
				log.Log().Debugf("Discovered repository %s", repo.FullName())
				_, exists := visitedRepositories[repo.FullName()]
				if exists {
					log.Log().Debugf("Repository %s already visited", repo.FullName())
					continue
				}

				visitedRepositories[repo.FullName()] = struct{}{}
				ctx := context.Background()
				ctx = context.WithValue(ctx, sContext.RunDataKey{}, make(map[string]string))
				doFilter := len(repositoryNames) == 0
				for _, t := range tasks {
					result := RunResult{
						RepositoryName: repo.FullName(),
						TaskName:       t.Name,
					}
					logger := log.Log().
						WithOptions(zap.Fields(
							log.FieldDryRun(r.DryRun),
							log.FieldRepo(repo.FullName()),
							log.FieldTask(t.Name),
						))
					ctx = sContext.WithLog(ctx, logger)
					result.Result, result.Error = r.Processor.Process(ctx, r.DryRun, repo, t, doFilter)
					if result.Error == nil {
						metrics.RunTaskSuccess.WithLabelValues(t.Name).Set(1)
					} else {
						metrics.RunTaskSuccess.WithLabelValues(t.Name).Set(0)
						success = false
						logger.Errorw("Task failed", "error", result.Error)
					}

					results = append(results, result)
				}
			}
		case err := <-errChan:
			if err != nil {
				return results, err
			}

			finishes += 1
		}
		if finishes == expectedFinishes {
			break
		}
	}

	if !r.DryRun {
		// Only update cache if this is not a dry run.
		// Without this guard, subsequent non dry runs would not recognize that they need to do anything.
		r.Cache.SetLastExecutionAt(time.Now().UnixMicro())
		r.Cache.UpdateCachedTasks(tasks)
		err = r.Cache.Write()
		if err != nil {
			return results, err
		}
	}

	r.TaskRegistry.Stop()

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

	cache := cache.NewJsonFile(path.Join(opts.DataDir(), cache.DefaultJsonFileName))
	taskRegistry := task.NewRegistry(opts)

	gitClient, err := git.New(opts)
	if err != nil {
		return nil, fmt.Errorf("new git client for run: %w", err)
	}

	e := &Run{
		Cache:  cache,
		DryRun: opts.Config.DryRun,
		Hosts:  opts.Hosts,
		Processor: &processor.Processor{
			DataDir: opts.DataDir(),
			Git:     gitClient,
		},
		PushGateway:  opts.PushGateway,
		TaskRegistry: taskRegistry,
	}
	return e.Run(repositoryNames, taskFiles, inputs)
}

func hasUpdatedTasks(cachedTasks []cache.CachedTask, tasks []*task.Task) bool {
	for _, t := range tasks {
		found := false
		for _, ct := range cachedTasks {
			if t.Name == ct.Name {
				found = true
				if t.Checksum() != ct.Checksum {
					return true
				}
			}
		}

		if !found {
			return true
		}
	}

	return false
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

// discoverRepositoriesFromHosts queries all hosts for available repositories.
func discoverRepositoriesFromHosts(
	hosts []host.Host,
	since *time.Time,
	repoChan chan []host.Repository,
	errChan chan error,
) int {
	expectedFinishes := len(hosts)
	for _, host := range hosts {
		log.Log().Infof("Listing repositories since %v", since)
		go host.ListRepositories(since, repoChan, errChan)
		if since != nil {
			expectedFinishes += 1
			log.Log().Info("Listing repositories with open pull requests")
			go host.ListRepositoriesWithOpenPullRequests(repoChan, errChan)
		}
	}

	return expectedFinishes
}

// discoverRepositoriesFromCLI takes a list of repository names and turns them into repositories.
func discoverRepositoriesFromCLI(
	hosts []host.Host,
	repositoryNames []string,
	repoChan chan []host.Repository,
	errChan chan error,
) int {
	log.Log().Info("Discovering repositories from CLI")
	go func() {
		for _, repoName := range repositoryNames {
			repo, err := findRepositoryInHosts(hosts, repoName)
			if err != nil {
				errChan <- err
				return
			}

			repoChan <- []host.Repository{repo}
		}

		errChan <- nil
	}()
	return 1
}

// findRepositoryInHosts queries all hosts to find the given repository, identified by its name.
func findRepositoryInHosts(hosts []host.Host, repositoryName string) (host.Repository, error) {
	for _, h := range hosts {
		repo, err := h.CreateFromName(repositoryName)
		if err != nil {
			return nil, err
		}

		if repo != nil {
			return repo, nil
		}
	}

	return nil, fmt.Errorf("no host found for repository '%s'", repositoryName)
}

// setInputs sets inputs passed to Run().
// It filters out tasks when an expected input is missing.
func setInputs(tasks []*task.Task, inputs map[string]string) []*task.Task {
	var tasksWithInputs []*task.Task
	for _, t := range tasks {
		err := t.UpdateInputs(inputs)
		if err == nil {
			tasksWithInputs = append(tasksWithInputs, t)
		} else {
			log.Log().Warnf("Deactivating Task due to missing inputs: %s", err)
		}
	}

	return tasksWithInputs
}
