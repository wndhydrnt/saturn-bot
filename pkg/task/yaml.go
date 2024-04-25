package task

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/wndhydrnt/saturn-sync/pkg/task/schema"
	"gopkg.in/yaml.v3"
)

func readTasksYaml(taskFile string) ([]Task, error) {
	var result []Task
	b, err := os.ReadFile(taskFile)
	if err != nil {
		return nil, fmt.Errorf("read task file '%s': %w", taskFile, err)
	}

	checksum := sha256.New()
	_, _ = checksum.Write(b)
	dec := yaml.NewDecoder(bytes.NewReader(b))
	for {
		task := &schema.Task{}
		err := dec.Decode(task)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("decode task file '%s' from YAML: %w", taskFile, err)
		}

		err = schema.Validate(task)
		if err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}

		wrapper := &Wrapper{}
		wrapper.Task = task
		wrapper.actions, err = createActionsForTask(wrapper.Task.Actions, taskFile)
		if err != nil {
			return nil, fmt.Errorf("parse actions of task file '%s': %w", taskFile, err)
		}

		wrapper.filters, err = createFiltersForTask(wrapper.Task.Filters)
		if err != nil {
			return nil, fmt.Errorf("parse filters of task file '%s': %w", taskFile, err)
		}

		for idx, plugin := range task.Plugins {
			pw, err := newPluginWrapper(startPluginOptions{
				hash:     checksum,
				config:   plugin.Configuration,
				filePath: plugin.Path,
			})
			if err != nil {
				return nil, fmt.Errorf("initialize plugin #%d: %w", idx, err)
			}

			wrapper.actions = append(wrapper.actions, pw.action)
			wrapper.filters = append(wrapper.filters, pw.filter)
		}

		wrapper.checksum = fmt.Sprintf("%x", checksum.Sum(nil))
		result = append(result, wrapper)
	}

	return result, nil
}
