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
		results, err := schema.Read(taskPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range results {
			entries = append(entries, Task{
				Hash:     fmt.Sprintf("%x", entry.Hash.Sum(nil)),
				TaskName: entry.Task.Name,
				TaskPath: entry.Path,
			})
		}
	}

	return entries, nil
}
