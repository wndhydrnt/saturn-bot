package command_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
)

var (
	runTestOpts = options.Opts{
		ActionFactories: action.BuiltInFactories,
		FilterFactories: filter.BuiltInFactories,
	}
)

type repoCacheItem struct {
	Name string
}

type mockHost struct {
	repositories                     []host.Repository
	repositoriesWithOpenPullRequests []host.Repository
}

func (m *mockHost) CreateFromJson(dec *json.Decoder) (host.Repository, error) {
	item := &repoCacheItem{}
	err := dec.Decode(item)
	if err != nil {
		return nil, fmt.Errorf("host mock: decode cache item: %w", err)
	}

	for _, repo := range m.repositories {
		if repo.FullName() == item.Name {
			return repo, nil
		}
	}

	return nil, fmt.Errorf("Repository not found")
}

func (m *mockHost) CreateFromName(name string) (host.Repository, error) {
	for _, repo := range m.repositories {
		if repo.FullName() == name {
			return repo, nil
		}
	}

	return nil, fmt.Errorf("Repository not found")
}

func (m *mockHost) ListRepositories(_ *time.Time, result chan []host.Repository, errChan chan error) {
	result <- m.repositories
	errChan <- nil
}

func (m *mockHost) ListRepositoriesWithOpenPullRequests(result chan []host.Repository, errChan chan error) {
	result <- m.repositoriesWithOpenPullRequests
	errChan <- nil
}

func (m *mockHost) AuthenticatedUser() (*host.UserInfo, error) {
	return nil, nil
}

func (m *mockHost) Name() string {
	return "git.local"
}

func createTestTask(nameFilter string) schema.Task {
	parts := strings.Split(nameFilter, "/")
	return schema.Task{
		Name: "Unit Test",
		Filters: []schema.Filter{
			{
				Filter: "repository",
				Params: schema.FilterParams{
					"host":  parts[0],
					"owner": parts[1],
					"name":  parts[2],
				},
			},
		},
	}
}

func createTestTaskFile(t schema.Task) string {
	b := &bytes.Buffer{}
	dec := yaml.NewEncoder(b)
	err := dec.Encode(&t)
	if err != nil {
		panic(err)
	}

	return createTempFile(b.String(), "*.yaml")
}

func createTempFile(content string, pattern string) string {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(content)
	if err != nil {
		panic(err)
	}

	return f.Name()
}

func setupRunRepoMock(ctrl *gomock.Controller, name string) *MockRepository {
	hostDetailMock := NewMockHostDetail(ctrl)
	hostDetailMock.EXPECT().Name().Return("git.local").AnyTimes()
	repo := NewMockRepository(ctrl)
	fullName := "git.local/unittest/" + name
	repo.EXPECT().FullName().Return(fullName).AnyTimes()
	repo.EXPECT().Host().Return(hostDetailMock).AnyTimes()
	repo.EXPECT().Owner().Return("unittest").AnyTimes()
	repo.EXPECT().Name().Return(name).AnyTimes()
	repo.EXPECT().Raw().Return(repoCacheItem{Name: fullName}).AnyTimes()
	return repo
}

func TestExecuteRunner_Run(t *testing.T) {
	tmpDir := t.TempDir()
	ctrl := gomock.NewController(t)
	repoOne := setupRunRepoMock(ctrl, "repoOne")
	repoTwo := setupRunRepoMock(ctrl, "repoTwo")
	hostm := &mockHost{
		repositories: []host.Repository{repoOne, repoTwo},
	}
	taskFile := createTestTaskFile(createTestTask("git.local/unittest/repo.*"))
	defer func() {
		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask *task.Task = &task.Task{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repoOne, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil, nil)
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repoTwo, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil, nil)

	defer gock.Off()
	gock.New("http://pgw.local").
		Put("/metrics/job/saturn-bot").
		Reply(200)
	pushGateway := push.New("http://pgw.local", "saturn-bot")
	taskRegistry := task.NewRegistry(runTestOpts)

	runner := &command.Run{
		DryRun:      false,
		Hosts:       []host.Host{hostm},
		Processor:   procMock,
		PushGateway: pushGateway,
		RepositoryCache: &host.RepositoryFileCache{
			Clock: clock.Default,
			Dir:   filepath.Join(tmpDir, "cache"),
		},
		TaskRegistry: taskRegistry,
	}
	_, err := runner.Run([]string{}, []string{taskFile}, map[string]string{})

	require.NoError(t, err)
	require.True(t, gock.IsDone(), "All HTTP requests sent")
}

