package task

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func readTasksYaml(
	actionFactories options.ActionFactories,
	filterFactories options.FilterFactories,
	pathJava string,
	pathPython string,
	taskFile string,
) ([]Task, error) {
	schemaTasks, hashes, err := schema.Read(taskFile)
	if err != nil {
		return nil, err
	}

	var result []Task
	for idx, schemaTask := range schemaTasks {
		if !schemaTask.Active {
			slog.Warn("Task deactivated", "task", schemaTask.Name)
			continue
		}

		wrapper := &Wrapper{}
		wrapper.Task = schemaTask
		wrapper.actions, err = createActionsForTask(wrapper.Task.Actions, actionFactories, taskFile)
		if err != nil {
			return nil, fmt.Errorf("parse actions of task: %w", err)
		}

		wrapper.filters, err = createFiltersForTask(wrapper.Task.Filters, filterFactories)
		if err != nil {
			return nil, fmt.Errorf("parse filters of task file '%s': %w", taskFile, err)
		}

		for idx, plugin := range schemaTask.Plugins {
			pw, err := newPluginWrapper(startPluginOptions{
				config:     plugin.Configuration,
				filePath:   resolvePluginPath(plugin.Path, taskFile),
				pathJava:   pathJava,
				pathPython: pathPython,
			})
			if err != nil {
				return nil, fmt.Errorf("initialize plugin #%d: %w", idx, err)
			}

			wrapper.actions = append(wrapper.actions, pw.action)
			wrapper.filters = append(wrapper.filters, pw.filter)
			wrapper.plugins = append(wrapper.plugins, pw)
		}

		wrapper.checksum = fmt.Sprintf("%x", hashes[idx].Sum(nil))
		result = append(result, wrapper)
	}

	return result, nil
}

// resolvePluginPath resolves the path of a plugin relative to the path of its task.
func resolvePluginPath(pluginPath, taskPath string) string {
	if filepath.IsAbs(pluginPath) {
		return pluginPath
	}

	dir := filepath.Dir(taskPath)
	return filepath.Join(dir, pluginPath)
}
