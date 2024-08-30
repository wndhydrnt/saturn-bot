package task

import (
	"cmp"
	"fmt"
	"path"
	"time"

	"github.com/gosimple/slug"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
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

type Task interface {
	Actions() []action.Action
	AutoMergeAfter() time.Duration
	BranchName() string
	Checksum() string
	Filters() []filter.Filter
	HasReachedChangeLimit() bool
	HasReachMaxOpenPRs() bool
	IncChangeLimitCount()
	IncOpenPRsCount()
	OnPrClosed(host.Repository) error
	OnPrCreated(host.Repository) error
	OnPrMerged(host.Repository) error
	PrTitle() string
	SetLogger(*zap.SugaredLogger)
	SourceTask() schema.Task
	Stop()
}

type Wrapper struct {
	actions                []action.Action
	autoMergeAfterDuration *time.Duration
	changeLimitCount       int
	checksum               string
	filters                []filter.Filter
	openPRs                int
	plugins                []*pluginWrapper
	Task                   schema.Task
}

func (tw *Wrapper) Actions() []action.Action {
	return tw.actions
}

func (tw *Wrapper) AddFilters(f ...filter.Filter) {
	tw.filters = append(tw.filters, f...)
}

func (tw *Wrapper) Filters() []filter.Filter {
	return tw.filters
}

func (tw *Wrapper) BranchName() string {
	return cmp.Or(tw.Task.BranchName, "saturn-bot--"+slug.Make(tw.Task.Name))
}

func (tw *Wrapper) Checksum() string {
	return tw.checksum
}

func (tw *Wrapper) AutoMergeAfter() time.Duration {
	if tw.Task.AutoMergeAfter == "" {
		return 0
	}

	if tw.autoMergeAfterDuration == nil {
		d, err := time.ParseDuration(tw.Task.AutoMergeAfter)
		if err != nil {
			panic(fmt.Sprintf("value of field `autoMergeAfter` of task %s is not a duration: %s", tw.Task.Name, err))
		}

		tw.autoMergeAfterDuration = &d
	}

	return *tw.autoMergeAfterDuration
}

func (tw *Wrapper) HasReachedChangeLimit() bool {
	if tw.Task.ChangeLimit == 0 {
		return false
	}

	return tw.changeLimitCount >= tw.Task.ChangeLimit
}

func (tw *Wrapper) HasReachMaxOpenPRs() bool {
	if tw.Task.MaxOpenPRs == 0 {
		return false
	}

	return tw.openPRs >= tw.Task.MaxOpenPRs
}

func (tw *Wrapper) IncChangeLimitCount() {
	if tw.Task.ChangeLimit > 0 {
		tw.changeLimitCount++
	}
}

func (tw *Wrapper) IncOpenPRsCount() {
	if tw.Task.MaxOpenPRs > 0 {
		tw.openPRs++
	}
}

func (tw *Wrapper) PrTitle() string {
	return cmp.Or(tw.Task.PrTitle, "Apply task "+tw.Task.Name)
}

func (tw *Wrapper) SetLogger(l *zap.SugaredLogger) {
	if l == nil {
		return
	}

	for _, p := range tw.plugins {
		p.setLogger(l)
	}
}

func (tw *Wrapper) SourceTask() schema.Task {
	return tw.Task
}

func (tw *Wrapper) OnPrClosed(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrClosed(repository)
		if err != nil {
			log.Log().Errorw("OnPrClosed event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Wrapper) OnPrCreated(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrCreated(repository)
		if err != nil {
			log.Log().Errorw("OnPrCreated event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Wrapper) OnPrMerged(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrMerged(repository)
		if err != nil {
			log.Log().Errorw("OnPrMerged event of plugin failed", zap.Error(err))
		}
	}

	return nil
}

func (tw *Wrapper) Stop() {
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
	tasks           []Task
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
func (tr *Registry) GetTasks() []Task {
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
