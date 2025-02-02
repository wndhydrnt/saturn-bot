package task

import (
	"bytes"
	"context"
	"fmt"
	htmlTemplate "html/template"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gosimple/slug"
	protoV1 "github.com/wndhydrnt/saturn-bot-go/protocol/v1"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
	"github.com/wndhydrnt/saturn-bot/pkg/plugin"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

func createFiltersForTask(filterDefs []schema.Filter, factories options.FilterFactories, hosts []host.Host) ([]filter.Filter, []filter.Filter, error) {
	var preClone []filter.Filter
	var postClone []filter.Filter
	if filterDefs == nil {
		return preClone, postClone, nil
	}

	filterCreateOptions := filter.CreateOptions{Hosts: hosts}
	for idx, def := range filterDefs {
		factory := factories.Find(def.Filter)
		if factory == nil {
			return nil, nil, fmt.Errorf("no filter registered for identifier %s", def.Filter)
		}

		preF, err := factory.CreatePreClone(filterCreateOptions, params.Params(def.Params))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize pre-clone filter %s at %d: %w", def.Filter, idx, err)
		}

		if preF != nil {
			if def.Reverse {
				preF = filter.NewReverse(preF)
			}

			preClone = append(preClone, preF)
		}

		postF, err := factory.CreatePostClone(filterCreateOptions, params.Params(def.Params))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize post-clone filter %s at %d: %w", def.Filter, idx, err)
		}

		if postF != nil {
			if def.Reverse {
				postF = filter.NewReverse(postF)
			}

			postClone = append(postClone, postF)
		}
	}

	return preClone, postClone, nil
}

type Task struct {
	schema.Task
	actions                []action.Action
	autoMergeAfterDuration *time.Duration
	changeLimitCount       int
	checksum               string
	filtersPreClone        []filter.Filter
	filtersPostClone       []filter.Filter
	openPRs                int
	path                   string // Path to the file that contains the task.
	plugins                []*plugin.Plugin
	templateBranchName     *htmlTemplate.Template
	templatePrTitle        *htmlTemplate.Template
	inputData              map[string]string
	inputValidators        map[string]*regexp.Regexp
}

func (tw *Task) Actions() []action.Action {
	return tw.actions
}

func (tw *Task) AddPreCloneFilters(f ...filter.Filter) {
	tw.filtersPreClone = append(tw.filtersPreClone, f...)
}

func (tw *Task) FiltersPreClone() []filter.Filter {
	return tw.filtersPreClone
}

func (tw *Task) FiltersPostClone() []filter.Filter {
	return tw.filtersPostClone
}

func (tw *Task) HasFilters() bool {
	return len(tw.filtersPreClone) > 0 || len(tw.filtersPostClone) > 0
}

