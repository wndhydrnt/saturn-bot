package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/wndhydrnt/saturn-bot/pkg/action"
	saturnContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
)

type TryRunner struct {
	ApplyActionsFunc func(actions []action.Action, ctx context.Context, dir string) error
	GitClient        git.GitClient
	Hosts            []host.Host
	Out              io.Writer
	Registry         *task.Registry
	RepositoryName   string
	TaskFile         string
	TaskName         string
}

func NewTryRunner(opts options.Opts, dataDir string, repositoryName string, taskFile string, taskName string) (*TryRunner, error) {
	if dataDir == "" {
		dataDir = path.Join(os.TempDir(), "saturn-bot")
	}

	// This code sets its own data dir.
	opts.Config.DataDir = &dataDir
	err := options.Initialize(&opts)
	if err != nil {
		return nil, fmt.Errorf("initialize options: %w", err)
	}

	gitClient, err := git.New(opts.Config)
	if err != nil {
		return nil, fmt.Errorf("new git client for try: %w", err)
	}

	return &TryRunner{
		ApplyActionsFunc: applyActionsInDirectory,
		GitClient:        gitClient,
		Hosts:            opts.Hosts,
		Out:              os.Stdout,
		Registry:         task.NewRegistry(opts),
		RepositoryName:   repositoryName,
		TaskFile:         taskFile,
		TaskName:         taskName,
	}, nil
}

func (r *TryRunner) Run() error {
	if r.RepositoryName == "" {
		return fmt.Errorf("required flag `--repository` is not set")
	}

	if r.TaskFile == "" {
		return fmt.Errorf("required flag `--task-file` is not set")
	}

	var repository host.Repository
	for _, host := range r.Hosts {
		var err error
		repository, err = host.CreateFromName(r.RepositoryName)
		if err != nil {
			return fmt.Errorf("create repository: %w", err)
		}

		if repository != nil {
			break
		}
	}

	if repository == nil {
		return fmt.Errorf("no host supports the repository")
	}

	err := r.Registry.ReadAll([]string{r.TaskFile})
	if err != nil {
		return err
	}

	tasks := r.Registry.GetTasks()
	if len(tasks) == 0 {
		fmt.Fprintf(r.Out, "⛔️ File %s does not contain any tasks\n", r.TaskFile)
		return nil
	}

	processed := false
	for _, task := range tasks {
		if r.TaskName != "" && task.SourceTask().Name != r.TaskName {
			continue
		}

		processed = true
		ctx := context.WithValue(context.Background(), saturnContext.RepositoryKey{}, repository)
		matched := true
		for _, filter := range task.Filters() {
			match, err := filter.Do(ctx)
			if err != nil {
				fmt.Fprintf(r.Out, "⛔️ Filter %s of task %s failed: %s\n", filter.String(), task.SourceTask().Name, err)
				continue
			}

			if match {
				fmt.Fprintf(r.Out, "✅ Filter %s of task %s matches\n", filter.String(), task.SourceTask().Name)
			} else {
				fmt.Fprintf(r.Out, "❌ Filter %s of task %s doesn't match\n", filter.String(), task.SourceTask().Name)
				matched = false
			}
		}

		if !matched {
			continue
		}

		fmt.Fprintf(r.Out, "🏗️ Cloning repository\n")
		checkoutPath, err := r.GitClient.Prepare(repository, false)
		if err != nil {
			fmt.Fprintf(r.Out, "⛔️ Failed to prepare repository %s: %s\n", repository.FullName(), err)
			continue
		}

		hasMergeConflict, err := r.GitClient.UpdateTaskBranch(task.BranchName(), false, repository)
		if err != nil {
			fmt.Fprintf(r.Out, "⛔️ Failed to prepare branch: %s\n", err)
			continue
		}

		if hasMergeConflict {
			fmt.Fprintf(r.Out, "⛔️ Merge conflict detected - view checkout in %s\n", checkoutPath)
			continue
		}

		fmt.Fprintf(r.Out, "🚜 Applying actions of task\n")
		ctx = context.WithValue(ctx, saturnContext.CheckoutPath{}, checkoutPath)
		err = r.ApplyActionsFunc(task.Actions(), ctx, checkoutPath)
		if err != nil {
			fmt.Fprintf(r.Out, "⛔️ %s\n", err)
			continue
		}

		result, err := r.GitClient.HasLocalChanges()
		if err != nil {
			fmt.Fprintf(r.Out, "⛔️ Check of local changes failed: %s\n", err)
			continue
		}

		if result {
			fmt.Fprintf(r.Out, "😍 Actions modified files - view checkout in %s\n", checkoutPath)
		} else {
			fmt.Fprintf(r.Out, "⚠️  No changes after applying actions - view checkout in %s\n", checkoutPath)
		}
	}

	if !processed {
		fmt.Fprintf(r.Out, "⛔️ Task %s not found in %s\n", r.TaskName, r.TaskFile)
	}

	return nil
}
