package processor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/action"
	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
	"go.uber.org/zap"
)

//go:generate go run --modfile=../../tools/go.mod golang.org/x/tools/cmd/stringer -type=Result
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
	ResultPushedDefaultBranch
)

type ProcessResult struct {
	Error       error
	PullRequest *host.PullRequest
	Result      Result
	Task        *task.Task
}

type Processor struct {
	DataDir          string
	Git              git.GitClient
	PullRequestCache *host.PullRequestCache
}

type RepositoryTaskProcessor interface {
	Process(dryRun bool, repo host.Repository, tasks []*task.Task, doFilter bool) []ProcessResult
}

func (p *Processor) Process(dryRun bool, repo host.Repository, tasks []*task.Task, doFilter bool) []ProcessResult {
	ctx := context.WithValue(context.Background(), sbcontext.RepositoryKey{}, repo)
	logger := log.Log().
		WithOptions(zap.Fields(
			log.FieldDryRun(dryRun),
			log.FieldRepo(repo.FullName()),
		))
	var results []ProcessResult
	var tasksAfterPreCloneFilters []*task.Task
	for _, t := range tasks {
		if !doFilter {
			tasksAfterPreCloneFilters = append(tasksAfterPreCloneFilters, t)
			continue
		}

		taskLogger := logger.With(log.FieldTask(t.Name))
		taskCtx := sbcontext.WithLog(ctx, taskLogger)
		taskCtx = sbcontext.WithRunData(taskCtx, t.RunData())
		match, preCloneResult, err := p.filterPreClone(taskCtx, t)
		result := ProcessResult{
			Task: t,
		}
		if err != nil {
			result.Error = err
			result.Result = preCloneResult
			results = append(results, result)
			taskLogger.Errorw("Task failed", "error", result.Error)
			continue
		}

		if !match {
			pr, err := p.handleFilteredRepository(taskCtx, t, repo, preCloneResult)
			if err != nil {
				taskLogger.Errorw("Failed to handle filtered task in prefilter", zap.Error(err))
			}

			if pr != nil {
				result.PullRequest = pr
			}

			result.Result = preCloneResult
			results = append(results, result)
		} else {
			tasksAfterPreCloneFilters = append(tasksAfterPreCloneFilters, t)
		}
	}

	if len(tasksAfterPreCloneFilters) == 0 {
		return results
	}

	checkoutDir, err := p.Git.Prepare(repo, false)
	if err != nil {
		// An error during the preparation of the git repository is the best indicator that
		// the repository has been deleted.
		// Log a warning, clean up and consider the repository as "not matching".
		log.Log().Warnf("Failed to clone or pull repository '%s' - cleaning up the repository", repo.FullName())
		err := p.Git.Cleanup(repo)
		if err != nil {
			log.Log().Errorf("Failed to clean up repository '%s'", repo.FullName())
		}

		for _, t := range tasksAfterPreCloneFilters {
			results = append(results, ProcessResult{
				Result: ResultNoMatch,
				Task:   t,
			})
		}

		return results
	}

	ctx = context.WithValue(ctx, sbcontext.CheckoutPath{}, checkoutDir)
	for _, t := range tasksAfterPreCloneFilters {
		taskLogger := logger.
			WithOptions(zap.Fields(
				log.FieldTask(t.Name),
			))
		taskCtx := sbcontext.WithLog(ctx, taskLogger)
		taskCtx = sbcontext.WithRunData(taskCtx, t.RunData())
		resultId, pr, err := p.processPostClone(taskCtx, repo, t, doFilter, dryRun)
		result := ProcessResult{
			PullRequest: pr,
			Result:      resultId,
			Task:        t,
		}
		if err != nil {
			result.Error = err
			taskLogger.Errorw("Task failed", "error", result.Error)
		}

		_ = p.updatePrCache(taskCtx, t, repo, pr)
		results = append(results, result)
	}

	return results
}

func (p *Processor) filterPreClone(ctx context.Context, task *task.Task) (bool, Result, error) {
	logger := sbcontext.Log(ctx)
	if task.HasReachMaxOpenPRs() {
		logger.Debug("Skipping task because Max Open PRs have been reached")
		return false, ResultSkip, nil
	}

	if task.HasReachedChangeLimit() {
		logger.Debug("Skipping task because Change Limit have been reached")
		return false, ResultSkip, nil
	}

	if !task.HasFilters() {
		// A task without filters is considered not matching.
		// Avoids accidentally applying a task to all repositories
		// because no filters are set.
		return false, ResultNoMatch, nil
	}

	match, err := matchTaskToRepository(ctx, task.FiltersPreClone(), logger)
	if err != nil {
		return false, ResultUnknown, err
	}

	if !match {
		return false, ResultNoMatch, nil
	}

	return true, 0, nil
}

