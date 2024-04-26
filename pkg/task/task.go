package task

import (
	"cmp"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/gosimple/slug"
	"github.com/wndhydrnt/saturn-sync/pkg/action"
	"github.com/wndhydrnt/saturn-sync/pkg/filter"
	"github.com/wndhydrnt/saturn-sync/pkg/host"
	"github.com/wndhydrnt/saturn-sync/pkg/task/schema"
)

func createActionsForTask(actionDefs *schema.TaskActions, taskPath string) ([]action.Action, error) {
	var result []action.Action
	if actionDefs == nil {
		return result, nil
	}

	for idx, fileCreate := range actionDefs.FileCreate {
		a, err := action.NewFileCreateFromTask(fileCreate, taskPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create fileCreate[%d] action: %w", idx, err)
		}

		result = append(result, a)
	}

	for _, fileDelete := range actionDefs.FileDelete {
		result = append(result, action.NewFileDelete(fileDelete.Path))
	}

	for _, lineDelete := range actionDefs.LineDelete {
		a, err := action.NewLineDelete(lineDelete.Line, lineDelete.Path)
		if err != nil {
			return nil, fmt.Errorf("initialize lineDelete action: %w", err)
		}

		result = append(result, a)
	}

	for _, lineInsert := range actionDefs.LineInsert {
		a, err := action.NewLineInsert(string(lineInsert.InsertAt), lineInsert.Line, lineInsert.Path)
		if err != nil {
			return nil, fmt.Errorf("initialize lineInsert action: %w", err)
		}

		result = append(result, a)
	}

	for _, lineReplace := range actionDefs.LineReplace {
		a, err := action.NewLineReplace(lineReplace.Line, lineReplace.Path, lineReplace.Search)
		if err != nil {
			return nil, fmt.Errorf("initialize lineReplace action: %w", err)
		}

		result = append(result, a)
	}

	return result, nil
}

func createFiltersForTask(filterDefs *schema.TaskFilters) ([]filter.Filter, error) {
	var result []filter.Filter
	if filterDefs == nil {
		return result, nil
	}

	for _, rnFilter := range filterDefs.RepositoryName {
		f, err := filter.NewRepositoryName(rnFilter.Names)
		if err != nil {
			return nil, err
		}

		if rnFilter.Reverse {
			result = append(result, filter.NewReverse(f))
		} else {
			result = append(result, f)
		}
	}

	for _, fFilter := range filterDefs.File {
		f := filter.NewFile(fFilter.Path)
		if fFilter.Reverse {
			result = append(result, filter.NewReverse(f))
		} else {
			result = append(result, f)
		}
	}

	for _, fcFilter := range filterDefs.FileContent {
		f, err := filter.NewFileContent(fcFilter.Path, fcFilter.Search)
		if err != nil {
			return nil, err
		}

		if fcFilter.Reverse {
			result = append(result, filter.NewReverse(f))
		} else {
			result = append(result, f)
		}
	}

	return result, nil
}

type Task interface {
	Actions() []action.Action
	AutoMergeAfter() time.Duration
	BranchName() string
	Checksum() string
	Filters() []filter.Filter
	OnPrClosed(host.Repository) error
	OnPrCreated(host.Repository) error
	OnPrMerged(host.Repository) error
	PrTitle() string
	SourceTask() *schema.Task
	Stop()
}

type Wrapper struct {
	actions                []action.Action
	autoMergeAfterDuration *time.Duration
	checksum               string
	filters                []filter.Filter
	plugins                []pluginWrapper
	Task                   *schema.Task
}

func (tw *Wrapper) Actions() []action.Action {
	return tw.actions
}

func (tw *Wrapper) Filters() []filter.Filter {
	return tw.filters
}

func (tw *Wrapper) BranchName() string {
	return cmp.Or(tw.Task.BranchName, "saturn-sync--"+slug.Make(tw.Task.Name))
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

func (tw *Wrapper) PrTitle() string {
	return cmp.Or(tw.Task.PrTitle, "Apply task "+tw.Task.Name)
}

func (tw *Wrapper) SourceTask() *schema.Task {
	return tw.Task
}

func (tw *Wrapper) OnPrClosed(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrClosed(repository)
		if err != nil {
			slog.Error("OnPrClosed event of plugin failed", "err", err)
		}
	}

	return nil
}

func (tw *Wrapper) OnPrCreated(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrCreated(repository)
		if err != nil {
			slog.Error("OnPrCreated event of plugin failed", "err", err)
		}
	}

	return nil
}

func (tw *Wrapper) OnPrMerged(repository host.Repository) error {
	for _, p := range tw.plugins {
		err := p.onPrMerged(repository)
		if err != nil {
			slog.Error("OnPrMerged event of plugin failed", "err", err)
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
	tasks []Task
}

func NewRegistry() *Registry {
	return &Registry{}
}

// GetTasks returns all tasks registered with the Registry.
// Should be called only after ReadAll() has been called at least once.
func (tr *Registry) GetTasks() []Task {
	return tr.tasks
}

// ReadAll takes a list of paths to task files and reads all tasks from thee files.
func (tr *Registry) ReadAll(taskFiles []string) error {
	for _, file := range taskFiles {
		err := tr.readTasks(file)
		if err != nil {
			return fmt.Errorf("failed to read tasks from file %s: %w", file, err)
		}
	}

	return nil
}

func (tr *Registry) readTasks(taskFile string) error {
	ext := path.Ext(taskFile)
	switch ext {
	case ".yml":
	case ".yaml":
		tasks, err := readTasksYaml(taskFile)
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

func calculateChecksum(filePath string, h hash.Hash) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file to calculate hash: %w", err)
	}

	defer f.Close()
	_, err = io.Copy(h, f)
	if err != nil {
		return fmt.Errorf("copy content to hash: %w", err)
	}

	return nil
}
