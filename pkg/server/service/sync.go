package service

import (
	"errors"
	"fmt"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
	"gorm.io/gorm"
)

type Sync struct {
	clock         clock.Clock
	db            *gorm.DB
	taskService   *TaskService
	workerService *WorkerService
}

func NewSync(clock clock.Clock, db *gorm.DB, taskService *TaskService, workerService *WorkerService) *Sync {
	return &Sync{
		clock:         clock,
		db:            db,
		taskService:   taskService,
		workerService: workerService,
	}
}

func (s *Sync) SyncTasksInDatabase() error {
	for _, t := range s.taskService.ListTasks() {
		var taskDB db.Task
		tx := s.db.Where("name = ?", t.Name).First(&taskDB)
		if tx.Error != nil {
			if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return tx.Error
			}

			log.Log().Debugf("Creating task %s in DB", t.Name)
			taskDB.Hash = t.Checksum()
			taskDB.Name = t.Name
			err := s.db.Transaction(func(tx *gorm.DB) error {
				if err := s.db.Save(&taskDB).Error; err != nil {
					return fmt.Errorf("create task '%s' in db: %w", t.Name, err)
				}

				_, err := s.workerService.ScheduleRun(db.RunReasonNew, nil, s.clock.Now(), t.Name, map[string]string{}, nil)
				if handleScheduleRunError(err) != nil {
					return fmt.Errorf("schedule run for new task '%s' in db: %w", t.Name, err)
				}

				return nil
			})
			if err != nil {
				return err
			}
		} else {
			if taskDB.Hash != t.Checksum() {
				taskDB.Hash = t.Checksum()
				log.Log().Debugf("Updating task %s in DB", taskDB.Name)
				_, err := s.workerService.ScheduleRun(db.RunReasonChanged, nil, s.clock.Now(), t.Name, map[string]string{}, nil)
				if handleScheduleRunError(err) != nil {
					return fmt.Errorf("schedule run for updated task '%s' in db: %w", t.Name, err)
				}

				if err := s.db.Save(&taskDB).Error; err != nil {
					return fmt.Errorf("update task '%s' in db: %w", t.Name, err)
				}
			}
		}
	}

	return nil
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
