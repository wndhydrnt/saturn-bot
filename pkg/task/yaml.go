package task

import (
	"fmt"

	"github.com/robfig/cron"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
)

func readTasksYaml(
	actionFactories options.ActionFactories,
	filterFactories options.FilterFactories,
	pathJava string,
	pathPython string,
	taskFile string,
) ([]*Task, error) {
	results, err := schema.Read(taskFile)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	for _, entry := range results {
		if !entry.Task.Active {
			log.Log().Warnf("Task %s deactivated", entry.Task.Name)
			continue
		}

		wrapper := &Task{}
		wrapper.Task = entry.Task
		wrapper.actions, err = createActionsForTask(wrapper.Task.Actions, actionFactories, entry.Path)
		if err != nil {
			return nil, fmt.Errorf("parse actions of task: %w", err)
		}

		wrapper.filters, err = createFiltersForTask(wrapper.Task.Filters, filterFactories)
		if err != nil {
			return nil, fmt.Errorf("parse filters of task file '%s': %w", entry.Path, err)
		}

		for idx, plugin := range entry.Task.Plugins {
			pw, err := newPluginWrapper(startPluginOptions{
				config:     plugin.Configuration,
				filePath:   plugin.PathAbs(entry.Path),
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

		schedule, err := cron.ParseStandard(entry.Task.Schedule)
		if err != nil {
			return nil, fmt.Errorf("parse schedule: %w", err)
		}
		wrapper.schedule = schedule

		wrapper.checksum = fmt.Sprintf("%x", entry.Hash.Sum(nil))
		tasks = append(tasks, wrapper)
	}

	return tasks, nil
}