func TestExecuteRunner_Run_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	ctrl := gomock.NewController(t)
	repo := setupRunRepoMock(ctrl, "repo")
	hostm := &mockHost{
		repositories: []host.Repository{repo},
	}
	taskFile := createTestTaskFile(createTestTask("git.local/unittest/repo.*"))
	defer func() {
		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask *task.Task = &task.Task{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), true, repo, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil, nil)
	taskRegistry := task.NewRegistry(runTestOpts)

	runner := &command.Run{
		DryRun:    true,
		Hosts:     []host.Host{hostm},
		Processor: procMock,
		RepositoryCache: &host.RepositoryFileCache{
			Clock: clock.Default,
			Dir:   filepath.Join(tmpDir, "cache"),
		},
		TaskRegistry: taskRegistry,
	}
	_, err := runner.Run([]string{}, []string{taskFile}, map[string]string{})

	require.NoError(t, err)
}

func TestExecuteRunner_Run_RepositoriesCLI(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := setupRunRepoMock(ctrl, "repo")
	hostm := &mockHost{
		repositories: []host.Repository{repo},
	}
	taskFile := createTestTaskFile(createTestTask("git.local/unittest/repo.*"))
	defer func() {
		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask *task.Task = &task.Task{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repo, gomock.AssignableToTypeOf(anyTask), false).
		Return(processor.ResultNoChanges, nil, nil)
	taskRegistry := task.NewRegistry(runTestOpts)

	runner := &command.Run{
		DryRun:       false,
		Hosts:        []host.Host{hostm},
		Processor:    procMock,
		TaskRegistry: taskRegistry,
	}
	_, err := runner.Run([]string{"git.local/unittest/repo"}, []string{taskFile}, map[string]string{})

	require.NoError(t, err)
}

func TestExecuteRunner_Run_Inputs(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := setupRunRepoMock(ctrl, "repo")
	hostm := &mockHost{
		repositories: []host.Repository{repo},
	}

	taskOk := createTestTask("git.local/unittest/repo.*")
	taskOk.Inputs = []schema.Input{
		{Name: "test"},
	}
	taskOk.Name = "TaskOk"
	taskOkFile := createTestTaskFile(taskOk)

	taskMissingInput := createTestTask("git.local/unittest/repo.*")
	taskMissingInput.Inputs = []schema.Input{
		{Name: "other"},
	}
	taskOk.Name = "TaskMissingInput"
	taskMissingInputFile := createTestTaskFile(taskMissingInput)

	defer func() {
		if err := os.Remove(taskOkFile); err != nil {
			panic(err)
		}
	}()
	procMock := NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()

	// Verifies that procMock.Process() gets called with the expected value.
	isTask := func(x any) bool {
		t := x.(*task.Task)
		return t.Name == "TaskOk"
	}

	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repo, gomock.Cond(isTask), false).
		Return(processor.ResultNoChanges, nil, nil)

	runner := &command.Run{
		DryRun:       false,
		Hosts:        []host.Host{hostm},
		Processor:    procMock,
		TaskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.Run([]string{"git.local/unittest/repo"}, []string{taskOkFile, taskMissingInputFile}, map[string]string{"test": "unit"})

	require.NoError(t, err)
}
