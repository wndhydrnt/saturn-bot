package service

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/adhocore/gronx"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrNoRun = errors.New("no next run")
)

type ErrorMissingInput struct {
	InputName string
	TaskName  string
}

func (e ErrorMissingInput) Error() string {
	return fmt.Sprintf("missing required input %s for task %s", e.InputName, e.TaskName)
}

type WorkerService struct {
	clock                 clock.Clock
	db                    *gorm.DB
	inShutdown            atomic.Bool
	shutdownCheckInterval time.Duration
	taskService           *TaskService
}

func NewWorkerService(clock clock.Clock, db *gorm.DB, taskService *TaskService, shutdownCheckInterval time.Duration) *WorkerService {
	return &WorkerService{
		clock:                 clock,
		db:                    db,
		shutdownCheckInterval: shutdownCheckInterval,
		taskService:           taskService,
	}
}

// DeletePendingRuns deletes all runs of the task identified by taskName that are pending.
func (ws *WorkerService) DeletePendingRuns(taskName string, tx *gorm.DB) error {
	if tx == nil {
		tx = ws.db
	}

	var runsOfTask []db.Run
	result := tx.Where("task_name = ?", taskName).
		Where("status = ?", db.RunStatusPending).
		Find(&runsOfTask)
	if result.Error != nil {
		return fmt.Errorf("find pending runs: %w", result.Error)
	}

	for _, run := range runsOfTask {
		result := tx.Delete(&run)
		if result.Error != nil {
			return fmt.Errorf("delete pending run %d: %w", run.ID, result.Error)
		}
	}

	return nil
}

func (ws *WorkerService) ScheduleRun(
	reason db.RunReason,
	repositoryNames []string,
	scheduleAfter time.Time,
	taskName string,
	runData map[string]string,
	tx *gorm.DB,
) (uint, error) {
	t, err := ws.taskService.GetTask(taskName)
	if err != nil {
		return 0, err
	}

	errs := task.ValidateInputs(runData, t)
	if len(errs) > 0 {
		return 0, sberror.NewInputError(errs, taskName)
	}

	var runDB db.Run
	if tx == nil {
		tx = ws.db
	}
	query := tx.
		Where("task_name = ?", taskName).
		Where("status = ?", db.RunStatusPending).
		Where("reason = ?", reason)

	repositoryNameList := db.StringList(repositoryNames)
	if len(repositoryNameList) == 0 {
		query = query.Where("repository_names is null")
	} else {
		query = query.Where("repository_names = ?", repositoryNameList)
	}

	runDataDb := db.StringMap(runData)
	if len(runDataDb) == 0 {
		query = query.Where("run_data is null")
	} else {
		query = query.Where("run_data = ?", runDataDb)
	}

	result := query.First(&runDB)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if len(repositoryNames) > 0 {
			log.Log().Debugf("Scheduling new run of task %s for repositories %v", taskName, repositoryNames)
		} else {
			log.Log().Debugf("Scheduling new run of task %s for all repositories", taskName)
		}

		run := db.Run{
			Reason:          reason,
			RepositoryNames: repositoryNameList,
			ScheduleAfter:   scheduleAfter,
			Status:          db.RunStatusPending,
			TaskName:        taskName,
			RunData:         runDataDb,
		}

		if err := tx.Save(&run).Error; err != nil {
			return 0, fmt.Errorf("schedule run: %w", err)
		}

		return run.ID, nil
	}

	if result.Error != nil {
		return 0, fmt.Errorf("read next scheduled run: %w", result.Error)
	}

	// Check for equality to prevent runs based on a cron schedule
	// to be scheduled twice.
	if !runDB.ScheduleAfter.Equal(scheduleAfter) {
		runDB.Reason = reason
		runDB.ScheduleAfter = scheduleAfter
		if err := tx.Save(&runDB).Error; err != nil {
			return 0, fmt.Errorf("update scheduleAfter of run: %w", err)
		}
	}

	return runDB.ID, nil
}

