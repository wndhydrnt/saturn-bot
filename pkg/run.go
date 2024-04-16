package pkg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/wndhydrnt/saturn-sync/pkg/action"
	"github.com/wndhydrnt/saturn-sync/pkg/cache"
	"github.com/wndhydrnt/saturn-sync/pkg/config"
	sContext "github.com/wndhydrnt/saturn-sync/pkg/context"
	"github.com/wndhydrnt/saturn-sync/pkg/git"
	"github.com/wndhydrnt/saturn-sync/pkg/host"
	sLog "github.com/wndhydrnt/saturn-sync/pkg/log"
	"github.com/wndhydrnt/saturn-sync/pkg/task"
	"github.com/wndhydrnt/saturn-sync/pkg/template"
)

var (
	ErrNoHostsConfigured = errors.New("no hosts configured")
)

type ApplyResult int

const (
	ApplyResultUnknown ApplyResult = iota
	ApplyResultAutoMergeTooEarly
	ApplyResultBranchModified
	ApplyResultChecksFailed
	ApplyResultConflict
	ApplyResultNoChanges
	ApplyResultPrCreated
	ApplyResultPrClosedBefore
	ApplyResultPrClosed
	ApplyResultPrMergedBefore
	ApplyResultPrMerged
	ApplyResultPrOpen
)

type executeRunner struct {
	applyTaskFunc func(ctx context.Context, dryRun bool, gitc git.GitClient, logger *slog.Logger, repo host.Repository, task task.Task, workDir string) (ApplyResult, error)
	cache         cache.Cache
	dryRun        bool
	git           git.GitClient
	hosts         []host.Host
	taskRegistry  *task.Registry
}

