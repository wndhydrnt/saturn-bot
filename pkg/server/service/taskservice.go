package service

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	sberror "github.com/wndhydrnt/saturn-bot/pkg/server/error"
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

// GetTask returns the task identified by taskName.
//
// It returns an error if no task matches taskName.
func (ts *TaskService) GetTask(taskName string) (*task.Task, error) {
	for _, entry := range ts.taskRegistry.GetTasks() {
		if entry.Name == taskName {
			return entry, nil
		}
	}

	return nil, sberror.NewTaskNotFoundError(taskName)
}

// EncodeTaskBase64 returns the content of the task identified by taskName.
//
// It returns an error if no task matches taskName.
func (ts *TaskService) EncodeTaskBase64(taskName string) (string, error) {
	task, err := ts.GetTask(taskName)
	if err != nil {
		return "", err
	}

	return encodeBase64(task.Path())
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
