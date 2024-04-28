package pkg

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	saturnContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
)

type TryRunner struct {
	applyActionsFunc func(actions []action.Action, ctx context.Context, dir string) error
	gitc             git.GitClient
	hosts            []host.Host
	out              io.Writer
	registry         *task.Registry
	repositoryName   string
	taskFile         string
	taskName         string
}

func NewTryRunner(configPath string, dataDir string, repositoryName string, taskFile string, taskName string) (*TryRunner, error) {
	cfg, err := config.Read(configPath)
	if err != nil {
		return nil, err
	}

	if dataDir == "" {
		dataDir = path.Join(os.TempDir(), "saturn-bot")
	}

	// This code sets its own data dir.
	cfg.DataDir = &dataDir
	err = initialize(cfg)
	if err != nil {
		return nil, err
	}

	hosts, err := createHostsFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create hosts from config: %w", err)
	}

	gitClient, err := git.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("new git client for try: %w", err)
	}

	return &TryRunner{
		applyActionsFunc: applyActionsInDirectory,
		gitc:             gitClient,
		hosts:            hosts,
		out:              os.Stdout,
		registry:         task.NewRegistry(),
		repositoryName:   repositoryName,
		taskFile:         taskFile,
		taskName:         taskName,
	}, nil
}

func (r *TryRunner) Run() error {
	if r.repositoryName == "" {
		return fmt.Errorf("required flag `--repository` is not set")
	}

	if r.taskFile == "" {
		return fmt.Errorf("required flag `--task-file` is not set")
	}

	var repository host.Repository
	for _, host := range r.hosts {
		var err error
		repository, err = host.CreateFromName(r.repositoryName)
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

	err := r.registry.ReadAll([]string{r.taskFile})
	if err != nil {
		return err
	}

	tasks := r.registry.GetTasks()
	if len(tasks) == 0 {
		fmt.Fprintf(r.out, "⛔️ File %s does not contain any tasks\n", r.taskFile)
		return nil
	}

	processed := false
	for _, task := range tasks {
		if r.taskName != "" && task.SourceTask().Name != r.taskName {
			continue
		}

		processed = true
		ctx := context.WithValue(context.Background(), saturnContext.RepositoryKey{}, repository)
		matched := true
		for _, filter := range task.Filters() {
			match, err := filter.Do(ctx)
			if err != nil {
				fmt.Fprintf(r.out, "⛔️ Filter %s of task %s failed: %s\n", filter.String(), task.SourceTask().Name, err)
				continue
			}

			if match {
				fmt.Fprintf(r.out, "✅ Filter %s of task %s matches\n", filter.String(), task.SourceTask().Name)
			} else {
				fmt.Fprintf(r.out, "❌ Filter %s of task %s doesn't match\n", filter.String(), task.SourceTask().Name)
				matched = false
			}
		}

		if !matched {
			continue
		}

		fmt.Fprintf(r.out, "🏗️ Cloning repository\n")
		checkoutPath, err := r.gitc.Prepare(repository, false)
		if err != nil {
			fmt.Fprintf(r.out, "⛔️ Failed to prepare repository %s: %s\n", repository.FullName(), err)
			continue
		}

		hasMergeConflict, err := r.gitc.UpdateTaskBranch(task.BranchName(), false, repository)
		if err != nil {
			fmt.Fprintf(r.out, "⛔️ Failed to prepare branch: %s\n", err)
			continue
		}

		if hasMergeConflict {
			fmt.Fprintf(r.out, "⛔️ Merge conflict detected - view checkout in %s\n", checkoutPath)
			continue
		}

		fmt.Fprintf(r.out, "🚜 Applying actions of task\n")
		ctx = context.WithValue(ctx, saturnContext.CheckoutPath{}, checkoutPath)
		err = r.applyActionsFunc(task.Actions(), ctx, checkoutPath)
		if err != nil {
			fmt.Fprintf(r.out, "⛔️ %s\n", err)
			continue
		}

		result, err := r.gitc.HasLocalChanges()
		if err != nil {
			fmt.Fprintf(r.out, "⛔️ Check of local changes failed: %s\n", err)
			continue
		}

		if result {
			fmt.Fprintf(r.out, "😍 Actions modified files - view checkout in %s\n", checkoutPath)
		} else {
			fmt.Fprintf(r.out, "⚠️  No changes after applying actions - view checkout in %s\n", checkoutPath)
		}
	}

	if !processed {
		fmt.Fprintf(r.out, "⛔️ Task %s not found in %s\n", r.taskName, r.taskFile)
	}

	return nil
}
