package task

import (
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

type Task struct {
	Hash     string
	TaskName string
	TaskPath string
}

func Load(taskPaths []string) ([]schema.ReadResult, error) {
	var entries []schema.ReadResult
	for _, taskPath := range taskPaths {
		results, err := schema.Read(taskPath)
		if err != nil {
			return nil, err
		}

		entries = append(entries, results...)
	}

	return entries, nil
}
