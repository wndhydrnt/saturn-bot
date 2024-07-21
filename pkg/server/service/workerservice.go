package service

import (
	"errors"
	"log/slog"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/task"
	"gorm.io/gorm"
)

var ErrNoRun = errors.New("no next run")

type WorkerService struct {
	db    *gorm.DB
	tasks []task.Task
}

func NewWorkerService(db *gorm.DB, tasks []task.Task) *WorkerService {
	return &WorkerService{db: db, tasks: tasks}
}

func (ws *WorkerService) findTask(name string) *task.Task {
	for _, t := range ws.tasks {
		if t.TaskName == name {
			return &t
		}
	}

	return nil
}

func (ws *WorkerService) NextRun() (db.Run, task.Task, error) {
	var run db.Run
	var runTask task.Task
	tx := ws.db.
		Where("status = ?", db.RunStatusPending).
		Order("schedule_after desc").
		First(&run)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return run, runTask, ErrNoRun
		}

		slog.Error("Failed to get next run", "error", tx.Error)
		return run, runTask, tx.Error
	}

	run.Status = db.RunStatusExecuting
	if err := ws.db.Save(&run).Error; err != nil {
		slog.Error("Update next run", "error", err)
		return run, runTask, err
	}

	task := ws.findTask(run.TaskName)
	if task == nil {
		if err := ws.db.Delete(&run).Error; err != nil {
			slog.Error("Delete run of unknown task", "error", tx.Error)
		}

		return run, runTask, ErrNoRun
	}

	return run, *task, nil
}

func (ws *WorkerService) ReportRun(req openapi.ReportWorkV1Request) error {
	var run db.Run
	tx := ws.db.First(&run, req.RunID)
	if tx.Error != nil {
		slog.Error("Failed to retrieve run by ID", "error", tx.Error)
		return tx.Error
	}

	err := ws.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		run.FinishedAt = &now
		run.Status = db.RunStatusFinished
		if err := tx.Save(&run).Error; err != nil {
			return err
		}

		for _, taskResult := range req.TaskResults {
			result := db.TaskResult{
				RepositoryName: taskResult.RepositoryName,
				Result:         uint(taskResult.Result),
				RunID:          run.ID,
				TaskName:       taskResult.TaskName,
			}
			if taskResult.Error != "" {
				result.Error = &taskResult.Error
			}

			if err := tx.Save(&result).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
