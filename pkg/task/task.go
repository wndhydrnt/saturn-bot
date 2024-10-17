package task

import (
	"bytes"
	"context"
	"fmt"
	htmlTemplate "html/template"
	"slices"
	"time"

	"github.com/gosimple/slug"
	"github.com/robfig/cron"
	protoV1 "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"github.com/wndhydrnt/saturn-bot/pkg/plugin"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate go-jsonschema --extra-imports -p schema -t ./schema/task.schema.json --output ./schema/schema.go

func createActionsForTask(actionDefs []schema.Action, factories options.ActionFactories, taskPath string) ([]action.Action, error) {
	var result []action.Action
	if actionDefs == nil {
		return result, nil
	}

	for idx, def := range actionDefs {
		factory := factories.Find(def.Action)
		if factory == nil {
			return nil, fmt.Errorf("no action registered for identifier %s", def.Action)
		}

		action, err := factory.Create(params.Params(def.Params), taskPath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize action %s at %d: %w", def.Action, idx, err)
		}

		result = append(result, action)
	}

	return result, nil
}

func createFiltersForTask(filterDefs []schema.Filter, factories options.FilterFactories) ([]filter.Filter, error) {
	var result []filter.Filter
	if filterDefs == nil {
		return result, nil
	}

	for idx, def := range filterDefs {
		factory := factories.Find(def.Filter)
		if factory == nil {
			return nil, fmt.Errorf("no filter registered for identifier %s", def.Filter)
		}

		fl, err := factory.Create(params.Params(def.Params))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize filter %s at %d: %w", def.Filter, idx, err)
		}

		if def.Reverse {
			fl = filter.NewReverse(fl)
		}

		result = append(result, fl)
	}

	slices.SortStableFunc(result, func(a filter.Filter, b filter.Filter) int {
		_, aOk := a.(*filter.Repository)
		_, bOk := b.(*filter.Repository)
		// Both of type Repository. Keep order.
		if aOk && bOk {
			return 0
		}
		// None of type Repository. Keep order.
		if !aOk && !bOk {
			return 0
		}
		// a is of type Repository.
		// Put it before b.
		if aOk && !bOk {
			return -1
		}
		// b is of type Repository.
		// Put it before a.
		return 1
	})

	return result, nil
}

type Task struct {
	schema.Task
	actions                []action.Action
	autoMergeAfterDuration *time.Duration
	changeLimitCount       int
	checksum               string
	filters                []filter.Filter
	openPRs                int
	plugins                []*plugin.Plugin
	schedule               cron.Schedule
	templateBranchName     *htmlTemplate.Template
	templatePrTitle        *htmlTemplate.Template
}

func (tw *Task) Actions() []action.Action {
	return tw.actions
}

func (tw *Task) AddFilters(f ...filter.Filter) {
	tw.filters = append(tw.filters, f...)
}

func (tw *Task) Filters() []filter.Filter {
	return tw.filters
}

func (tw *Task) RenderBranchName(data template.Data) (string, error) {
	if tw.BranchName == "" {
		return "saturn-bot--" + slug.Make(tw.Name), nil
	}

	if tw.templateBranchName == nil {
		var parseErr error
		tw.templateBranchName, parseErr = htmlTemplate.New("").Parse(tw.BranchName)
		if parseErr != nil {
			return "", fmt.Errorf("parse branch name template: %w", parseErr)
		}
	}

	buf := &bytes.Buffer{}
	err := tw.templateBranchName.Execute(buf, data)
	if err != nil {
		return "", fmt.Errorf("render branch name template: %w", err)
	}

	return buf.String(), nil
}

func (tw *Task) Checksum() string {
	return tw.checksum
}

func (tw *Task) CalcAutoMergeAfter() time.Duration {
	if tw.AutoMergeAfter == "" {
		return 0
	}

	if tw.autoMergeAfterDuration == nil {
		d, err := time.ParseDuration(tw.AutoMergeAfter)
		if err != nil {
			panic(fmt.Sprintf("value of field `autoMergeAfter` of task %s is not a duration: %s", tw.Name, err))
		}

		tw.autoMergeAfterDuration = &d
	}

	return *tw.autoMergeAfterDuration
}

func (tw *Task) HasReachedChangeLimit() bool {
	if tw.ChangeLimit == 0 {
		return false
	}

	return tw.changeLimitCount >= tw.ChangeLimit
}

func (tw *Task) HasReachMaxOpenPRs() bool {
	if tw.MaxOpenPRs == 0 {
		return false
	}

	return tw.openPRs >= tw.MaxOpenPRs
}

func (tw *Task) IncChangeLimitCount() {
	if tw.ChangeLimit > 0 {
		tw.changeLimitCount++
	}
}

func (tw *Task) IncOpenPRsCount() {
	if tw.MaxOpenPRs > 0 {
		tw.openPRs++
	}
}

func (tw *Task) IsWithinSchedule() bool {
	if tw.schedule == nil {
		return true
	}

	now := time.Now()
	zero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return tw.schedule.Next(zero).Before(now)
}

var (
	tplPrTitleDefault = htmlTemplate.Must(htmlTemplate.New("titleDefault").Parse("saturn-bot: task {{.TaskName}}"))
)

