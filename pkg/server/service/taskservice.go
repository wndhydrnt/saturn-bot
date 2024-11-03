package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"gorm.io/gorm"
)

type TaskService struct {
	db    *gorm.DB
	tasks []schema.ReadResult
}

func NewTaskService(db *gorm.DB, tasks []schema.ReadResult) *TaskService {
	return &TaskService{db: db, tasks: tasks}
}

func (ts *TaskService) SyncDbTasks() error {
	for _, t := range ts.tasks {
		var taskDB db.Task
		tx := ts.db.Where("name = ?", t.Task.Name).First(&taskDB)
		if tx.Error != nil {
			if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return tx.Error
			}

			log.Log().Debugf("Creating task %s in DB", t.Task.Name)
			taskDB.Hash = t.Sha256
			taskDB.Name = t.Task.Name
			err := ts.db.Transaction(func(tx *gorm.DB) error {
				if err := ts.db.Save(&taskDB).Error; err != nil {
					return fmt.Errorf("create task '%s' in db: %w", t.Task.Name, err)
				}

				run := db.Run{
					Reason:        0,
					ScheduleAfter: time.Now(),
					Status:        db.RunStatusPending,
					TaskName:      t.Task.Name,
				}
				if err := ts.db.Save(&run).Error; err != nil {
					return fmt.Errorf("schedule run for new task '%s' in db: %w", t.Task.Name, err)
				}

				return nil
			})
			if err != nil {
				return err
			}
		} else {
			if taskDB.Hash != t.Sha256 {
				taskDB.Hash = t.Sha256
				log.Log().Debugf("Updating task %s in DB", taskDB.Name)
				run := db.Run{
					Reason:        0,
					ScheduleAfter: time.Now(),
					Status:        db.RunStatusPending,
					TaskName:      t.Task.Name,
				}
				if err := ts.db.Save(&run).Error; err != nil {
					return fmt.Errorf("schedule run for updated task '%s' in db: %w", t.Task.Name, err)
				}

				if err := ts.db.Save(&taskDB).Error; err != nil {
					return fmt.Errorf("update task '%s' in db: %w", t.Task.Name, err)
				}
			}
		}
	}

	return nil
}

func (ts *TaskService) GetTask(taskName string) (*schema.ReadResult, string) {
	for _, entry := range ts.tasks {
		if entry.Task.Name == taskName {
			content, err := encodeBase64(entry.Path)
			if err != nil {
				return nil, ""
			}

			return &entry, content
		}
	}

	return nil, ""
}

func (ts *TaskService) ListTasks() []schema.ReadResult {
	return ts.tasks
}

func encodeBase64(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read content of task file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}
