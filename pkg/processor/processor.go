package processor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/action"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
	"go.uber.org/zap"
)

//go:generate stringer -type=Result
type Result int

const (
	ResultUnknown Result = iota
	ResultAutoMergeTooEarly
	ResultBranchModified
	ResultChecksFailed
	ResultConflict
	ResultNoChanges
	ResultPrCreated
	ResultPrClosedBefore
	ResultPrClosed
	ResultPrMergedBefore
	ResultPrMerged
	ResultPrOpen
	ResultNoMatch
	ResultSkip
)

type Processor struct {
	DataDir string
	Git     git.GitClient
}

type RepositoryTaskProcessor interface {
	Process(ctx context.Context, dryRun bool, repo host.Repository, task *task.Task, doFilter bool) (Result, error)
}

func (p *Processor) Process(
	ctx context.Context,
	dryRun bool,
	repo host.Repository,
	task *task.Task,
	doFilter bool,
) (Result, error) {
	ctx = sbcontext.WithRunData(ctx, task.InputData())
	logger := sbcontext.Log(ctx)
	logger.Debug("Processing repository")
	if !task.IsWithinSchedule() {
		logger.Debug("Skipping task because it is outside of schedule")
		return ResultSkip, nil
	}

	if task.HasReachMaxOpenPRs() {
		logger.Debug("Skipping task because Max Open PRs have been reached")
		return ResultSkip, nil
	}

	if task.HasReachedChangeLimit() {
		logger.Debug("Skipping task because Change Limit have been reached")
		return ResultSkip, nil
	}

	ctx = context.WithValue(ctx, sbcontext.RepositoryKey{}, repo)

	if doFilter {
		match, err := matchTaskToRepository(ctx, task, logger)
		if err != nil {
			return ResultUnknown, err
		}

		if !match {
			return ResultNoMatch, nil
		}
	}

	logger.Info("Task matches repository")
	lck := &locker{}
	err := lck.lock(p.DataDir, repo)
	if err != nil {
		return ResultUnknown, fmt.Errorf("lock of repository '%s' failed: %w", repo.FullName(), err)
	}

	defer func() {
		err := lck.unlock()
		if err != nil {
			logger.Error("Failed to unlock repository")
		}
	}()

	workDir, err := p.Git.Prepare(repo, false)
	if err != nil {
		return ResultUnknown, fmt.Errorf("prepare of git repository failed: %w", err)
	}

	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, workDir)
	result, err := applyTaskToRepository(ctx, dryRun, p.Git, logger, repo, task, workDir)
	if err != nil {
		return ResultUnknown, fmt.Errorf("task failed: %w", err)
	}

	if result == ResultPrCreated || result == ResultPrOpen {
		task.IncOpenPRsCount()
	}

	if result == ResultPrCreated || result == ResultPrMerged {
		task.IncChangeLimitCount()
	}

	return result, nil
}

func matchTaskToRepository(ctx context.Context, task *task.Task, logger *zap.SugaredLogger) (bool, error) {
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
			logger.Debugf("Filter %s does not match", filter.String())
			return false, nil
		}
	}

	return true, nil
}