func (p *Processor) processPostClone(ctx context.Context, repo host.Repository, task *task.Task, doFilter, dryRun bool) (Result, *host.PullRequest, error) {
	lck := &locker{}
	err := lck.lock(p.DataDir, repo)
	if err != nil {
		return ResultUnknown, nil, fmt.Errorf("lock of repository '%s' failed: %w", repo.FullName(), err)
	}

	logger := sbcontext.Log(ctx)
	defer func() {
		err := lck.unlock()
		if err != nil {
			logger.Error("Failed to unlock repository")
		}
	}()

	if doFilter {
		match, err := matchTaskToRepository(ctx, task.FiltersPostClone(), logger)
		if err != nil {
			return ResultUnknown, nil, err
		}

		if !match {
			result := ResultNoMatch
			pr, err := p.handleFilteredRepository(ctx, task, repo, result)
			if err != nil {
				logger.Errorw("Failed to handle filtered task in postfilter", zap.Error(err))
			}

			return result, pr, nil
		}
	}

	logger.Info("Task matches repository")
	checkoutDir := ctx.Value(sbcontext.CheckoutPath{}).(string)
	result, prDetail, err := p.applyTaskToRepository(ctx, dryRun, p.Git, logger, repo, task, checkoutDir)
	if err != nil {
		return ResultUnknown, prDetail, fmt.Errorf("task failed: %w", err)
	}

	if IsPrOpen(result) {
		task.IncOpenPRsCount()
	}

	if result == ResultPrCreated || result == ResultPrMerged || result == ResultPushedDefaultBranch {
		task.IncChangeLimitCount()
	}

	return result, prDetail, nil
}

// handleFilteredRepository wraps other functions that should be executed when a repository doesn't match the filters of the current task.
//
// It returns the [host.PullRequest], if it exists in the cache, for further processing.
//
// It only acts if the [host.PullRequestCache] contains the PR. It is a no-op if the cache is empty.
// Callers of the [Processor] need to ensure that the cache is up-to-date.
func (p *Processor) handleFilteredRepository(ctx context.Context, t *task.Task, repo host.Repository, result Result) (*host.PullRequest, error) {
	if p.PullRequestCache == nil {
		return nil, nil
	}

	branchName, err := t.RenderBranchName(template.FromContext(ctx))
	if err != nil {
		return nil, err
	}

	cachedPr := p.PullRequestCache.Get(branchName, repo.FullName())
	if cachedPr == nil {
		return nil, nil
	}

	defer p.PullRequestCache.Delete(branchName, repo.FullName())

	err = closePrForNonMatchingRepo(cachedPr, repo, result)
	if err != nil {
		return nil, fmt.Errorf("close pull request of non-matching task: %w", err)
	}

	return cachedPr, nil
}

// updatePrCache updates a Pull Request in the cache.
// It performs no action of no cache is defined.
func (p *Processor) updatePrCache(ctx context.Context, t *task.Task, repo host.Repository, pr *host.PullRequest) error {
	if p.PullRequestCache == nil {
		return nil
	}

	if pr == nil {
		return nil
	}

	branchName, err := t.RenderBranchName(template.FromContext(ctx))
	if err != nil {
		return fmt.Errorf("render branch name to update pr cache: %w", err)
	}

	p.PullRequestCache.Set(branchName, repo.FullName(), pr)
	return nil
}

// closePrForNonMatchingRepo tries to close an open Pull Request for a repository that has been filtered out.
// A repository can be filtered out while the PR created by saturn-bot is still open if the user has
// changed the filters and the repository doesn't match the updated filters.
//
// It updates the state of pr.
func closePrForNonMatchingRepo(pr *host.PullRequest, repo host.Repository, result Result) error {
	if pr.State == host.PullRequestStateOpen && (result == ResultNoMatch || result == ResultSkip) {
		err := repo.ClosePullRequest("", pr)
		if err != nil {
			return err
		}

		pr.State = host.PullRequestStateClosed
		return nil
	}

	return nil
}

