package command_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"go.uber.org/mock/gomock"
)

var (
	tryTestOpts = options.Opts{
		ActionFactories: action.BuiltInFactories,
		FilterFactories: filter.BuiltInFactories,
	}
)

func setupTryRepoMock(ctrl *gomock.Controller) *MockRepository {
	hostDetailMock := NewMockHostDetail(ctrl)
	hostDetailMock.EXPECT().Name().Return("git.local").AnyTimes()
	repoMock := NewMockRepository(ctrl)
	repoMock.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	repoMock.EXPECT().Host().Return(hostDetailMock).AnyTimes()
	repoMock.EXPECT().Owner().Return("unit").AnyTimes()
	repoMock.EXPECT().Name().Return("test").AnyTimes()
	repoMock.EXPECT().WebUrl().Return("http://git.local/unit/test").AnyTimes()
	return repoMock
}

func TestTryRunner_Run_FilesModified(t *testing.T) {
	ctrl := gomock.NewController(t)
	repoMock := setupTryRepoMock(ctrl)
	hostMock := NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName("git.local/unit/test").Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := NewMockGitClient(ctrl)
	gitcMock.EXPECT().Prepare(repoMock, false).Return("/checkout", nil)
	gitcMock.EXPECT().UpdateTaskBranch("saturn-bot--unit-test", false, repoMock).Return(false, nil)
	gitcMock.EXPECT().HasLocalChanges().Return(true, nil)
	out := &bytes.Buffer{}
	content := `name: Unit Test
filters:
  - filter: repository
    params:
      host: git.local
      owner: unit
      name: test`
	taskFile := createTempFile(content, "*.yaml")

	underTest := &command.TryRunner{
		ApplyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		GitClient:        gitcMock,
		Hosts:            []host.Host{hostMock},
		Out:              out,
		Registry:         registry,
		RepositoryName:   "git.local/unit/test",
		TaskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Filter repository(host=^git.local$,owner=^unit$,name=^test$) of task Unit Test matches")
	assert.Contains(t, out.String(), "Actions modified files")
}

func TestTryRunner_Run_NoChanges(t *testing.T) {
	repoName := "git.local/unit/test"
	ctrl := gomock.NewController(t)
	repoMock := setupTryRepoMock(ctrl)
	hostMock := NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName(repoName).Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := NewMockGitClient(ctrl)
	gitcMock.EXPECT().Prepare(repoMock, false).Return("/checkout", nil)
	gitcMock.EXPECT().UpdateTaskBranch("saturn-bot--unit-test", false, repoMock).Return(false, nil)
	gitcMock.EXPECT().HasLocalChanges().Return(false, nil)
	out := &bytes.Buffer{}
	taskFile := createTestTaskFile(createTestTask(repoName))

	underTest := &command.TryRunner{
		ApplyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		GitClient:        gitcMock,
		Hosts:            []host.Host{hostMock},
		Out:              out,
		Registry:         registry,
		RepositoryName:   repoName,
		TaskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "No changes after applying actions")
}

func TestTryRunner_Run_EmptyTaskFile(t *testing.T) {
	repoName := "git.local/unit/test"
	ctrl := gomock.NewController(t)
	repoMock := NewMockRepository(ctrl)
	hostMock := NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName(repoName).Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := NewMockGitClient(ctrl)
	out := &bytes.Buffer{}
	taskFile := createTempFile("", "*.yaml")

	underTest := &command.TryRunner{
		ApplyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		GitClient:        gitcMock,
		Hosts:            []host.Host{hostMock},
		Out:              out,
		Registry:         registry,
		RepositoryName:   repoName,
		TaskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), ".yaml does not contain any tasks")
}

