package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/server/task"
)

type TaskService struct {
	tasks []task.Task
}

func NewTaskService(tasks []task.Task) *TaskService {
	return &TaskService{tasks: tasks}
}

func (ts *TaskService) GetTaskV1(_ context.Context, taskName string) (openapi.ImplResponse, error) {
	for _, entry := range ts.tasks {
		if entry.TaskName == taskName {
			content, err := encodeBase64(entry.TaskPath)
			if err != nil {
				return openapi.Response(http.StatusInternalServerError, serverError), nil
			}

			body := openapi.GetTaskV1200Response{
				Name:    entry.TaskName,
				Hash:    entry.Hash,
				Content: content,
			}
			return openapi.Response(http.StatusOK, body), nil
		}
	}

	return openapi.Response(http.StatusNotFound, openapi.Error{Error: "Not Found", Message: "Task unknown"}), nil
}

func (ts *TaskService) ListTasksV1(_ context.Context) (openapi.ImplResponse, error) {
	body := openapi.ListTasksV1200Response{
		Tasks: []string{},
	}
	for _, entry := range ts.tasks {
		body.Tasks = append(body.Tasks, entry.TaskName)
	}

	return openapi.Response(http.StatusOK, body), nil
}

func encodeBase64(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read content of task file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}
