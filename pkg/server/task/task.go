package task

import (
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

// Load reads all tasks from the file system.
// It does not use [github.com/wndhydrnt/saturn-bot/pkg/task.Registry]
// because the server doesn't need to start the tasks, which the registry would do.
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
