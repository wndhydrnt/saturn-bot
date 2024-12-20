package command_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
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

type mockHost struct {
	repositories                     []host.Repository
	repositoriesWithOpenPullRequests []host.Repository
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
	return "mock"
}

func createTestCache(taskFilePath string) string {
	f, err := os.Open(taskFilePath)
	if err != nil {
		panic(fmt.Sprintf("createTestCache: open task file: %s", err))
	}

	// Decode and encode the task to align with what saturn-bot does
	defer f.Close()
	var tempTask schema.Task
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&tempTask); err != nil {
		panic(fmt.Sprintf("createTestCache: decode from YAML: %s", err))
	}

	h := sha256.New()
	enc := yaml.NewEncoder(h)
	if err := enc.Encode(&tempTask); err != nil {
		panic(fmt.Sprintf("createTestCache: encode to YAML: %s", err))
	}

	if err := enc.Close(); err != nil {
		panic(fmt.Sprintf("createTestCache: close YAML encoder: %s", err))
	}

	checksum := fmt.Sprintf("%x", h.Sum(nil))
	tpl := `{
		"LastExecutionAt":1709382940609266,
		"Tasks":[{"Checksum":"%s","Name":"Unit Test"}]
	}`
	content := fmt.Sprintf(tpl, checksum)
	return createTempFile(content, "*.json")
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
	repo.EXPECT().FullName().Return("git.local/unittest/" + name).AnyTimes()
	repo.EXPECT().Host().Return(hostDetailMock).AnyTimes()
	repo.EXPECT().Owner().Return("unittest").AnyTimes()
	repo.EXPECT().Name().Return(name).AnyTimes()
	return repo
}

func TestExecuteRunner_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := setupRunRepoMock(ctrl, "repo")
	repoWithPr := setupRunRepoMock(ctrl, "repoWithPr")
	hostm := &mockHost{
		repositories:                     []host.Repository{repo},
		repositoriesWithOpenPullRequests: []host.Repository{repoWithPr},
	}
	taskFile := createTestTaskFile(createTestTask("git.local/unittest/repo.*"))
	cacheFile := createTestCache(taskFile)
	cache := cache.NewJsonFile(cacheFile)
	_ = cache.Read()
	cacheLastExecutionBefore := cache.GetLastExecutionAt()
	defer func() {
		if err := os.Remove(cacheFile); err != nil {
			panic(err)
		}

		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask *task.Task = &task.Task{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repo, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil)
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repoWithPr, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil)

	defer gock.Off()
	gock.New("http://pgw.local").
		Put("/metrics/job/saturn-bot").
		Reply(200)
	pushGateway := push.New("http://pgw.local", "saturn-bot")

	runner := &command.Run{
		Cache:        cache,
		DryRun:       false,
		Hosts:        []host.Host{hostm},
		Processor:    procMock,
		PushGateway:  pushGateway,
		TaskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.Run([]string{}, []string{taskFile}, map[string]string{})

	require.NoError(t, err)
	assert.NotEqual(t, cacheLastExecutionBefore, cache.GetLastExecutionAt(), "Updates the lat execution time in the cache")
	require.True(t, gock.IsDone(), "All HTTP requests sent")
}

func TestExecuteRunner_Run_DryRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := setupRunRepoMock(ctrl, "repo")
	hostm := &mockHost{
		repositories: []host.Repository{repo},
	}
	taskFile := createTestTaskFile(createTestTask("git.local/unittest/repo.*"))
	cacheFile := createTestCache(taskFile)
	cache := cache.NewJsonFile(cacheFile)
	_ = cache.Read()
	cacheLastExecutionBefore := cache.GetLastExecutionAt()
	cacheTasks := cache.GetCachedTasks()
	defer func() {
		if err := os.Remove(cacheFile); err != nil {
			panic(err)
		}

		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask *task.Task = &task.Task{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), true, repo, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil)

	runner := &command.Run{
		Cache:        cache,
		DryRun:       true,
		Hosts:        []host.Host{hostm},
		Processor:    procMock,
		TaskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.Run([]string{}, []string{taskFile}, map[string]string{})

	require.NoError(t, err)
	assert.Equal(t, cacheLastExecutionBefore, cache.GetLastExecutionAt(), "Does not update the last execution time because dryRun is true")
	assert.Equal(t, cacheTasks, cache.GetCachedTasks(), "Does not update the cached tasks because dryRun is true")
}

func TestExecuteRunner_Run_RepositoriesCLI(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := setupRunRepoMock(ctrl, "repo")
	hostm := &mockHost{
		repositories: []host.Repository{repo},
	}
	taskFile := createTestTaskFile(createTestTask("git.local/unittest/repo.*"))
	cacheFile := createTestCache(taskFile)
	cache := cache.NewJsonFile(cacheFile)
	_ = cache.Read()
	cacheLastExecutionBefore := cache.GetLastExecutionAt()
	defer func() {
		if err := os.Remove(cacheFile); err != nil {
			panic(err)
		}

		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask *task.Task = &task.Task{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repo, gomock.AssignableToTypeOf(anyTask), false).
		Return(processor.ResultNoChanges, nil)

	runner := &command.Run{
		Cache:        cache,
		DryRun:       false,
		Hosts:        []host.Host{hostm},
		Processor:    procMock,
		TaskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.Run([]string{"git.local/unittest/repo"}, []string{taskFile}, map[string]string{})

	require.NoError(t, err)
	assert.NotEqual(t, cacheLastExecutionBefore, cache.GetLastExecutionAt(), "Updates the last execution time in the cache")
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

	cacheFile := createTestCache(taskOkFile)
	cache := cache.NewJsonFile(cacheFile)
	_ = cache.Read()
	defer func() {
		if err := os.Remove(cacheFile); err != nil {
			panic(err)
		}

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
		Return(processor.ResultNoChanges, nil)

	runner := &command.Run{
		Cache:        cache,
		DryRun:       false,
		Hosts:        []host.Host{hostm},
		Processor:    procMock,
		TaskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.Run([]string{"git.local/unittest/repo"}, []string{taskOkFile, taskMissingInputFile}, map[string]string{"test": "unit"})

	require.NoError(t, err)
}