func TestTryRunner_Run_TaskName(t *testing.T) {
	repoName := "git.local/unit/test"
	ctrl := gomock.NewController(t)
	repoMock := NewMockRepository(ctrl)
	hostMock := NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName(repoName).Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := NewMockGitClient(ctrl)
	out := &bytes.Buffer{}
	taskFile := createTestTaskFile(createTestTask(repoName))

	underTest := &command.TryRunner{
		ApplyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		GitClient:        gitcMock,
		Hosts:            []host.Host{hostMock},
		Out:              out,
		Registry:         registry,
		RepositoryName:   repoName,
		TaskFile:         taskFile,
		TaskName:         "Other",
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Task Other not found in ")
}

func TestTryRunner_Run_FilterDoesNotMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	repoMock := setupTryRepoMock(ctrl)
	hostMock := NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName("git.local/unit/test").Return(repoMock, nil)
	registry := task.NewRegistry(options.Opts{FilterFactories: filter.BuiltInFactories})
	gitcMock := NewMockGitClient(ctrl)
	out := &bytes.Buffer{}
	content := `name: Unit Test
filters:
  - filter: repository
    params:
      host: git.local
      owner: unit
      name: no-match`
	taskFile := createTempFile(content, "*.yaml")

	underTest := &command.TryRunner{
		ApplyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		GitClient:        gitcMock,
		Hosts:            []host.Host{hostMock},
		Out:              out,
		Registry:         registry,
		RepositoryName:   "git.local/unit/test",
		TaskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Filter repository(host=^git.local$,owner=^unit$,name=^no-match$) of task Unit Test doesn't match")
}

func TestTryRunner_Run_UnsetRepositoryName(t *testing.T) {
	ctrl := gomock.NewController(t)
	hostMock := NewMockHost(ctrl)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := NewMockGitClient(ctrl)
	taskFile := createTestTaskFile(createTestTask("git.local/unit/test"))

	underTest := &command.TryRunner{
		ApplyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		GitClient:        gitcMock,
		Hosts:            []host.Host{hostMock},
		Out:              &bytes.Buffer{},
		Registry:         registry,
		RepositoryName:   "",
		TaskFile:         taskFile,
	}
	err := underTest.Run()

	assert.EqualError(t, err, "required flag `--repository` is not set")
}

func TestTryRunner_Run_UnsetTaskFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	hostMock := NewMockHost(ctrl)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := NewMockGitClient(ctrl)

	underTest := &command.TryRunner{
		ApplyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		GitClient:        gitcMock,
		Hosts:            []host.Host{hostMock},
		Out:              &bytes.Buffer{},
		Registry:         registry,
		RepositoryName:   "git.local/unit/test",
		TaskFile:         "",
	}
	err := underTest.Run()

	assert.EqualError(t, err, "required flag `--task-file` is not set")
}

func TestTryRunner_Run_InputsMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	repoMock := NewMockRepository(ctrl)
	hostMock := NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName("git.local/unit/test").Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := NewMockGitClient(ctrl)
	task := createTestTask("git.local/unit/test")
	task.Inputs = append(task.Inputs, schema.Input{
		Name: "test-input",
	})
	taskFile := createTestTaskFile(task)
	out := &bytes.Buffer{}

	underTest := &command.TryRunner{
		ApplyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		GitClient:        gitcMock,
		Hosts:            []host.Host{hostMock},
		Out:              out,
		Registry:         registry,
		RepositoryName:   "git.local/unit/test",
		TaskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err, "should not return an error")
	assert.Contains(t, out.String(), "Missing input: input test-input not set and has no default value", "should inform the user that the input isn't set")
}

func TestNewTryRunner(t *testing.T) {
	configRaw := `gitlabToken: "abc"
gitUserEmail: "test@unittest.local"
gitUserName: "unittest"`
	configFile := createTempFile(configRaw, "*.yaml")
	cfg, err := config.Read(configFile)
	require.NoError(t, err, "should parse configuration successfully")
	opts, err := options.ToOptions(cfg)
	require.NoError(t, err, "should convert configuration to options successfully")
	dataDir := filepath.Join(os.TempDir(), "saturn-bot")

	runner, err := command.NewTryRunner(opts, dataDir, "git.local/unit/test", "task.yaml", "Unit Test", map[string]string{})

	require.NoError(t, err)
	assert.NotNil(t, runner.ApplyActionsFunc)
	assert.Implements(t, (*git.GitClient)(nil), runner.GitClient)
	assert.IsType(t, []host.Host{}, runner.Hosts)
	assert.Equal(t, runner.Out, os.Stdout)
	assert.IsType(t, &task.Registry{}, runner.Registry)
	assert.Equal(t, "git.local/unit/test", runner.RepositoryName)
	assert.Equal(t, "task.yaml", runner.TaskFile)
	assert.Equal(t, "Unit Test", runner.TaskName)
	assert.DirExists(t, dataDir, "creates data directory because an empty value has been passed to NewTryRunner()")
	// Make sure to remove the directory to avoid false-positives
	err = os.RemoveAll(dataDir)
	require.NoError(t, err)
}
