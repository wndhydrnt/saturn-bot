package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/config"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/mock"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"go.uber.org/mock/gomock"
)

var (
	tryTestOpts = options.Opts{
		ActionFactories: action.BuiltInFactories,
		FilterFactories: filter.BuiltInFactories,
	}
)

func TestTryRunner_Run_FilesModified(t *testing.T) {
	ctrl := gomock.NewController(t)
	repoMock := mock.NewMockRepository(ctrl)
	repoMock.EXPECT().Host().Return("git.local")
	repoMock.EXPECT().Owner().Return("unit")
	repoMock.EXPECT().Name().Return("test")
	hostMock := mock.NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName("git.local/unit/test").Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := mock.NewMockGitClient(ctrl)
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

	underTest := &TryRunner{
		applyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		gitc:             gitcMock,
		hosts:            []host.Host{hostMock},
		out:              out,
		registry:         registry,
		repositoryName:   "git.local/unit/test",
		taskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Filter repository(host=^git.local$,owner=^unit$,name=^test$) of task Unit Test matches")
	assert.Contains(t, out.String(), "Actions modified files")
}

func TestTryRunner_Run_NoChanges(t *testing.T) {
	repoName := "git.local/unit/test"
	ctrl := gomock.NewController(t)
	repoMock := mock.NewMockRepository(ctrl)
	repoMock.EXPECT().Host().Return("git.local")
	repoMock.EXPECT().Owner().Return("unit")
	repoMock.EXPECT().Name().Return("test")
	hostMock := mock.NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName(repoName).Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := mock.NewMockGitClient(ctrl)
	gitcMock.EXPECT().Prepare(repoMock, false).Return("/checkout", nil)
	gitcMock.EXPECT().UpdateTaskBranch("saturn-bot--unit-test", false, repoMock).Return(false, nil)
	gitcMock.EXPECT().HasLocalChanges().Return(false, nil)
	out := &bytes.Buffer{}
	taskFile := createTestTaskFile(createTestTask(repoName))

	underTest := &TryRunner{
		applyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		gitc:             gitcMock,
		hosts:            []host.Host{hostMock},
		out:              out,
		registry:         registry,
		repositoryName:   repoName,
		taskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "No changes after applying actions")
}

func TestTryRunner_Run_EmptyTaskFile(t *testing.T) {
	repoName := "git.local/unit/test"
	ctrl := gomock.NewController(t)
	repoMock := mock.NewMockRepository(ctrl)
	hostMock := mock.NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName(repoName).Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := mock.NewMockGitClient(ctrl)
	out := &bytes.Buffer{}
	taskFile := createTempFile("", "*.yaml")

	underTest := &TryRunner{
		applyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		gitc:             gitcMock,
		hosts:            []host.Host{hostMock},
		out:              out,
		registry:         registry,
		repositoryName:   repoName,
		taskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), ".yaml does not contain any tasks")
}

func TestTryRunner_Run_TaskName(t *testing.T) {
	repoName := "git.local/unit/test"
	ctrl := gomock.NewController(t)
	repoMock := mock.NewMockRepository(ctrl)
	hostMock := mock.NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName(repoName).Return(repoMock, nil)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := mock.NewMockGitClient(ctrl)
	out := &bytes.Buffer{}
	taskFile := createTestTaskFile(createTestTask(repoName))

	underTest := &TryRunner{
		applyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		gitc:             gitcMock,
		hosts:            []host.Host{hostMock},
		out:              out,
		registry:         registry,
		repositoryName:   repoName,
		taskFile:         taskFile,
		taskName:         "Other",
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Task Other not found in ")
}

func TestTryRunner_Run_FilterDoesNotMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	repoMock := mock.NewMockRepository(ctrl)
	repoMock.EXPECT().Host().Return("git.local")
	repoMock.EXPECT().Owner().Return("unit")
	repoMock.EXPECT().Name().Return("test")
	hostMock := mock.NewMockHost(ctrl)
	hostMock.EXPECT().CreateFromName("git.local/unit/test").Return(repoMock, nil)
	registry := task.NewRegistry(options.Opts{FilterFactories: filter.BuiltInFactories})
	gitcMock := mock.NewMockGitClient(ctrl)
	out := &bytes.Buffer{}
	content := `name: Unit Test
filters:
  - filter: repository
    params:
      host: git.local
      owner: unit
      name: no-match`
	taskFile := createTempFile(content, "*.yaml")

	underTest := &TryRunner{
		applyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		gitc:             gitcMock,
		hosts:            []host.Host{hostMock},
		out:              out,
		registry:         registry,
		repositoryName:   "git.local/unit/test",
		taskFile:         taskFile,
	}
	err := underTest.Run()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Filter repository(host=^git.local$,owner=^unit$,name=^no-match$) of task Unit Test doesn't match")
}

func TestTryRunner_Run_UnsetRepositoryName(t *testing.T) {
	ctrl := gomock.NewController(t)
	hostMock := mock.NewMockHost(ctrl)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := mock.NewMockGitClient(ctrl)
	taskFile := createTestTaskFile(createTestTask("git.local/unit/test"))

	underTest := &TryRunner{
		applyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		gitc:             gitcMock,
		hosts:            []host.Host{hostMock},
		out:              &bytes.Buffer{},
		registry:         registry,
		repositoryName:   "",
		taskFile:         taskFile,
	}
	err := underTest.Run()

	assert.EqualError(t, err, "required flag `--repository` is not set")
}

func TestTryRunner_Run_UnsetTaskFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	hostMock := mock.NewMockHost(ctrl)
	registry := task.NewRegistry(tryTestOpts)
	gitcMock := mock.NewMockGitClient(ctrl)

	underTest := &TryRunner{
		applyActionsFunc: func(actions []action.Action, ctx context.Context, dir string) error { return nil },
		gitc:             gitcMock,
		hosts:            []host.Host{hostMock},
		out:              &bytes.Buffer{},
		registry:         registry,
		repositoryName:   "git.local/unit/test",
		taskFile:         "",
	}
	err := underTest.Run()

	assert.EqualError(t, err, "required flag `--task-file` is not set")
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

	runner, err := NewTryRunner(opts, "", "git.local/unit/test", "task.yaml", "Unit Test")

	require.NoError(t, err)
	assert.NotNil(t, runner.applyActionsFunc)
	assert.Implements(t, (*git.GitClient)(nil), runner.gitc)
	assert.IsType(t, []host.Host{}, runner.hosts)
	assert.Equal(t, runner.out, os.Stdout)
	assert.IsType(t, &task.Registry{}, runner.registry)
	assert.Equal(t, "git.local/unit/test", runner.repositoryName)
	assert.Equal(t, "task.yaml", runner.taskFile)
	assert.Equal(t, "Unit Test", runner.taskName)
	assert.DirExists(t, filepath.Join(os.TempDir(), "saturn-bot"), "creates data directory because an empty value has been passed to NewTryRunner()")
}
