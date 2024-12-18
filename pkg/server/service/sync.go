package service

import (
	"errors"
	"fmt"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"gorm.io/gorm"
)

// Sync synchronizes tasks on startup.
type Sync struct {
	clock         clock.Clock
	db            *gorm.DB
	taskService   *TaskService
	workerService *WorkerService
}

// NewSync returns a new Sync.
func NewSync(clock clock.Clock, db *gorm.DB, taskService *TaskService, workerService *WorkerService) *Sync {
	return &Sync{
		clock:         clock,
		db:            db,
		taskService:   taskService,
		workerService: workerService,
	}
}

// SyncTasksInDatabase checks if any tasks have changed since that last start of the server.
//
// It schedules a new run of a task if the task has changed on disk.
// If a task has been deleted on disk, it flags the task as inactive in the database.
func (s *Sync) SyncTasksInDatabase() error {
	var errs []error
	var knownTasksNames []string
	for _, t := range s.taskService.ListTasks() {
		var taskDB db.Task
		tx := s.db.Where("name = ?", t.Name).First(&taskDB)
		if tx.Error != nil {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				knownTasksNames = append(knownTasksNames, t.Name)
				if err := s.createTask(t); err != nil {
					errs = append(errs, err)
				}
			} else {
				errs = append(errs, tx.Error)
			}
		} else {
			knownTasksNames = append(knownTasksNames, t.Name)
			if err := s.updateTask(t, taskDB); err != nil {
				errs = append(errs, err)
			}
		}
	}

	var inactiveTasks []db.Task
	tx := s.db.Not(map[string]any{"name": knownTasksNames}).Find(&inactiveTasks)
	if tx.Error == nil {
		for _, inactiveTask := range inactiveTasks {
			err := s.deactivateTask(inactiveTask)
			if err != nil {
				errs = append(errs, fmt.Errorf("find inactive tasks: %w", tx.Error))
			}
		}
	} else {
		errs = append(errs, fmt.Errorf("find inactive tasks: %w", tx.Error))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *Sync) createTask(t *task.Task) error {
	log.Log().Debugf("Creating task %s in DB", t.Name)
	taskDB := db.Task{
		Active: true,
		Hash:   t.Checksum(),
		Name:   t.Name,
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.db.Save(&taskDB).Error; err != nil {
			return fmt.Errorf("create task '%s' in db: %w", t.Name, err)
		}

		now := s.clock.Now()
		scheduleAfter := now
		cronTime := calcNextCronTime(now, t)
		if cronTime != nil {
			scheduleAfter = ptr.From(cronTime)
		}

		_, err := s.workerService.ScheduleRun(db.RunReasonNew, nil, scheduleAfter, t.Name, map[string]string{}, tx)
		if handleScheduleRunError(err) != nil {
			return fmt.Errorf("schedule run for new task '%s' in db: %w", t.Name, err)
		}

		return nil
	})
}

func (s *Sync) deactivateTask(t db.Task) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		t.Active = false
		saveResult := s.db.Save(t)
		if saveResult.Error != nil {
			return fmt.Errorf("deactivate task %s: %w", t.Name, saveResult.Error)
		}

		return s.workerService.DeletePendingRuns(t.Name, tx)
	})
}

func (s *Sync) updateTask(t *task.Task, taskDB db.Task) error {
	if taskDB.Hash == t.Checksum() {
		return nil
	}

	log.Log().Debugf("Updating task %s in DB", taskDB.Name)
	return s.db.Transaction(func(tx *gorm.DB) error {
		now := s.clock.Now()
		scheduleAfter := now
		cronTime := calcNextCronTime(now, t)
		if cronTime != nil {
			scheduleAfter = ptr.From(cronTime)
		}

		_, err := s.workerService.ScheduleRun(db.RunReasonChanged, nil, scheduleAfter, t.Name, map[string]string{}, tx)
		if handleScheduleRunError(err) != nil {
			return fmt.Errorf("schedule run for updated task '%s' in db: %w", t.Name, err)
		}

		if err := tx.Save(&taskDB).Error; err != nil {
			return fmt.Errorf("update task '%s' in db: %w", t.Name, err)
		}

		return nil
	})
}

func handleScheduleRunError(err error) error {
	if err == nil {
		return nil
	}

	var clientErr sberror.Client
	if errors.As(err, &clientErr) {
		if clientErr.ErrorID() == sberror.ClientIDInputMissing {
			log.Log().Infof("Not scheduling new run of task: %s", clientErr.Error())
			return nil
		}
	}

	return err
}