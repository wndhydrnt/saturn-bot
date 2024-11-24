package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrNoRun = errors.New("no next run")

type WorkerService struct {
	clock        clock.Clock
	db           *gorm.DB
	taskRegistry *task.Registry
}

func NewWorkerService(clock clock.Clock, db *gorm.DB, taskRegistry *task.Registry) *WorkerService {
	return &WorkerService{
		clock:        clock,
		db:           db,
		taskRegistry: taskRegistry,
	}
}

func (ws *WorkerService) ScheduleRun(
	reason db.RunReason,
	repositoryNames []string,
	scheduleAfter time.Time,
	taskName string,
	tx *gorm.DB,
) (uint, error) {
	var runDB db.Run
	if tx == nil {
		tx = ws.db
	}
	query := tx.
		Where("task_name = ?", taskName).
		Where("status = ?", db.RunStatusPending)

	repositoryNameList := db.StringList(repositoryNames)
	if len(repositoryNameList) == 0 {
		query = query.Where("repository_names is null")
	} else {
		query = query.Where("repository_names = ?", repositoryNameList)
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
		}

		if err := tx.Save(&run).Error; err != nil {
			return 0, fmt.Errorf("schedule run: %w", err)
		}

		return run.ID, nil
	}

	if result.Error != nil {
		return 0, fmt.Errorf("read next scheduled run: %w", tx.Error)
	}

	if runDB.ScheduleAfter.After(scheduleAfter) {
		runDB.Reason = reason
		runDB.ScheduleAfter = scheduleAfter
		if err := tx.Save(&runDB).Error; err != nil {
			return 0, fmt.Errorf("update scheduleAfter of run: %w", err)
		}
	}

	return runDB.ID, nil
}

func (ws *WorkerService) findTask(name string) *task.Task {
	for _, t := range ws.taskRegistry.GetTasks() {
		if t.Task.Name == name {
			return t
		}
	}

	return nil
}

func (ws *WorkerService) NextRun() (db.Run, *task.Task, error) {
	var run db.Run
	tx := ws.db.
		Where("status = ?", db.RunStatusPending).
		Order("schedule_after desc").
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

	task := ws.findTask(run.TaskName)
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

	err := ws.db.Transaction(func(tx *gorm.DB) error {
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

		next := 1 * time.Hour
		for _, taskResult := range req.TaskResults {
			result := db.TaskResult{
				RepositoryName: taskResult.RepositoryName,
				Result:         uint(taskResult.Result), // #nosec G115 -- no info by gosec on how to fix this
				RunID:          runCurrent.ID,
				TaskName:       taskResult.TaskName,
			}
			if taskResult.Error != nil {
				result.Error = taskResult.Error
			}

			if err := tx.Save(&result).Error; err != nil {
				return err
			}

			procResult := processor.Result(taskResult.Result)
			nextNew := nextSchedule(procResult)
			if nextNew < next {
				next = nextNew
			}
		}

		_, err := ws.ScheduleRun(db.RunReasonNext, runCurrent.RepositoryNames, ws.clock.Now().Add(next), runCurrent.TaskName, tx)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

type ListRunsOptions struct {
	TaskName string
}

func (ws *WorkerService) ListRuns(opts ListRunsOptions, listOpts ListOptions) ([]db.Run, int64, error) {
	var runs []db.Run
	query := ws.db
	if opts.TaskName != "" {
		query = query.Where("task_name = ?", opts.TaskName)
	}

	result := query.
		Offset(listOpts.Offset()).
		Limit(listOpts.Limit).
		Order("schedule_after DESC").
		Find(&runs)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	var count int64
	queryCount := ws.db.Model(&db.Run{})
	if opts.TaskName != "" {
		queryCount = queryCount.Where("task_name = ?", opts.TaskName)
	}

	countResult := queryCount.Count(&count)
	if countResult.Error != nil {
		return nil, 0, countResult.Error
	}

	return runs, count, result.Error
}

func nextSchedule(r processor.Result) time.Duration {
	switch r {
	case processor.ResultUnknown:
		return 15 * time.Minute
	case processor.ResultAutoMergeTooEarly:
		return 60 * time.Minute
	case processor.ResultPrCreated:
		return 15 * time.Minute
	case processor.ResultPrOpen:
		return 15 * time.Minute
	case processor.ResultNoChanges:
		return 60 * time.Minute
	default:
		return 60 * time.Minute
	}
}
