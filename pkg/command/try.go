package command

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/action"
	saturnContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
)

type TryRunner struct {
	ApplyActionsFunc func(actions []action.Action, ctx context.Context, dir string) error
	GitClient        git.GitClient
	Hosts            []host.Host
	Inputs           map[string]string
	Out              io.Writer
	Registry         *task.Registry
	RepositoryName   string
	TaskFile         string
	TaskName         string
}

func NewTryRunner(opts options.Opts, dataDir string, repositoryName string, taskFile string, taskName string, inputs map[string]string) (*TryRunner, error) {
	if dataDir != "" {
		// This code can set its own data dir.
		opts.Config.DataDir = &dataDir
	}
	err := options.Initialize(&opts)
	if err != nil {
		return nil, fmt.Errorf("initialize options: %w", err)
	}

	gitClient, err := git.New(opts)
	if err != nil {
		return nil, fmt.Errorf("new git client for try: %w", err)
	}

	return &TryRunner{
		ApplyActionsFunc: applyActionsInDirectory,
		GitClient:        gitClient,
		Hosts:            opts.Hosts,
		Inputs:           inputs,
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

	repository, err := host.NewRepositoryFromName(r.Hosts, r.RepositoryName)
	if err != nil {
		return err
	}

	err = r.Registry.ReadAll([]string{r.TaskFile})
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
		if r.TaskName != "" && task.Name != r.TaskName {
			continue
		}

		processed = true
		err := task.SetInputs(r.Inputs)
		if err != nil {
			fmt.Fprintf(r.Out, "⚠️  Missing input: %s\n", err)
			continue
		}

		ctx := context.WithValue(context.Background(), saturnContext.RepositoryKey{}, repository)
		matched := true
		for _, filter := range task.Filters() {
			match, err := filter.Do(ctx)
			if err != nil {
				fmt.Fprintf(r.Out, "⛔️ Filter %s of task %s failed: %s\n", filter.String(), task.Name, err)
				continue
			}

			if match {
				fmt.Fprintf(r.Out, "✅ Filter %s of task %s matches\n", filter.String(), task.Name)
			} else {
				fmt.Fprintf(r.Out, "❌ Filter %s of task %s doesn't match\n", filter.String(), task.Name)
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

		templateData := template.Data{
			Run: make(map[string]string),
			Repository: template.DataRepository{
				FullName: repository.FullName(),
				Host:     repository.Host().Name(),
				Name:     repository.Name(),
				Owner:    repository.Owner(),
				WebUrl:   repository.WebUrl(),
			},
			TaskName: task.Name,
		}
		branchName, err := task.RenderBranchName(templateData)
		if err != nil {
			fmt.Fprintf(r.Out, "⛔️ Failed to render branch name template %s: %s\n", task.BranchName, err)
			continue
		}

		hasMergeConflict, err := r.GitClient.UpdateTaskBranch(branchName, false, repository)
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
