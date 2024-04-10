package task

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"

	proto "github.com/wndhydrnt/saturn-sync-go/protocol/v1"
	"gopkg.in/yaml.v3"
)

func readTasksYaml(taskFile string) ([]Task, error) {
	var result []Task
	b, err := os.ReadFile(taskFile)
	if err != nil {
		return nil, fmt.Errorf("read task file '%s': %w", taskFile, err)
	}

	checksum := sha256.Sum256(b)
	dec := yaml.NewDecoder(bytes.NewReader(b))
	for {
		task := &proto.Task{}
		err := dec.Decode(task)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("decode task file '%s' from YAML: %w", taskFile, err)
		}

		wrapper := &Wrapper{}
		wrapper.checksum = fmt.Sprintf("%x", checksum)
		wrapper.Task = task
		wrapper.actions, err = createActionsForTask(wrapper.Task.Actions, taskFile)
		if err != nil {
			return nil, fmt.Errorf("parse actions of task file '%s': %w", taskFile, err)
		}

		wrapper.filters, err = createFiltersForTask(wrapper.Task.Filters)
		if err != nil {
			return nil, fmt.Errorf("parse filters of task file '%s': %w", taskFile, err)
		}

		result = append(result, wrapper)
	}

	return result, nil
}