func (ws *WorkerService) findTask(name string) (*task.Task, error) {
	t, err := ws.taskService.GetTask(name)
	return t, err
}

func (ws *WorkerService) NextRun() (db.Run, *task.Task, error) {
	var run db.Run
	if ws.shuttingDown() {
		return run, nil, ErrNoRun
	}

	tx := ws.db.
		Where("status = ?", db.RunStatusPending).
		Order("schedule_after asc").
		First(&run)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return run, nil, ErrNoRun
		}

		log.Log().Errorw("Failed to get next run", zap.Error(tx.Error))
		return run, nil, tx.Error
	}

	if run.ScheduleAfter.After(ws.clock.Now()) {
		return run, nil, ErrNoRun
	}

	run.StartedAt = ptr.To(ws.clock.Now())
	run.Status = db.RunStatusRunning
	if err := ws.db.Save(&run).Error; err != nil {
		log.Log().Errorw("Update next run", zap.Error(err))
		return run, nil, err
	}

	task, _ := ws.findTask(run.TaskName)
	if task == nil {
		if err := ws.db.Delete(&run).Error; err != nil {
			log.Log().Error("Delete run of unknown task", zap.Error(tx.Error))
		}

		return run, nil, ErrNoRun
	}

	return run, task, nil
}

func (ws *WorkerService) ReportRun(req openapi.ReportWorkV1Request) error {
	log.Log().Debugf("Report of run %d", req.RunID)
	var runCurrent db.Run
	tx := ws.db.First(&runCurrent, req.RunID)
	if tx.Error != nil {
		log.Log().Errorw("Failed to retrieve run by ID", zap.Error(tx.Error))
		return tx.Error
	}

	task, err := ws.taskService.GetTask(req.Task.Name)
	if err != nil {
		return err
	}

	err = ws.db.Transaction(func(tx *gorm.DB) error {
		runCurrent.FinishedAt = ptr.To(ws.clock.Now())
		if req.Error == nil {
			runCurrent.Status = db.RunStatusFinished
		} else {
			runCurrent.Error = req.Error
			runCurrent.Status = db.RunStatusFailed
		}

		if err := tx.Save(&runCurrent).Error; err != nil {
			return err
		}

		prIsOpen := false
		for _, taskResult := range req.TaskResults {
			if !prIsOpen && processor.IsPrOpen(processor.Result(taskResult.Result)) {
				prIsOpen = true
			}

			var resultDb db.TaskResult
			resultDbStmt := tx.Select("task_results.*").
				Joins("INNER JOIN runs ON task_results.run_id = runs.id").
				Where("task_results.repository_name = ?", taskResult.RepositoryName).
				Where("runs.task_name = ?", runCurrent.TaskName).
				Order("created_at DESC").
				First(&resultDb)
			if resultDbStmt.Error != nil && !errors.Is(resultDbStmt.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("read most recent task result: %w", resultDbStmt.Error)
			}

			status := mapTaskResultStateFromApiToDb(taskResult.State)
			if resultDb.Status == status && isSamePullRequestUrl(taskResult.PullRequestUrl, resultDb.PullRequestUrl) {
				// No change in status. Skip this result.
				continue
			}

			result := db.TaskResult{
				RepositoryName: taskResult.RepositoryName,
				Result:         taskResult.Result,
				RunID:          runCurrent.ID,
				Status:         status,
			}
			if taskResult.Error != nil {
				result.Error = taskResult.Error
			}

			if taskResult.PullRequestUrl != nil {
				result.PullRequestUrl = taskResult.PullRequestUrl
			}

			if err := tx.Save(&result).Error; err != nil {
				return err
			}
		}

		next := calcNextScheduleTime(runCurrent, ws.clock.Now(), task, prIsOpen)
		if next != nil {
			_, err := ws.ScheduleRun(runCurrent.Reason, runCurrent.RepositoryNames, ptr.From(next), runCurrent.TaskName, runCurrent.RunData, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

type ListRunsOptions struct {
	Status   []db.RunStatus
	TaskName string
}

func (ws *WorkerService) ListRuns(opts ListRunsOptions, listOpts *ListOptions) ([]db.Run, error) {
	query := ws.db
	if len(opts.Status) > 0 {
		query = query.Where("status IN ?", opts.Status)
	}

	if opts.TaskName != "" {
		query = query.Where("task_name = ?", opts.TaskName)
	}

	var runs []db.Run
	result := query.
		Offset(listOpts.Offset()).
		Limit(listOpts.Limit).
		Order("schedule_after DESC").
		Find(&runs)

	if result.Error != nil {
		return nil, result.Error
	}

	var count int64
	queryCount := ws.db.Model(&db.Run{})
	if opts.Status != nil {
		queryCount = queryCount.Where("status IN ?", opts.Status)
	}

	if opts.TaskName != "" {
		queryCount = queryCount.Where("task_name = ?", opts.TaskName)
	}

	countResult := queryCount.Count(&count)
	if countResult.Error != nil {
		return nil, countResult.Error
	}

	listOpts.SetTotalItems(int(count))
	return runs, result.Error
}

// DeleteRun deletes a run by its id.
func (ws *WorkerService) DeleteRun(id int) error {
	run, err := ws.GetRun(id)
	if err != nil {
		return err
	}

	if run.Reason != db.RunReasonManual {
		return sberror.NewRunCannotDeleteError()
	}

	if run.Status != db.RunStatusPending {
		return sberror.NewRunCannotDeleteError()
	}

	result := ws.db.Delete(&run)
	if result.Error != nil {
		return fmt.Errorf("delete run '%d': %w", id, result.Error)
	}

	return nil
}

// GetRun returns a [db.Run] identified by id.
//
// It returns an error if no run is found.
func (ws *WorkerService) GetRun(id int) (db.Run, error) {
	var run db.Run
	result := ws.db.Where("id = ?", id).First(&run)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return run, sberror.NewRunNotFoundError(id)
		}

		return run, fmt.Errorf("get run by id: %w", result.Error)
	}

	return run, nil
}

type ListTaskResultsOptions struct {
	RepositoryName string
	RunId          int
	Status         []db.TaskResultStatus
}

// ListTaskResults returns a list of [db.TaskResult].
// Items are ordered by the field CreatedAt in descending order.
//
// Allows filtering via [ListTaskResultsOptions] and pagination via [ListOptions].
func (ws *WorkerService) ListTaskResults(opts ListTaskResultsOptions, listOpts *ListOptions) ([]db.TaskResult, error) {
	query := ws.db
	if opts.RepositoryName != "" {
		query = query.Where("repository_name = ?", opts.RepositoryName)
	}

	if opts.RunId != 0 {
		query = query.Where("run_id = ?", opts.RunId)
	}

	if len(opts.Status) > 0 {
		query = query.Where("status IN ?", opts.Status)
	}

	var taskResults []db.TaskResult
	result := query.
		Offset(listOpts.Offset()).
		Limit(listOpts.Limit).
		Order("created_at DESC").
		Find(&taskResults)
	if result.Error != nil {
		return nil, fmt.Errorf("list task results: %w", result.Error)
	}

	var count int64
	queryCount := ws.db.Model(&db.TaskResult{})
	if opts.RepositoryName != "" {
		queryCount = queryCount.Where("repository_name = ?", opts.RepositoryName)
	}

	if opts.RunId != 0 {
		queryCount = queryCount.Where("run_id = ?", opts.RunId)
	}

	if len(opts.Status) > 0 {
		queryCount = queryCount.Where("status IN ?", opts.Status)
	}

	countResult := queryCount.Count(&count)
	if countResult.Error != nil {
		return nil, countResult.Error
	}

	listOpts.SetTotalItems(int(count))
	return taskResults, nil
}

func (ws *WorkerService) Shutdown(ctx context.Context) error {
	ws.inShutdown.Store(true)

	runCount, err := ws.countActiveRuns()
	if err != nil {
		return fmt.Errorf("list running runs on shutdown: %w", err)
	}

	if runCount == 0 {
		return nil
	}

	checkInterval := ws.shutdownCheckInterval
	if ws.shutdownCheckInterval == 0 {
		checkInterval = time.Second
	}

	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ticker.C:
			runCount, err := ws.countActiveRuns()
			if err != nil {
				ticker.Stop()
				return fmt.Errorf("list running runs on shutdown: %w", err)
			}

			if runCount == 0 {
				ticker.Stop()
				return nil
			}

			log.Log().Warnf("Waiting for %d runs to finish before shutdown", runCount)

		case <-ctx.Done():
			markErr := ws.markActiveRunsAsFailed("Run failed to report before shutdown")
			return fmt.Errorf("shutdown worker service: %w", errors.Join(ctx.Err(), markErr))
		}
	}
}

func (ws *WorkerService) shuttingDown() bool {
	return ws.inShutdown.Load()
}

func (ws *WorkerService) countActiveRuns() (int64, error) {
	var count int64
	result := ws.db.
		Model(&db.Run{}).
		Where("status = ?", db.RunStatusRunning).
		Count(&count)
	return count, result.Error
}

func (ws *WorkerService) markActiveRunsAsFailed(errMsg string) error {
	var runs []db.Run
	findResult := ws.db.
		Where("status = ?", db.RunStatusRunning).
		Find(&runs)
	if findResult.Error != nil {
		return fmt.Errorf("find runs to mark as failed on shutdown: %w", findResult.Error)
	}

	for _, run := range runs {
		t, err := ws.findTask(run.TaskName)
		if err != nil {
			return fmt.Errorf("find task to mark as failed on shutdown: %w", err)
		}

		err = ws.ReportRun(openapi.ReportWorkV1Request{
			Error: ptr.To(errMsg),
			RunID: int(run.ID),
			Task: openapi.WorkTaskV1{
				Hash: t.Checksum(),
				Name: t.Name,
			},
		})
		if err != nil {
			return fmt.Errorf("mark run %d of task %s as failed on shutdown: %w", run.ID, t.Name, err)
		}
	}

	return nil
}

func calcNextScheduleTime(run db.Run, now time.Time, t *task.Task, isOpen bool) *time.Time {
	// If task defines a cron trigger, always adhere to the cron schedule.
	if run.Reason == db.RunReasonCron {
		return calcNextCronTime(now, t)
	}

	// If at least one PR is open, schedule a new run to keep getting status updates.
	if isOpen {
		log.Log().Info("Scheduling new run because pull requests are open")
		if t.AutoMerge {
			return ptr.To(run.ScheduleAfter.Add(1 * time.Hour))
		} else {
			return ptr.To(run.ScheduleAfter.Add(24 * time.Hour))
		}
	}

	return nil
}

func calcNextCronTime(now time.Time, t *task.Task) *time.Time {
	if t == nil {
		return nil
	}

	if t.Trigger == nil {
		return nil
	}

	if t.Trigger.Cron == nil {
		return nil
	}

	nextTick, _ := gronx.NextTickAfter(ptr.From(t.Trigger.Cron), now, true)
	return ptr.To(nextTick)
}

func mapTaskResultStateFromApiToDb(state openapi.TaskResultStateV1) db.TaskResultStatus {
	return db.TaskResultStatus(state)
}

func isSamePullRequestUrl(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil && b != nil {
		return false
	}

	if a != nil && b == nil {
		return false
	}

	return ptr.From(a) == ptr.From(b)
}
