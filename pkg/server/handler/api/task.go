package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/server/handler/api/openapi"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

type taskEntry struct {
	hash     []byte
	taskName string
	taskPath string
}

type TaskService struct {
	tasks []taskEntry
}

func NewTaskService(taskPaths []string) (*TaskService, error) {
	var entries []taskEntry
	for _, taskPath := range taskPaths {
		tasks, checksum, err := schema.Read(taskPath)
		if err != nil {
			return nil, err
		}

		for _, t := range tasks {
			entries = append(entries, taskEntry{
				hash:     checksum.Sum(nil),
				taskName: t.Name,
				taskPath: taskPath,
			})
		}
	}

	return &TaskService{tasks: entries}, nil
}

func (ts *TaskService) GetTaskV1(_ context.Context, taskName string) (openapi.ImplResponse, error) {
	for _, entry := range ts.tasks {
		if entry.taskName == taskName {
			content, err := encodeBase64(entry.taskPath)
			if err != nil {
				return openapi.Response(http.StatusInternalServerError, serverError), nil
			}

			body := openapi.GetTaskV1200Response{
				Name:    entry.taskName,
				Hash:    fmt.Sprintf("%x", entry.hash),
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
		body.Tasks = append(body.Tasks, entry.taskName)
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