func matchTaskToRepository(ctx context.Context, filters []filter.Filter, logger *zap.SugaredLogger) (bool, error) {
	for _, filter := range filters {
		match, err := filter.Do(ctx)
		if err != nil {
			return false, fmt.Errorf("filter %s failed: %w", filter.String(), err)
		}

		if !match {
			return false, nil
		}
	}

	return true, nil
}

func applyTaskToDefaultBranch(ctx context.Context, dryRun bool, gitc git.GitClient, logger *zap.SugaredLogger, repo host.Repository, task *task.Task, workDir string) (Result, error) {
	_, _, err := gitc.Execute("checkout", repo.BaseBranch())
	if err != nil {
		var gitErr *git.GitCommandError
		if errors.As(err, &gitErr) {
			if strings.Contains(gitErr.Error(), "did not match any file(s) known to git") {
				logger.Debug("Repository is empty")
				return ResultNoMatch, nil
			}
		}

		return ResultUnknown, fmt.Errorf("checkout default branch %s: %w", repo.BaseBranch(), err)
	}

	err = applyActionsInDirectory(task.Actions(), ctx, workDir)
	if err != nil {
		return ResultUnknown, err
	}

	hasLocalChanges, err := gitc.HasLocalChanges()
	if err != nil {
		return ResultUnknown, fmt.Errorf("check for local changes in default branch failed: %w", err)
	}

	if !hasLocalChanges {
		return ResultNoChanges, nil
	}

	if !dryRun {
		err = gitc.CommitChanges(task.CommitMessage)
		if err != nil {
			return ResultUnknown, fmt.Errorf("committing changes to default branch failed: %w", err)
		}

		logger.Debug("Pushing changes to default branch")
		err = gitc.Push(repo.BaseBranch(), false)
		if err != nil {
			return ResultUnknown, fmt.Errorf("push to default branch failed: %w", err)
		}
	}

	return ResultPushedDefaultBranch, nil
}