func (tw *Task) RenderPrTitle(data template.Data) (string, error) {
	if tw.templatePrTitle == nil {
		if tw.PrTitle == "" {
			tw.templatePrTitle = tplPrTitleDefault
		} else {
			var parseErr error
			tw.templatePrTitle, parseErr = htmlTemplate.New("").Parse(tw.PrTitle)
			if parseErr != nil {
				return "", fmt.Errorf("parse pr title template: %w", parseErr)
			}
		}
	}

	buf := &bytes.Buffer{}
	err := tw.templatePrTitle.Execute(buf, data)
	if err != nil {
		return "", fmt.Errorf("render pr title template: %w", err)
	}

	return buf.String(), nil
}

func (tw *Task) OnPrClosed(ctx context.Context) error {
	for _, p := range tw.plugins {
		_, err := p.OnPrClosed(&protoV1.OnPrClosedRequest{Context: plugin.NewContext(ctx)})
		if err != nil {
			log.Log().Errorw("OnPrClosed event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Task) OnPrCreated(ctx context.Context) error {
	for _, p := range tw.plugins {
		_, err := p.OnPrCreated(&protoV1.OnPrCreatedRequest{Context: plugin.NewContext(ctx)})
		if err != nil {
			log.Log().Errorw("OnPrCreated event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Task) OnPrMerged(ctx context.Context) error {
	for _, p := range tw.plugins {
		_, err := p.OnPrMerged(&protoV1.OnPrMergedRequest{Context: plugin.NewContext(ctx)})
		if err != nil {
			log.Log().Errorw("OnPrMerged event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Task) Stop() {
	for _, p := range tw.plugins {
		p.Stop()
	}
}

// Registry contains all tasks.
type Registry struct {
	actionFactories options.ActionFactories
	filterFactories options.FilterFactories
	pathJava        string
	pathPython      string
	pluginLogLevel  zapcore.Level
	tasks           []*Task
}

func NewRegistry(opts options.Opts) *Registry {
	lvl, err := zapcore.ParseLevel(string(opts.Config.PluginLogLevel))
	if err != nil {
		lvl = zapcore.DebugLevel
	}

	return &Registry{
		actionFactories: opts.ActionFactories,
		filterFactories: opts.FilterFactories,
		pathJava:        opts.Config.JavaPath,
		pathPython:      opts.Config.PythonPath,
		pluginLogLevel:  lvl,
	}
}

// GetTasks returns all tasks registered with the Registry.
// Should be called only after ReadAll() has been called at least once.
func (tr *Registry) GetTasks() []*Task {
	return tr.tasks
}

// ReadAll takes a list of paths to task files and reads all tasks from the files.
func (tr *Registry) ReadAll(taskFiles []string) error {
	for _, path := range taskFiles {
		err := tr.readTasks(path)
		if err != nil {
			return fmt.Errorf("failed to read tasks from file %s: %w", path, err)
		}
	}

	return nil
}

func (tr *Registry) readTasks(taskFile string) error {
	results, err := schema.Read(taskFile)
	if err != nil {
		return err
	}

	for _, entry := range results {
		if !entry.Task.Active {
			log.Log().Warnf("Task %s deactivated", entry.Task.Name)
			continue
		}

		wrapper := &Task{}
		wrapper.Task = entry.Task
		wrapper.actions, err = createActionsForTask(wrapper.Task.Actions, tr.actionFactories, entry.Path)
		if err != nil {
			return fmt.Errorf("parse actions of task: %w", err)
		}

		wrapper.filters, err = createFiltersForTask(wrapper.Task.Filters, tr.filterFactories)
		if err != nil {
			return fmt.Errorf("parse filters of task file '%s': %w", entry.Path, err)
		}

		for idx, taskPlugin := range entry.Task.Plugins {
			opts := plugin.StartOptions{
				Config:       taskPlugin.Configuration,
				Exec:         plugin.GetPluginExec(taskPlugin.PathAbs(entry.Path), tr.pathJava, tr.pathPython),
				OnDataStderr: plugin.NewStderrHandler(tr.pluginLogLevel),
				OnDataStdout: plugin.NewStdoutHandler(tr.pluginLogLevel),
			}

			p := &plugin.Plugin{}
			err := p.Start(opts)
			if err != nil {
				return fmt.Errorf("initialize plugin #%d: %w", idx, err)
			}

			wrapper.actions = append(wrapper.actions, p)
			wrapper.filters = append(wrapper.filters, p)
			wrapper.plugins = append(wrapper.plugins, p)
		}

		schedule, err := cron.ParseStandard(entry.Task.Schedule)
		if err != nil {
			return fmt.Errorf("parse schedule: %w", err)
		}
		wrapper.schedule = schedule

		wrapper.checksum = fmt.Sprintf("%x", entry.Hash.Sum(nil))
		tr.tasks = append(tr.tasks, wrapper)
	}

	return nil
}

// Stop notifies every task that this Registry is being stopped.
// This usually happens at the end of a run.
// Every task can then execute code to clean itself up.
func (tr *Registry) Stop() {
	for _, t := range tr.tasks {
		t.Stop()
	}
}
