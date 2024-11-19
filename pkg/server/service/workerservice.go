package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrNoRun = errors.New("no next run")

type WorkerService struct {
	db    *gorm.DB
	tasks []schema.ReadResult
}

func NewWorkerService(db *gorm.DB, tasks []schema.ReadResult) *WorkerService {
	return &WorkerService{db: db, tasks: tasks}
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
		runDB.ScheduleAfter = scheduleAfter
		if err := tx.Save(&runDB).Error; err != nil {
			return 0, fmt.Errorf("update scheduleAfter of run: %w", err)
		}
	}

	return runDB.ID, nil
}

func (ws *WorkerService) findTask(name string) *schema.ReadResult {
	for _, t := range ws.tasks {
		if t.Task.Name == name {
			return &t
		}
	}

	return nil
}

func (ws *WorkerService) NextRun() (db.Run, schema.ReadResult, error) {
	var run db.Run
	var runTask schema.ReadResult
	tx := ws.db.
		Where("status = ?", db.RunStatusPending).
		Order("schedule_after desc").
		First(&run)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return run, runTask, ErrNoRun
		}

		log.Log().Errorw("Failed to get next run", zap.Error(tx.Error))
		return run, runTask, tx.Error
	}

	if run.ScheduleAfter.After(time.Now()) {
		return run, runTask, ErrNoRun
	}

	run.Status = db.RunStatusRunning
	if err := ws.db.Save(&run).Error; err != nil {
		log.Log().Errorw("Update next run", zap.Error(err))
		return run, runTask, err
	}

	task := ws.findTask(run.TaskName)
	if task == nil {
		if err := ws.db.Delete(&run).Error; err != nil {
			log.Log().Error("Delete run of unknown task", zap.Error(tx.Error))
		}

		return run, runTask, ErrNoRun
	}

	return run, *task, nil
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
		runCurrent.FinishedAt = ptr.To(time.Now())
		if req.Error == "" {
			runCurrent.Status = db.RunStatusFinished
		} else {
			runCurrent.Error = ptr.To(req.Error)
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
			if taskResult.Error != "" {
				result.Error = &taskResult.Error
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

		_, err := ws.ScheduleRun(db.RunReasonNext, runCurrent.RepositoryNames, time.Now().Add(next), runCurrent.TaskName, tx)
		if err != nil {
			return err
		}

		return nil
	})

	return err
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