func (p *Processor) applyTaskToRepository(ctx context.Context, dryRun bool, gitc git.GitClient, logger *zap.SugaredLogger, repo host.Repository, task *task.Task, workDir string) (Result, *host.PullRequest, error) {
	if task.PushToDefaultBranch {
		result, err := applyTaskToDefaultBranch(ctx, dryRun, gitc, logger, repo, task, workDir)
		return result, nil, err
	}

	logger.Debug("Applying actions of task to repository")
	ctx = updateTemplateVars(ctx, repo, task)
	branchName, err := task.RenderBranchName(template.FromContext(ctx))
	if err != nil {
		return ResultUnknown, nil, fmt.Errorf("get branch name: %w", err)
	}

	prID, err := p.findPullRequest(branchName, repo)
	if err != nil && !errors.Is(err, host.ErrPullRequestNotFound) {
		return ResultUnknown, nil, fmt.Errorf("find pull request: %w", err)
	}

	if prID != nil && prID.State == host.PullRequestStateClosed {
		if task.MergeOnce {
			logger.Info("Existing PR has been closed")
			return ResultPrClosedBefore, prID, nil
		} else {
			logger.Debug("Previous pull request closed - resetting to create a new pull request")
			prID = nil
		}
	}

	if prID != nil && prID.State == host.PullRequestStateMerged && task.MergeOnce {
		logger.Info("Existing PR has been merged")
		return ResultPrMergedBefore, prID, nil
	}

	if prID != nil && task.CreateOnly {
		logger.Info("PR exists and is create only")
		return ResultPrOpen, prID, nil
	}

	if prID != nil && prID.State == host.PullRequestStateOpen {
		ctx = context.WithValue(ctx, sbcontext.PullRequestKey{}, *prID)

		if task.AutoCloseAfter > 0 && !prID.CreatedAt.IsZero() {
			dur := time.Duration(task.AutoCloseAfter) * time.Second
			if time.Now().After(prID.CreatedAt.Add(dur)) {
				logger.Info("Auto-closing pull request")
				if !dryRun {
					msg := fmt.Sprintf("Pull request has been open for longer than %s. Closing automatically.", dur.String())
					err := repo.ClosePullRequest(msg, prID)
					if err != nil {
						return ResultUnknown, prID, fmt.Errorf("auto-close pull request: %w", err)
					}
				}

				return ResultPrClosed, prID, nil
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
				return ResultUnknown, prID, err
			}
		}
	}

	// If no PR is currently open, always force-rebase.
	// This ensures that commits made by users to a
	// branch created by previous runs get removed.
	if prID != nil && prID.State != host.PullRequestStateOpen {
		forceRebase = true
	}

	hasConflict, err := gitc.UpdateTaskBranch(branchName, forceRebase, repo)
	if err != nil {
		var branchModifiedErr *git.BranchModifiedError
		if errors.As(err, &branchModifiedErr) && prID != nil && prID.State == host.PullRequestStateOpen {
			logger.Warn("Branch contains commits not made by saturn-bot")
			body, err := template.RenderBranchModified(template.BranchModifiedInput{
				Checksums:     branchModifiedErr.Checksums,
				DefaultBranch: repo.BaseBranch(),
			})
			if err != nil {
				return ResultUnknown, prID, err
			}

			logger.Debug("Creating pull request comment because the branch was modified")
			if !dryRun {
				err := host.CreatePullRequestCommentWithIdentifier(body, "branch-modified", prID, repo)
				if err != nil {
					return ResultUnknown, prID, fmt.Errorf("create comment on merge request: %w", err)
				}
			}

			return ResultBranchModified, prID, nil
		}

		var emptyErr git.EmptyRepositoryError
		if errors.Is(err, emptyErr) {
			logger.Debug("Repository is empty")
			return ResultNoMatch, prID, nil
		}

		return ResultUnknown, prID, fmt.Errorf("update of git branch of task failed: %w", err)
	}

	err = applyActionsInDirectory(task.Actions(), ctx, workDir)
	if err != nil {
		return ResultUnknown, prID, err
	}

	hasLocalChanges, err := gitc.HasLocalChanges()
	if err != nil {
		return ResultUnknown, prID, fmt.Errorf("check for local changes failed: %w", err)
	}

	if hasLocalChanges {
		commitMsg := task.CommitMessage
		err := gitc.CommitChanges(commitMsg)
		if err != nil {
			return ResultUnknown, prID, fmt.Errorf("committing changes failed: %w", err)
		}
	}

	hasChangesInRemoteDefaultBranch, err := gitc.HasRemoteChanges(repo.BaseBranch())
	if err != nil {
		return ResultUnknown, prID, fmt.Errorf("check for remote changes in default branch failed: %w", err)
	}

	if !hasChangesInRemoteDefaultBranch && prID != nil && prID.State == host.PullRequestStateOpen {
		logger.Info("Closing pull request because base branch contains all changes")
		if !dryRun {
			err := repo.ClosePullRequest("Everything up-to-date. Closing.", prID)
			if err != nil {
				return ResultUnknown, prID, fmt.Errorf("close pull request: %w", err)
			}
		}

		logger.Info("Deleting source branch because base branch contains all changes")
		if !dryRun {
			err := repo.DeleteBranch(prID)
			if err != nil {
				return ResultUnknown, prID, fmt.Errorf("delete branch: %w", err)
			}
		}

		err = task.OnPrClosed(ctx)
		if err != nil {
			return ResultUnknown, prID, fmt.Errorf("pr closed event failed: %w", err)
		}

		return ResultPrClosed, prID, nil
	}

	hasRemoteChanges, err := gitc.HasRemoteChanges(branchName)
	if err != nil {
		return ResultUnknown, prID, fmt.Errorf("check for remote changes failed: %w", err)
	}

	hasChanges := (hasLocalChanges && hasRemoteChanges) || hasConflict
	if hasChanges {
		logger.Debug("Pushing changes")
		if !dryRun {
			err := gitc.Push(branchName, true)
			if err != nil {
				return ResultUnknown, prID, fmt.Errorf("push failed: %w", err)
			}
		}
	} else {
		logger.Info("No changes after applying actions")
	}

	ctx = updateTemplateVars(ctx, repo, task)
	prTitle, err := task.RenderPrTitle(template.FromContext(ctx))
	if err != nil {
		return ResultUnknown, prID, err
	}

	prData := host.PullRequestData{
		Assignees:      getAssignees(ctx, task),
		AutoMerge:      task.AutoMerge,
		AutoMergeAfter: task.CalcAutoMergeAfter(),
		Body:           task.PrBody,
		Labels:         task.Labels,
		MergeOnce:      task.MergeOnce,
		Reviewers:      getReviewers(ctx, task),
		TaskName:       task.Name,
		TemplateData:   template.FromContext(ctx),
		Title:          prTitle,
	}

	// Always create if branch of task contains changes compared to default branch and no PR has been created yet.
	// Create if branch of task contains changes and the PR has been merged or closed before.
	if (hasChangesInRemoteDefaultBranch && prID == nil) || (hasChanges && (prID == nil || prID.State == host.PullRequestStateMerged || prID.State == host.PullRequestStateClosed)) {
		logger.Info("Creating pull request")
		if !dryRun {
			prID, err = repo.CreatePullRequest(branchName, prData)
			if err != nil {
				return ResultUnknown, prID, fmt.Errorf("failed to create pull request: %w", err)
			}
		}

		err = task.OnPrCreated(ctx)
		if err != nil {
			return ResultUnknown, prID, fmt.Errorf("pr created event failed: %w", err)
		}

		return ResultPrCreated, prID, nil
	}

	// Try to merge if auto-merge is enabled, no new changes have been detected and the pull request is open
	if task.AutoMerge && !hasChanges && prID != nil && prID.State == host.PullRequestStateOpen {
		success, err := repo.HasSuccessfulPullRequestBuild(prID)
		if err != nil {
			return ResultUnknown, prID, fmt.Errorf("check for successful pull request build failed: %w", err)
		}

		if !success {
			return ResultChecksFailed, prID, nil
		}

		if !canMergeAfter(prID.CreatedAt, task.CalcAutoMergeAfter()) {
			logger.Info("Too early to merge pull request")
			return ResultAutoMergeTooEarly, prID, nil
		}

		canMerge, err := repo.CanMergePullRequest(prID)
		if err != nil {
			return ResultUnknown, prID, fmt.Errorf("check if pull request can be merged: %w", err)
		}

		if !canMerge {
			logger.Warn("Cannot merge pull request")
			return ResultConflict, prID, nil
		}

		logger.Info("Merging pull request")
		if !dryRun {
			err := repo.MergePullRequest(!task.KeepBranchAfterMerge, prID)
			if err != nil {
				return ResultUnknown, prID, fmt.Errorf("failed to merge pull request: %w", err)
			}
		}

		err = task.OnPrMerged(ctx)
		if err != nil {
			return ResultUnknown, prID, fmt.Errorf("pr merged event failed: %w", err)
		}

		prID.State = host.PullRequestStateMerged
		return ResultPrMerged, prID, nil
	}

	if prID != nil && prID.State == host.PullRequestStateOpen {
		logger.Debug("Updating pull request")
		if !dryRun {
			err := repo.UpdatePullRequest(prData, prID)
			if err != nil {
				return ResultUnknown, prID, fmt.Errorf("failed to update pull request: %w", err)
			}
		}

		return ResultPrOpen, prID, nil
	}

	return ResultNoChanges, prID, nil
}

