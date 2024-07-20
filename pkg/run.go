package pkg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
	sContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
)

var (
	ErrNoHostsConfigured = errors.New("no hosts configured")
)

type executeRunner struct {
	cache        cache.Cache
	dryRun       bool
	hosts        []host.Host
	processor    processor.RepositoryTaskProcessor
	taskRegistry *task.Registry
}

func (r *executeRunner) run(repositoryNames, taskFiles []string) error {
	if len(r.hosts) == 0 {
		return ErrNoHostsConfigured
	}

	err := r.cache.Read()
	if err != nil {
		return err
	}

	var since *time.Time
	if r.cache.GetLastExecutionAt() != 0 {
		ts := time.UnixMicro(r.cache.GetLastExecutionAt())
		since = &ts
	}

	err = r.taskRegistry.ReadAll(taskFiles)
	if err != nil {
		return err
	}

	tasks := r.taskRegistry.GetTasks()
	if len(tasks) == 0 {
		slog.Warn("0 tasks loaded from files - stopping")
		return nil
	}

	needsAllRepositories := hasUpdatedTasks(r.cache.GetCachedTasks(), tasks)
	if needsAllRepositories {
		since = nil
	}

	repos := make(chan []host.Repository)
	errChan := make(chan error)
	var expectedFinishes int
	if len(repositoryNames) > 0 {
		expectedFinishes = discoverRepositoriesFromCLI(r.hosts, repositoryNames, repos, errChan)
	} else {
		expectedFinishes = discoverRepositoriesFromHosts(r.hosts, since, repos, errChan)
	}
	finishes := 0
	visitedRepositories := map[string]struct{}{}
	success := true
	for {
		select {
		case repoList := <-repos:
			for _, repo := range repoList {
				slog.Debug("Repository discovered", "repository", repo.FullName())
				_, exists := visitedRepositories[repo.FullName()]
				if exists {
					slog.Debug("Repository already visited", "repository", repo.FullName())
					continue
				}

				visitedRepositories[repo.FullName()] = struct{}{}
				ctx := context.Background()
				ctx = context.WithValue(ctx, sContext.TemplateVarsKey{}, make(map[string]string))
				doFilter := len(repositoryNames) == 0
				for _, t := range tasks {
					_, err := r.processor.Process(ctx, r.dryRun, repo, t, doFilter)
					if err != nil {
						success = false
						slog.Error("Task failed", "err", err)
					}
				}
			}
		case err := <-errChan:
			if err != nil {
				return err
			}

			finishes += 1
		}
		if finishes == expectedFinishes {
			break
		}
	}

	if !r.dryRun {
		// Only update cache if this is not a dry run.
		// Without this guard, subsequent non dry runs would not recognize that they need to do anything.
		r.cache.SetLastExecutionAt(time.Now().UnixMicro())
		r.cache.UpdateCachedTasks(tasks)
		err = r.cache.Write()
		if err != nil {
			return err
		}
	}

	r.taskRegistry.Stop()

	if !success {
		return fmt.Errorf("errors occurred, check previous log messages")
	}
	slog.Info("Run finished")
	return nil
}

func ExecuteRun(opts options.Opts, repositoryNames, taskFiles []string) error {
	err := options.Initialize(opts)
	if err != nil {
		return fmt.Errorf("initialize options: %w", err)
	}

	cache := cache.NewJsonFile(path.Join(*opts.Config.DataDir, cache.DefaultJsonFileName))
	taskRegistry := task.NewRegistry(opts)

	gitClient, err := git.New(opts.Config)
	if err != nil {
		return fmt.Errorf("new git client for run: %w", err)
	}

	e := &executeRunner{
		cache:        cache,
		dryRun:       opts.Config.DryRun,
		hosts:        opts.Hosts,
		processor:    &processor.Processor{Git: gitClient},
		taskRegistry: taskRegistry,
	}
	return e.run(repositoryNames, taskFiles)
}

func hasUpdatedTasks(cachedTasks []cache.CachedTask, tasks []task.Task) bool {
	for _, t := range tasks {
		found := false
		for _, ct := range cachedTasks {
			if t.SourceTask().Name == ct.Name {
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
		slog.Info("Listing repositories", "updated_since", fmt.Sprintf("%v", since))
		go host.ListRepositories(since, repoChan, errChan)
		if since != nil {
			expectedFinishes += 1
			slog.Info("Listing repositories with open pull requests")
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
	slog.Info("Discovering repositories from CLI")
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
