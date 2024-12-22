package service

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/server/db"
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

type ListTaskResultsOptions struct {
	Result   []int
	RunId    int
	TaskName string
}

func (ts *TaskService) ListTaskResults(opts ListTaskResultsOptions, listOpts *ListOptions) ([]db.TaskResult, error) {
	query := ts.db
	if opts.RunId != 0 {
		query = query.Where("run_id = ?", opts.RunId)
	}

	if opts.TaskName != "" {
		query = query.Where("task_name = ?", opts.TaskName)
	}

	if len(opts.Result) > 0 {
		query = query.Where("result IN ?", opts.Result)
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

	return taskResults, nil
}