// findPullRequest looks up the pull request.
// It defers to the repo if the cache doesn't contain the pull request.
func (p *Processor) findPullRequest(branchName string, repo host.Repository) (*host.PullRequest, error) {
	if p.PullRequestCache != nil {
		if pr := p.PullRequestCache.Get(branchName, repo.FullName()); pr != nil {
			return pr, nil
		}
	}

	return repo.FindPullRequest(branchName)
}

func needsRebaseByUser(repo host.Repository, pr *host.PullRequest) bool {
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

// getAssignees merges static assignees from a task with dynamic assignees from run data.
func getAssignees(ctx context.Context, t *task.Task) []string {
	return mergeUsers(ctx, sbcontext.RunDataKeyAssignees, t.Assignees)
}

// getReviewers merges static reviewers from a task with dynamic reviewers from run data.
func getReviewers(ctx context.Context, t *task.Task) []string {
	return mergeUsers(ctx, sbcontext.RunDataKeyReviewers, t.Reviewers)
}

func mergeUsers(ctx context.Context, key string, static []string) []string {
	runData := sbcontext.RunData(ctx)
	dataRaw, ok := runData[key]
	if !ok {
		return static
	}

	users := strings.Split(dataRaw, ",")
	users = append(users, static...)
	slices.Sort(users)
	return slices.Compact(users)
}

// IsPrOpen returns true for all types of results which indicate
// that a pull request is still open.
func IsPrOpen(result Result) bool {
	// A bit verbose but better than a single, long case.
	switch result {
	case ResultPrCreated:
		return true
	case ResultPrOpen:
		return true
	case ResultAutoMergeTooEarly:
		return true
	case ResultBranchModified:
		return true
	case ResultChecksFailed:
		return true
	case ResultConflict:
		return true
	default:
		return false
	}
}
