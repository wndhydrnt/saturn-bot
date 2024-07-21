package task

import (
	"fmt"

	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

type Task struct {
	Hash     string
	TaskName string
	TaskPath string
}

func Load(taskPaths []string) ([]Task, error) {
	var entries []Task
	for _, taskPath := range taskPaths {
		tasks, h, err := schema.Read(taskPath)
		if err != nil {
			return nil, err
		}

		for _, t := range tasks {
			entries = append(entries, Task{
				Hash:     fmt.Sprintf("%x", h.Sum(nil)),
				TaskName: t.Name,
				TaskPath: taskPath,
			})
		}
	}

	return entries, nil
}