func (tw *Task) RenderBranchName(data template.Data) (string, error) {
	var name string
	if tw.BranchName == "" {
		name = "saturn-bot--" + slug.Make(tw.Name)
	} else {
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

		name = buf.String()
	}

	// Some git hosts place a limit on character length of branch names.
	if len(name) > 230 {
		return name[:230], nil
	}

	return name, nil
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

// SetInputs takes inputs supplied via the command-line and stores them for later processing.
// It returns an error if an expected input isn't supplied and no default value for the input has been set.
func (tw *Task) SetInputs(cliInputs map[string]string) error {
	if tw.inputData == nil {
		tw.inputData = map[string]string{}
	}

	for _, input := range tw.Inputs {
		value := cliInputs[input.Name]
		if value == "" {
			if input.Default == nil {
				return fmt.Errorf("input %s not set and has no default value", input.Name)
			} else {
				tw.inputData[input.Name] = ptr.From(input.Default)
			}
		} else {
			tw.inputData[input.Name] = value
		}
	}

	return nil
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

// InputData returns a map where every key is an input defined by the task
// and every value is what has been set via the command-line for that key.
func (tw *Task) InputData() map[string]string {
	if tw.inputData == nil {
		tw.inputData = map[string]string{}
	}

	return tw.inputData
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

func (tw *Task) Path() string {
	return tw.path
}

func (tw *Task) Stop() {
	for _, p := range tw.plugins {
		p.Stop()
	}
}

// ValidateInputs validates the values in data.
// It returns a list of errors where each error details a failed validation.
func ValidateInputs(data map[string]string, t *Task) []error {
	var errors []error
	for _, input := range t.Inputs {
		value := data[input.Name]
		if value == "" && input.Default == nil {
			errors = append(errors, fmt.Errorf("missing value for input '%s'", input.Name))
			continue
		}

		validator := t.inputValidators[input.Name]
		if validator != nil && !validator.MatchString(value) {
			errors = append(errors, fmt.Errorf("value of input '%s' does not match regular expression '%s'", input.Name, validator.String()))
			continue
		}

		if len(input.Options) > 0 && slices.Contains(input.Options, value) {
			errors = append(errors, fmt.Errorf("value of input '%s' must be one of '%s'", input.Name, strings.Join(input.Options, ",")))
			continue
		}
	}

	return errors
}

// Registry contains all tasks.
type Registry struct {
	actionFactories options.ActionFactories
	filterFactories options.FilterFactories
	hosts           []host.Host
	isCi            bool
	pathJava        string
	pathPython      string
	pluginLogLevel  zapcore.Level
	skipPlugins     bool
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
		hosts:           opts.Hosts,
		isCi:            opts.IsCi,
		pathJava:        opts.Config.JavaPath,
		pathPython:      opts.Config.PythonPath,
		pluginLogLevel:  lvl,
		skipPlugins:     opts.SkipPlugins,
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
		err := tr.ReadTasks(path)
		if err != nil {
			return fmt.Errorf("failed to read tasks from file %s: %w", path, err)
		}
	}

	return nil
}

// ReadTasks reads all tasks defined in taskFile and adds them to the registry.
func (tr *Registry) ReadTasks(taskFile string) error {
	results, err := schema.Read(taskFile)
	if err != nil {
		return err
	}

	for _, entry := range results {
		if !entry.Task.Active {
			log.Log().Warnf("Task %s deactivated", entry.Task.Name)
			continue
		}

		wrapper := &Task{
			checksum: entry.Sha256,
			path:     entry.Path,
		}
		wrapper.Task = entry.Task
		wrapper.actions, err = createActionsForTask(wrapper.Task.Actions, tr.actionFactories, entry.Path)
		if err != nil {
			return fmt.Errorf("parse actions of task: %w", err)
		}

		wrapper.filtersPreClone, wrapper.filtersPostClone, err = createFiltersForTask(wrapper.Task.Filters, tr.filterFactories, tr.hosts)
		if err != nil {
			return fmt.Errorf("parse filters of task file '%s': %w", entry.Path, err)
		}

		for idx, taskPlugin := range entry.Task.Plugins {
			if tr.skipPlugins {
				continue
			}

			p, err := tr.startPlugin(entry.Path, taskPlugin)
			if err != nil {
				return fmt.Errorf("initialize plugin #%d: %w", idx, err)
			}

			wrapper.actions = append(wrapper.actions, p)
			wrapper.filtersPreClone = append(wrapper.filtersPreClone, p)
			wrapper.plugins = append(wrapper.plugins, p)
		}

		if len(entry.Task.Inputs) > 0 {
			wrapper.inputValidators = map[string]*regexp.Regexp{}
			for _, input := range entry.Task.Inputs {
				if input.Validation != nil {
					// No need to check for error. Has been validated during parsing.
					r, _ := regexp.Compile(ptr.From(input.Validation))
					wrapper.inputValidators[input.Name] = r
				}
			}
		}

		tr.tasks = append(tr.tasks, wrapper)
	}

	return nil
}

func (tr *Registry) startPlugin(taskPath string, taskPlugin schema.Plugin) (*plugin.Plugin, error) {
	pluginConfiguration := make(schema.PluginConfiguration, len(taskPlugin.Configuration))
	// Copy to not modify the original
	for k, v := range taskPlugin.Configuration {
		pluginConfiguration[k] = v
	}

	if tr.isCi {
		pluginConfiguration["saturn-bot.ci"] = "true"
	}

	opts := plugin.StartOptions{
		Config:       pluginConfiguration,
		Exec:         plugin.GetPluginExec(taskPlugin.PathAbs(taskPath), tr.pathJava, tr.pathPython),
		OnDataStderr: plugin.NewStderrHandler(tr.pluginLogLevel),
		OnDataStdout: plugin.NewStdoutHandler(tr.pluginLogLevel),
	}

	p := &plugin.Plugin{}
	err := p.Start(opts)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Stop notifies every task that this Registry is being stopped.
// This usually happens at the end of a run.
// Every task can then execute code to clean itself up.
func (tr *Registry) Stop() {
	for _, t := range tr.tasks {
		t.Stop()
	}
}