func (r *executeRunner) run(taskFiles []string) error {
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

	r.taskRegistry.ReadAll(taskFiles)
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
	expectedFinishes := len(r.hosts)
	for _, host := range r.hosts {
		slog.Info("Listing repositories", "updated_since", fmt.Sprintf("%v", since))
		go host.ListRepositories(since, repos, errChan)
		if since != nil {
			expectedFinishes += 1
			slog.Info("Listing repositories with open pull requests")
			go host.ListRepositoriesWithOpenPullRequests(repos, errChan)
		}
	}
	finishes := 0
	visitedRepositories := map[string]struct{}{}
	success := true
	for {
		select {
		case repoList := <-repos:
			for _, repo := range repoList {
				_, exists := visitedRepositories[repo.FullName()]
				if exists {
					slog.Debug("Repository already visited", "repository", repo.FullName())
					continue
				}

				visitedRepositories[repo.FullName()] = struct{}{}
				slog.Debug("Processing repository", "repository", repo.FullName())
				ctx := context.WithValue(context.Background(), sContext.RepositoryKey{}, repo)
				tasksToApply := findMatchingTasksForRepository(ctx, repo, tasks)
				if len(tasksToApply) < 1 {
					slog.Debug("No task matches the repository", "repository", repo.FullName())
					continue
				}

				workDir, err := r.git.Prepare(repo, false)
				if err != nil {
					return fmt.Errorf("prepare of git repository failed: %w", err)
				}

				ctx = context.WithValue(ctx, sContext.CheckoutPath{}, workDir)
				for _, taskToApply := range tasksToApply {
					logger := slog.With("dryRun", r.dryRun, "repository", repo.FullName(), "task", taskToApply.SourceTask().GetName())
					logger.Info("Task matches repository")
					_, err := r.applyTaskFunc(ctx, r.dryRun, r.git, logger, repo, taskToApply, workDir)
					if err != nil {
						success = false
						logger.Error("Task failed", "err", err)
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

	r.cache.SetLastExecutionAt(time.Now().UnixMicro())
	r.cache.UpdateCachedTasks(tasks)
	err = r.cache.Write()
	if err != nil {
		return err
	}

	r.taskRegistry.Stop()

	if !success {
		return fmt.Errorf("errors occurred, check previous log messages")
	}
	return nil
}

func ExecuteRun(cfgFile string, taskFiles []string) error {
	cfg, err := config.Read(cfgFile)
	if err != nil {
		return err
	}

	sLog.InitLog(cfg.LogFormat, cfg.LogLevel, cfg.GitLogLevel)
	cache := cache.NewJsonFile(path.Join(cfg.DataDir, cache.DefaultJsonFileName))
	taskRegistry := task.NewRegistry(cfg.CustomMarshaled())

	hosts, err := createHostsFromConfig(cfg)
	if err != nil {
		return err
	}

	gitClient, err := git.New(cfg)
	if err != nil {
		return fmt.Errorf("new git client for run: %w", err)
	}

	e := &executeRunner{
		applyTaskFunc: applyTaskToRepository,
		cache:         cache,
		dryRun:        cfg.DryRun,
		git:           gitClient,
		hosts:         hosts,
		taskRegistry:  taskRegistry,
	}
	return e.run(taskFiles)
}

func hasUpdatedTasks(cachedTasks []cache.CachedTask, tasks []task.Task) bool {
	for _, t := range tasks {
		found := false
		for _, ct := range cachedTasks {
			if t.SourceTask().GetName() == ct.Name {
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

func findMatchingTasksForRepository(ctx context.Context, repository host.Repository, tasks []task.Task) []task.Task {
	var matchingTasks []task.Task
	for _, t := range tasks {
		match, err := matchTaskToRepository(ctx, t)
		if err != nil {
			slog.Error("Filter of task failed - skipping", "err", err, "task", t.SourceTask().GetName(), "repository", repository.FullName())
			continue
		}

		if match {
			matchingTasks = append(matchingTasks, t)
		}
	}

	return matchingTasks
}

func matchTaskToRepository(ctx context.Context, task task.Task) (bool, error) {
	if len(task.Filters()) == 0 {
		// A task without filters is considered not matching.
		// Avoids accidentally applying a task to all repositories
		// because no filters are set.
		return false, nil
	}

	for _, filter := range task.Filters() {
		match, err := filter.Do(ctx)
		if err != nil {
			return false, fmt.Errorf("filter %s failed: %w", filter.String(), err)
		}

		if !match {
			slog.Debug("Filter does not match", "filter", filter.String(), "task", task.SourceTask().GetName())
			return false, nil
		}
	}

	return true, nil
}

func applyTaskToRepository(ctx context.Context, dryRun bool, gitc git.GitClient, logger *slog.Logger, repo host.Repository, task task.Task, workDir string) (ApplyResult, error) {
	logger.Debug("Applying actions of task to repository")
	prID, err := repo.FindPullRequest(task.BranchName())
	if err != nil && !errors.Is(err, host.ErrPullRequestNotFound) {
		return ApplyResultUnknown, fmt.Errorf("find pull request: %w", err)
	}

	if prID != nil && repo.IsPullRequestClosed(prID) && task.SourceTask().GetMergeOnce() {
		logger.Info("Existing PR has been closed")
		return ApplyResultPrClosedBefore, nil
	}

	if prID != nil && repo.IsPullRequestMerged(prID) && task.SourceTask().GetMergeOnce() {
		logger.Info("Existing PR has been merged")
		return ApplyResultPrMergedBefore, nil
	}

	if prID != nil && task.SourceTask().GetCreateOnly() {
		logger.Info("PR exists and is create only")
		return ApplyResultPrOpen, nil
	}

	forceRebase := prID != nil && needsRebaseByUser(repo, prID)
	if forceRebase {
		// Do not keep the comment around when the user wants to rebase
		logger.Debug("Deleting pull request comment because user requested a force-rebase")
		if !dryRun {
			err := host.DeletePullRequestCommentByIdentifier("branch-modified", prID, repo)
			if err != nil {
				return ApplyResultUnknown, err
			}
		}
	}

	hasConflict, err := gitc.UpdateTaskBranch(task.BranchName(), forceRebase, repo)
	if err != nil {
		var branchModifiedErr *git.BranchModifiedError
		if errors.As(err, &branchModifiedErr) && prID != nil && repo.IsPullRequestOpen(prID) {
			logger.Warn("Branch contains commits not made by saturn-sync")
			body, err := template.RenderBranchModified(template.BranchModifiedInput{
				Checksums:     branchModifiedErr.Checksums,
				DefaultBranch: repo.BaseBranch(),
			})
			if err != nil {
				return ApplyResultUnknown, err
			}

			logger.Debug("Creating pull request comment because the branch was modified")
			if !dryRun {
				err := host.CreatePullRequestCommentWithIdentifier(body, "branch-modified", prID, repo)
				if err != nil {
					return ApplyResultUnknown, fmt.Errorf("create comment on merge request: %w", err)
				}
			}

			return ApplyResultBranchModified, nil
		}

		return ApplyResultUnknown, fmt.Errorf("update of git branch of task failed: %w", err)
	}

	err = applyActionsInDirectory(task.Actions(), ctx, workDir)
	if err != nil {
		return ApplyResultUnknown, err
	}

	hasLocalChanges, err := gitc.HasLocalChanges()
	if err != nil {
		return ApplyResultUnknown, fmt.Errorf("check for local changes failed: %w", err)
	}

	if hasLocalChanges {
		commitMsg := task.SourceTask().GetCommitMessage()
		err := gitc.CommitChanges(commitMsg)
		if err != nil {
			return ApplyResultUnknown, fmt.Errorf("committing changes failed: %w", err)
		}
	}

	hasChangesInRemoteDefaultBranch, err := gitc.HasRemoteChanges(repo.BaseBranch())
	if err != nil {
		return ApplyResultUnknown, fmt.Errorf("check for remote changes in default branch failed: %w", err)
	}

	if !hasChangesInRemoteDefaultBranch && prID != nil && repo.IsPullRequestOpen(prID) {
		logger.Info("Closing pull request because base branch contains all changes")
		if !dryRun {
			err := repo.ClosePullRequest("Everything up-to-date. Closing.", prID)
			if err != nil {
				return ApplyResultUnknown, fmt.Errorf("close pull request: %w", err)
			}
		}

		logger.Info("Deleting source branch because base branch contains all changes")
		if !dryRun {
			err := repo.DeleteBranch(prID)
			if err != nil {
				return ApplyResultUnknown, fmt.Errorf("delete branch: %w", err)
			}
		}

		err = task.OnPrClosed(repo)
		if err != nil {
			return ApplyResultUnknown, fmt.Errorf("pr closed event failed: %w", err)
		}

		return ApplyResultPrClosed, nil
	}

	hasRemoteChanges, err := gitc.HasRemoteChanges(task.BranchName())
	if err != nil {
		return ApplyResultUnknown, fmt.Errorf("check for remote changes failed: %w", err)
	}

	hasChanges := (hasLocalChanges && hasRemoteChanges) || hasConflict
	if hasChanges {
		logger.Debug("Pushing changes")
		if !dryRun {
			err := gitc.Push(task.BranchName())
			if err != nil {
				return ApplyResultUnknown, fmt.Errorf("push failed: %w", err)
			}
		}
	} else {
		logger.Info("No changes after applying actions")
	}

	autoMergeAfter := task.AutoMergeAfter()
	prData := host.PullRequestData{
		AutoMerge:      task.SourceTask().GetAutoMerge(),
		AutoMergeAfter: &autoMergeAfter,
		Body:           task.SourceTask().GetPrBody(),
		Labels:         task.SourceTask().GetLabels(),
		MergeOnce:      task.SourceTask().GetMergeOnce(),
		TaskName:       task.SourceTask().GetName(),
		TemplateData: map[string]any{
			"RepositoryFullName": repo.FullName(),
			"RepositoryHost":     repo.Host(),
			"RepositoryName":     repo.Name(),
			"RepositoryOwner":    repo.Owner(),
			"RepositoryWebUrl":   repo.WebUrl(),
			"TaskName":           task.SourceTask().GetName(),
		},
		Title: task.SourceTask().GetPrTitle(),
	}

	// Always create if branch of task contains changes compared to default branch and no PR has been created yet.
	// Create if branch of task contains changes and the PR has been merged or closed before.
	if (hasChangesInRemoteDefaultBranch && prID == nil) || (hasChanges && (prID == nil || repo.IsPullRequestMerged(prID) || repo.IsPullRequestClosed(prID))) {
		logger.Info("Creating pull request")
		if !dryRun {
			err := repo.CreatePullRequest(task.BranchName(), prData)
			if err != nil {
				return ApplyResultUnknown, fmt.Errorf("failed to create pull request: %w", err)
			}
		}

		err = task.OnPrCreated(repo)
		if err != nil {
			return ApplyResultUnknown, fmt.Errorf("pr created event failed: %w", err)
		}

		return ApplyResultPrCreated, nil
	}

	// Try to merge if auto-merge is enabled, no new changes have been detected and the pull request is open
	if task.SourceTask().GetAutoMerge() && !hasChanges && prID != nil && repo.IsPullRequestOpen(prID) {
		success, err := repo.HasSuccessfulPullRequestBuild(prID)
		if err != nil {
			return ApplyResultUnknown, fmt.Errorf("check for successful pull request build failed: %w", err)
		}

		if !success {
			return ApplyResultChecksFailed, nil
		}

		if !canMergeAfter(repo.GetPullRequestCreationTime(prID), task.AutoMergeAfter()) {
			logger.Info("Too early to merge pull request")
			return ApplyResultAutoMergeTooEarly, nil
		}

		canMerge, err := repo.CanMergePullRequest(prID)
		if err != nil {
			return ApplyResultUnknown, fmt.Errorf("check if pull request can be merged: %w", err)
		}

		if !canMerge {
			logger.Warn("Cannot merge pull request")
			return ApplyResultConflict, nil
		}

		logger.Info("Merging pull request")
		if !dryRun {
			err := repo.MergePullRequest(!task.SourceTask().GetKeepBranchAfterMerge(), prID)
			if err != nil {
				return ApplyResultUnknown, fmt.Errorf("failed to merge pull request: %w", err)
			}
		}

		err = task.OnPrMerged(repo)
		if err != nil {
			return ApplyResultUnknown, fmt.Errorf("pr merged event failed: %w", err)
		}

		return ApplyResultPrMerged, nil
	}

	if prID != nil && repo.IsPullRequestOpen(prID) {
		logger.Debug("Updating pull request")
		if !dryRun {
			err := repo.UpdatePullRequest(prData, prID)
			if err != nil {
				return ApplyResultUnknown, fmt.Errorf("failed to update pull request: %w", err)
			}
		}

		return ApplyResultPrOpen, nil
	}

	return ApplyResultNoChanges, nil
}

func needsRebaseByUser(repo host.Repository, pr any) bool {
	body := repo.GetPullRequestBody(pr)
	return strings.Contains(body, "[x] If you want to rebase this PR")
}

func canMergeAfter(createdAt time.Time, mergeAfter time.Duration) bool {
	cutoff := createdAt.Add(mergeAfter)
	now := time.Now().UTC()
	return now.After(cutoff)
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
