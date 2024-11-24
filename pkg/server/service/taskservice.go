package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"gorm.io/gorm"
)

type TaskService struct {
	clock        clock.Clock
	db           *gorm.DB
	taskRegistry *task.Registry
}

func NewTaskService(clock clock.Clock, db *gorm.DB, taskRegistry *task.Registry) *TaskService {
	return &TaskService{
		clock:        clock,
		db:           db,
		taskRegistry: taskRegistry,
	}
}

func (ts *TaskService) SyncDbTasks() error {
	for _, t := range ts.taskRegistry.GetTasks() {
		var taskDB db.Task
		tx := ts.db.Where("name = ?", t.Name).First(&taskDB)
		if tx.Error != nil {
			if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return tx.Error
			}

			log.Log().Debugf("Creating task %s in DB", t.Name)
			taskDB.Hash = t.Checksum()
			taskDB.Name = t.Name
			err := ts.db.Transaction(func(tx *gorm.DB) error {
				if err := ts.db.Save(&taskDB).Error; err != nil {
					return fmt.Errorf("create task '%s' in db: %w", t.Name, err)
				}

				run := db.Run{
					Reason:        db.RunReasonNew,
					ScheduleAfter: ts.clock.Now(),
					Status:        db.RunStatusPending,
					TaskName:      t.Name,
				}
				if err := ts.db.Save(&run).Error; err != nil {
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
				run := db.Run{
					Reason:        db.RunReasonChanged,
					ScheduleAfter: ts.clock.Now(),
					Status:        db.RunStatusPending,
					TaskName:      t.Name,
				}
				if err := ts.db.Save(&run).Error; err != nil {
					return fmt.Errorf("schedule run for updated task '%s' in db: %w", t.Name, err)
				}

				if err := ts.db.Save(&taskDB).Error; err != nil {
					return fmt.Errorf("update task '%s' in db: %w", t.Name, err)
				}
			}
		}
	}

	return nil
}

func (ts *TaskService) GetTask(taskName string) (*task.Task, string) {
	for _, entry := range ts.taskRegistry.GetTasks() {
		if entry.Name == taskName {
			content, err := encodeBase64(entry.Path())
			if err != nil {
				return nil, ""
			}

			return entry, content
		}
	}

	return nil, ""
}

func (ts *TaskService) ListTasks() []*task.Task {
	return ts.taskRegistry.GetTasks()
}

func encodeBase64(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read content of task file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}
