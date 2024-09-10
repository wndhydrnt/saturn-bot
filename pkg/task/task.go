package task

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"path"
	"time"

	"github.com/gosimple/slug"
	"github.com/robfig/cron"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
	"go.uber.org/zap"
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

		action, err := factory.Create(def.Params, taskPath)
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

		fl, err := factory.Create(def.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize filter %s at %d: %w", def.Filter, idx, err)
		}

		if def.Reverse {
			fl = filter.NewReverse(fl)
		}

		result = append(result, fl)
	}

	return result, nil
}

// type Task interface {
// 	Actions() []action.Action
// 	AutoMergeAfter() time.Duration
// 	BranchName(template.Data) (string, error)
// 	Checksum() string
// 	Filters() []filter.Filter
// 	HasReachedChangeLimit() bool
// 	HasReachMaxOpenPRs() bool
// 	IncChangeLimitCount()
// 	IncOpenPRsCount()
// 	IsWithinSchedule() bool
// 	OnPrClosed(host.Repository) error
// 	OnPrCreated(host.Repository) error
// 	OnPrMerged(host.Repository) error
// 	PrTitle(template.Data) (string, error)
// 	SetLogger(*zap.SugaredLogger)
// 	SourceTask() schema.Task
// 	Stop()
// }

type Task struct {
	schema.Task
	actions                []action.Action
	autoMergeAfterDuration *time.Duration
	changeLimitCount       int
	checksum               string
	filters                []filter.Filter
	openPRs                int
	plugins                []*pluginWrapper
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

func (tw *Task) SetLogger(l *zap.SugaredLogger) {
	if l == nil {
		return
	}

	for _, p := range tw.plugins {
		p.setLogger(l)
	}
}

func (tw *Task) OnPrClosed(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrClosed(repository)
		if err != nil {
			log.Log().Errorw("OnPrClosed event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Task) OnPrCreated(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrCreated(repository)
		if err != nil {
			log.Log().Errorw("OnPrCreated event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Task) OnPrMerged(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrMerged(repository)
		if err != nil {
			log.Log().Errorw("OnPrMerged event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Task) Stop() {
	for _, p := range tw.plugins {
		p.stop()
	}
}

// Registry contains all tasks.
type Registry struct {
	actionFactories options.ActionFactories
	filterFactories options.FilterFactories
	pathJava        string
	pathPython      string
	tasks           []*Task
}

func NewRegistry(opts options.Opts) *Registry {
	return &Registry{
		actionFactories: opts.ActionFactories,
		filterFactories: opts.FilterFactories,
		pathJava:        opts.Config.JavaPath,
		pathPython:      opts.Config.PythonPath,
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
	ext := path.Ext(taskFile)
	switch ext {
	case ".yml":
	case ".yaml":
		tasks, err := readTasksYaml(tr.actionFactories, tr.filterFactories, tr.pathJava, tr.pathPython, taskFile)
		if err != nil {
			return err
		}

		tr.tasks = append(tr.tasks, tasks...)
		return nil
	default:
		return fmt.Errorf("unsupported extension: %s", ext)
	}

	return fmt.Errorf("unsupported extension: %s", ext)
}

// Stop notifies every task that this Registry is being stopped.
// This usually happens at the end of a run.
// Every task can then execute code to clean itself up.
func (tr *Registry) Stop() {
	for _, t := range tr.tasks {
		t.Stop()
	}
}