func applyTaskToRepository(ctx context.Context, dryRun bool, gitc git.GitClient, logger *zap.SugaredLogger, repo host.Repository, task *task.Task, workDir string) (Result, error) {
	logger.Debug("Applying actions of task to repository")
	ctx = updateTemplateVars(ctx, repo, task)
	branchName, err := task.RenderBranchName(template.FromContext(ctx))
	if err != nil {
		return ResultUnknown, fmt.Errorf("get branch name: %w", err)
	}

	prID, err := repo.FindPullRequest(branchName)
	if err != nil && !errors.Is(err, host.ErrPullRequestNotFound) {
		return ResultUnknown, fmt.Errorf("find pull request: %w", err)
	}

	if prID != nil && repo.IsPullRequestClosed(prID) && task.MergeOnce {
		logger.Info("Existing PR has been closed")
		return ResultPrClosedBefore, nil
	}

	if prID != nil && repo.IsPullRequestMerged(prID) && task.MergeOnce {
		logger.Info("Existing PR has been merged")
		return ResultPrMergedBefore, nil
	}

	if prID != nil && task.CreateOnly {
		logger.Info("PR exists and is create only")
		return ResultPrOpen, nil
	}

	if prID != nil && repo.IsPullRequestOpen(prID) {
		prInfo := repo.PullRequest(prID)
		if prInfo != nil {
			ctx = context.WithValue(ctx, sbcontext.PullRequestKey{}, *prInfo)
		}

		if task.AutoCloseAfter > 0 && prInfo.CreatedAt != nil {
			dur := time.Duration(task.AutoCloseAfter) * time.Second
			if time.Now().After(prInfo.CreatedAt.Add(dur)) {
				logger.Info("Auto-closing pull request")
				if !dryRun {
					msg := fmt.Sprintf("Pull request has been open for longer than %s. Closing automatically.", dur.String())
					err := repo.ClosePullRequest(msg, prID)
					if err != nil {
						return ResultUnknown, fmt.Errorf("auto-close pull request: %w", err)
					}
				}

				return ResultPrClosed, nil
			}
		}
	}

	forceRebase := prID != nil && needsRebaseByUser(repo, prID)
	if forceRebase {
		// Do not keep the comment around when the user wants to rebase
		logger.Debug("Deleting pull request comment because user requested a force-rebase")
		if !dryRun {
			err := host.DeletePullRequestCommentByIdentifier("branch-modified", prID, repo)
			if err != nil {
				return ResultUnknown, err
			}
		}
	}

	hasConflict, err := gitc.UpdateTaskBranch(branchName, forceRebase, repo)
	if err != nil {
		var branchModifiedErr *git.BranchModifiedError
		if errors.As(err, &branchModifiedErr) && prID != nil && repo.IsPullRequestOpen(prID) {
			logger.Warn("Branch contains commits not made by saturn-bot")
			body, err := template.RenderBranchModified(template.BranchModifiedInput{
				Checksums:     branchModifiedErr.Checksums,
				DefaultBranch: repo.BaseBranch(),
			})
			if err != nil {
				return ResultUnknown, err
			}

			logger.Debug("Creating pull request comment because the branch was modified")
			if !dryRun {
				err := host.CreatePullRequestCommentWithIdentifier(body, "branch-modified", prID, repo)
				if err != nil {
					return ResultUnknown, fmt.Errorf("create comment on merge request: %w", err)
				}
			}

			return ResultBranchModified, nil
		}

		var emptyErr git.EmptyRepositoryError
		if errors.Is(err, emptyErr) {
			logger.Debug("Repository is empty")
			return ResultNoMatch, nil
		}

		return ResultUnknown, fmt.Errorf("update of git branch of task failed: %w", err)
	}

	err = applyActionsInDirectory(task.Actions(), ctx, workDir)
	if err != nil {
		return ResultUnknown, err
	}

	hasLocalChanges, err := gitc.HasLocalChanges()
	if err != nil {
		return ResultUnknown, fmt.Errorf("check for local changes failed: %w", err)
	}

	if hasLocalChanges {
		commitMsg := task.CommitMessage
		err := gitc.CommitChanges(commitMsg)
		if err != nil {
			return ResultUnknown, fmt.Errorf("committing changes failed: %w", err)
		}
	}

	hasChangesInRemoteDefaultBranch, err := gitc.HasRemoteChanges(repo.BaseBranch())
	if err != nil {
		return ResultUnknown, fmt.Errorf("check for remote changes in default branch failed: %w", err)
	}

	if !hasChangesInRemoteDefaultBranch && prID != nil && repo.IsPullRequestOpen(prID) {
		logger.Info("Closing pull request because base branch contains all changes")
		if !dryRun {
			err := repo.ClosePullRequest("Everything up-to-date. Closing.", prID)
			if err != nil {
				return ResultUnknown, fmt.Errorf("close pull request: %w", err)
			}
		}

		logger.Info("Deleting source branch because base branch contains all changes")
		if !dryRun {
			err := repo.DeleteBranch(prID)
			if err != nil {
				return ResultUnknown, fmt.Errorf("delete branch: %w", err)
			}
		}

		err = task.OnPrClosed(ctx)
		if err != nil {
			return ResultUnknown, fmt.Errorf("pr closed event failed: %w", err)
		}

		return ResultPrClosed, nil
	}

	hasRemoteChanges, err := gitc.HasRemoteChanges(branchName)
	if err != nil {
		return ResultUnknown, fmt.Errorf("check for remote changes failed: %w", err)
	}

	hasChanges := (hasLocalChanges && hasRemoteChanges) || hasConflict
	if hasChanges {
		logger.Debug("Pushing changes")
		if !dryRun {
			err := gitc.Push(branchName)
			if err != nil {
				return ResultUnknown, fmt.Errorf("push failed: %w", err)
			}
		}
	} else {
		logger.Info("No changes after applying actions")
	}

	ctx = updateTemplateVars(ctx, repo, task)
	prTitle, err := task.RenderPrTitle(template.FromContext(ctx))
	if err != nil {
		return ResultUnknown, err
	}

	prData := host.PullRequestData{
		Assignees:      task.Assignees,
		AutoMerge:      task.AutoMerge,
		AutoMergeAfter: task.CalcAutoMergeAfter(),
		Body:           task.PrBody,
		Labels:         task.Labels,
		MergeOnce:      task.MergeOnce,
		Reviewers:      task.Reviewers,
		TaskName:       task.Name,
		TemplateData:   template.FromContext(ctx),
		Title:          prTitle,
	}

	// Always create if branch of task contains changes compared to default branch and no PR has been created yet.
	// Create if branch of task contains changes and the PR has been merged or closed before.
	if (hasChangesInRemoteDefaultBranch && prID == nil) || (hasChanges && (prID == nil || repo.IsPullRequestMerged(prID) || repo.IsPullRequestClosed(prID))) {
		logger.Info("Creating pull request")
		if !dryRun {
			err := repo.CreatePullRequest(branchName, prData)
			if err != nil {
				return ResultUnknown, fmt.Errorf("failed to create pull request: %w", err)
			}
		}

		err = task.OnPrCreated(ctx)
		if err != nil {
			return ResultUnknown, fmt.Errorf("pr created event failed: %w", err)
		}

		return ResultPrCreated, nil
	}

	// Try to merge if auto-merge is enabled, no new changes have been detected and the pull request is open
	if task.AutoMerge && !hasChanges && prID != nil && repo.IsPullRequestOpen(prID) {
		success, err := repo.HasSuccessfulPullRequestBuild(prID)
		if err != nil {
			return ResultUnknown, fmt.Errorf("check for successful pull request build failed: %w", err)
		}

		if !success {
			return ResultChecksFailed, nil
		}

		if !canMergeAfter(repo.GetPullRequestCreationTime(prID), task.CalcAutoMergeAfter()) {
			logger.Info("Too early to merge pull request")
			return ResultAutoMergeTooEarly, nil
		}

		canMerge, err := repo.CanMergePullRequest(prID)
		if err != nil {
			return ResultUnknown, fmt.Errorf("check if pull request can be merged: %w", err)
		}

		if !canMerge {
			logger.Warn("Cannot merge pull request")
			return ResultConflict, nil
		}

		logger.Info("Merging pull request")
		if !dryRun {
			err := repo.MergePullRequest(!task.KeepBranchAfterMerge, prID)
			if err != nil {
				return ResultUnknown, fmt.Errorf("failed to merge pull request: %w", err)
			}
		}

		err = task.OnPrMerged(ctx)
		if err != nil {
			return ResultUnknown, fmt.Errorf("pr merged event failed: %w", err)
		}

		return ResultPrMerged, nil
	}

	if prID != nil && repo.IsPullRequestOpen(prID) {
		logger.Debug("Updating pull request")
		if !dryRun {
			err := repo.UpdatePullRequest(prData, prID)
			if err != nil {
				return ResultUnknown, fmt.Errorf("failed to update pull request: %w", err)
			}
		}

		return ResultPrOpen, nil
	}

	return ResultNoChanges, nil
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

func updateTemplateVars(ctx context.Context, repo host.Repository, tk *task.Task) context.Context {
	data := template.FromContext(ctx)
	runData := sbcontext.RunData(ctx)
	for k, v := range runData {
		data.Run[k] = v
	}

	data.Repository.FullName = repo.FullName()
	data.Repository.Host = repo.Host().Name()
	data.Repository.Name = repo.Name()
	data.Repository.Owner = repo.Owner()
	data.Repository.WebUrl = repo.WebUrl()
	if tk != nil {
		data.TaskName = tk.Name
	}

	return template.UpdateContext(ctx, data)
}
