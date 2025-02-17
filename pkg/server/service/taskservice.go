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

// ListRecentTaskResultsByTaskOptions
type ListRecentTaskResultsByTaskOptions struct {
	TaskName string
	Status   []db.TaskResultStatus
}

// ListRecentTaskResultsByTask returns a list of the recent result for each repository filtered by task and, optionally, status.
// It supports pagination through listOpts.
func (ts *TaskService) ListRecentTaskResultsByTask(opts ListRecentTaskResultsByTaskOptions, listOpts *ListOptions) ([]db.TaskResult, error) {
	_, err := ts.GetTask(opts.TaskName)
	if err != nil {
		return nil, err
	}

	// This sub-query returns the latest entry for each repository
	subQ := ts.db.
		Table("task_results").
		Select("MAX(task_results.created_at), task_results.error, task_results.id, task_results.repository_name, task_results.result, task_results.run_id, task_results.status, task_results.pull_request_url").
		Joins("INNER JOIN runs ON task_results.run_id = runs.id").
		Where("runs.task_name = ?", opts.TaskName).
		Group("task_results.repository_name").
		Order("task_results.created_at DESC")
	// Prepare query. Used by both other queries that return data and count rows.
	baseQ := ts.db.Table("(?)", subQ)
	q := baseQ.
		Offset(listOpts.Offset()).
		Limit(listOpts.Limit)
	if len(opts.Status) > 0 {
		q = q.Where("status IN ?", opts.Status)
	}

	var taskResults []db.TaskResult
	listResult := q.Find(&taskResults)
	if listResult.Error != nil {
		return nil, listResult.Error
	}

	countQuery := ts.db.Table("(?)", subQ)
	if len(opts.Status) > 0 {
		countQuery = countQuery.Where("status IN ?", opts.Status)
	}

	var count int64
	countResult := countQuery.Count(&count)
	if countResult.Error != nil {
		return nil, countResult.Error
	}

	listOpts.SetTotalItems(int(count))
	return taskResults, nil
}

func encodeBase64(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read content of task file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}
