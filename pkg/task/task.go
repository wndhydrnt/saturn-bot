package task

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosimple/slug"
	proto "github.com/wndhydrnt/saturn-sync-go/protocol/v1"
	"github.com/wndhydrnt/saturn-sync/pkg/action"
	"github.com/wndhydrnt/saturn-sync/pkg/filter"
	"github.com/wndhydrnt/saturn-sync/pkg/host"
)

func createActionsForTask(actionDefs *proto.Actions, taskPath string) ([]action.Action, error) {
	var result []action.Action
	if actionDefs == nil {
		return result, nil
	}

	for _, fileAction := range actionDefs.File {
		contentReader, err := createContentReader(taskPath, fileAction.GetContent())
		if err != nil {
			return nil, fmt.Errorf("create content reader of file action: %w", err)
		}
		a, err := action.NewFile(
			contentReader,
			fileAction.GetMode(),
			fileAction.GetPath(),
			fileAction.GetState(),
			fileAction.Overwrite,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, a)
	}

	for _, lineInFileAction := range actionDefs.LineInFile {
		a, err := action.NewLineInFile(
			lineInFileAction.GetInsertAt(),
			lineInFileAction.GetLine(),
			lineInFileAction.GetPath(),
			lineInFileAction.GetRegex(),
			lineInFileAction.GetState(),
		)

		if err != nil {
			return nil, err
		}

		result = append(result, a)
	}

	return result, nil
}

func createContentReader(taskPath, value string) (io.ReadCloser, error) {
	if !strings.HasPrefix(value, "$file:") {
		// Action is not requesting a file.
		// Return the string and adhere to io.ReadCloser.
		return io.NopCloser(strings.NewReader(value)), nil
	}

	filePath := strings.TrimSpace(strings.TrimPrefix(value, "$file:"))
	taskDir := filepath.Dir(taskPath)
	abs, err := filepath.Abs(filepath.Join(taskDir, filePath))
	if err != nil {
		return nil, fmt.Errorf("get absolute path of %s: %w", filePath, err)
	}

	if !strings.HasPrefix(abs, taskDir) {
		return nil, fmt.Errorf("path escapes directory of task")
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open content reader: %w", err)
	}

	return f, nil
}

func createFiltersForTask(filterDefs *proto.Filters) ([]filter.Filter, error) {
	var result []filter.Filter
	if filterDefs == nil {
		return result, nil
	}

	for _, rnFilter := range filterDefs.RepositoryNames {
		f, err := filter.NewRepositoryNames(rnFilter.GetNames())
		if err != nil {
			return nil, err
		}

		if rnFilter.GetReverse() {
			result = append(result, filter.NewReverse(f))
		} else {
			result = append(result, f)
		}
	}

	for _, fFilter := range filterDefs.Files {
		f := filter.NewFileExists(fFilter.GetPath())
		if fFilter.GetReverse() {
			result = append(result, filter.NewReverse(f))
		} else {
			result = append(result, f)
		}
	}

	for _, fcFilter := range filterDefs.FileContents {
		f, err := filter.NewFileContainsLine(fcFilter.GetPath(), fcFilter.GetSearch())
		if err != nil {
			return nil, err
		}

		if fcFilter.GetReverse() {
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
	SourceTask() *proto.Task
	Stop() error
}

type Wrapper struct {
	actions         []action.Action
	checksum        string
	filters         []filter.Filter
	Task            *proto.Task
	onPrClosedFunc  func(host.Repository) error
	onPrCreatedFunc func(host.Repository) error
	onPrMergedFunc  func(host.Repository) error
	stopFunc        func() error
}

func (tw *Wrapper) Actions() []action.Action {
	return tw.actions
}

func (tw *Wrapper) Filters() []filter.Filter {
	return tw.filters
}

func (tw *Wrapper) BranchName() string {
	if tw.Task.GetBranchName() == "" {
		return "saturn-sync--" + slug.Make(tw.Task.GetName())
	}

	return tw.Task.GetBranchName()
}

func (tw *Wrapper) Checksum() string {
	return tw.checksum
}

func (tw *Wrapper) AutoMergeAfter() time.Duration {
	return time.Second * time.Duration(tw.Task.GetAutoMergeAfterSeconds())
}

func (tw *Wrapper) PrTitle() string {
	if tw.Task.GetPrTitle() == "" {
		return "Apply task " + tw.Task.GetName()
	}

	return tw.Task.GetPrTitle()
}

func (tw *Wrapper) SourceTask() *proto.Task {
	return tw.Task
}

func (tw *Wrapper) OnPrClosed(repository host.Repository) error {
	if tw.onPrCreatedFunc != nil {
		return tw.onPrClosedFunc(repository)
	}

	return nil
}

func (tw *Wrapper) OnPrCreated(repository host.Repository) error {
	if tw.onPrCreatedFunc != nil {
		return tw.onPrCreatedFunc(repository)
	}

	return nil
}

func (tw *Wrapper) OnPrMerged(repository host.Repository) error {
	if tw.onPrMergedFunc != nil {
		return tw.onPrMergedFunc(repository)
	}

	return nil
}

func (tw *Wrapper) Stop() error {
	if tw.stopFunc != nil {
		return tw.stopFunc()
	}

	return nil
}

// Registry contains all tasks.
type Registry struct {
	customConfig []byte
	tasks        []Task
}

func NewRegistry(customConfig []byte) *Registry {
	return &Registry{customConfig: customConfig}
}

// GetTasks returns all tasks registered with the Registry.
// Should be called only after ReadAll() has been called at least once.
func (tr *Registry) GetTasks() []Task {
	return tr.tasks
}

// ReadAll takes a list of paths to task files and reads all tasks from thee files.
func (tr *Registry) ReadAll(taskFiles []string) {
	for _, file := range taskFiles {
		err := tr.readTasks(file)
		if err != nil {
			slog.Error("Failed to read tasks from file", "err", err, "file", file)
		}
	}
}

func (tr *Registry) readTasks(taskFile string) error {
	ext := path.Ext(taskFile)
	switch ext {
	case "":
		tasks, err := startPlugin(&startPluginOptions{
			customConfig: tr.customConfig,
			executable:   taskFile,
			filePath:     taskFile,
		})
		if err != nil {
			return fmt.Errorf("start plugin of file '%s': %w", taskFile, err)
		}

		tr.tasks = append(tr.tasks, tasks...)
		return nil
	case ".py":
		tasks, err := startPlugin(&startPluginOptions{
			args:         []string{taskFile},
			customConfig: tr.customConfig,
			executable:   "python",
			filePath:     taskFile,
		})
		if err != nil {
			return fmt.Errorf("start plugin of file '%s': %w", taskFile, err)
		}

		tr.tasks = append(tr.tasks, tasks...)
		return nil
	case ".yml":
	case ".yaml":
		tasks, err := readTasksYaml(taskFile)
		if err != nil {
			return err
		}

		tr.tasks = append(tr.tasks, tasks...)
		return nil
	default:
		return fmt.Errorf("unsupported file: %s", taskFile)
	}

	return fmt.Errorf("unsupported file: %s", taskFile)
}

// Stop notifies every task that this Registry is being stopped.
// This usually happens at the end of a run.
// Every task can then execute code to clean itself up.
func (tr *Registry) Stop() {
	for _, t := range tr.tasks {
		_ = t.Stop()
	}
}

func calculateChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file to create hash: %w", err)
	}

	defer f.Close()
	hash := sha256.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return "", fmt.Errorf("copy content to sha256: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
