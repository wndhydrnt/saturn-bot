package command

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/mock"
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

func TestExecuteRunner_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().FullName().Return("git.local/unittest/repo").AnyTimes()
	repo.EXPECT().Host().Return("git.local").AnyTimes()
	repo.EXPECT().Owner().Return("unittest").AnyTimes()
	repo.EXPECT().Name().Return("repo").AnyTimes()
	repoWithPr := mock.NewMockRepository(ctrl)
	repoWithPr.EXPECT().FullName().Return("git.local/unittest/repoWithPr").AnyTimes()
	repoWithPr.EXPECT().Host().Return("git.local").AnyTimes()
	repoWithPr.EXPECT().Owner().Return("unittest").AnyTimes()
	repoWithPr.EXPECT().Name().Return("repoWithPr").AnyTimes()
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
	procMock := mock.NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask task.Task = &task.Wrapper{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repo, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil)
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repoWithPr, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil)

	runner := &run{
		cache:        cache,
		dryRun:       false,
		hosts:        []host.Host{hostm},
		processor:    procMock,
		taskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.run(map[string]string{}, []string{}, []string{taskFile})

	require.NoError(t, err)
	assert.NotEqual(t, cacheLastExecutionBefore, cache.GetLastExecutionAt(), "Updates the lat execution time in the cache")
}

func TestExecuteRunner_Run_DryRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().FullName().Return("git.local/unittest/repo").AnyTimes()
	repo.EXPECT().Host().Return("git.local").AnyTimes()
	repo.EXPECT().Owner().Return("unittest").AnyTimes()
	repo.EXPECT().Name().Return("repo").AnyTimes()
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
	procMock := mock.NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask task.Task = &task.Wrapper{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), true, repo, gomock.AssignableToTypeOf(anyTask), true).
		Return(processor.ResultNoChanges, nil)

	runner := &run{
		cache:        cache,
		dryRun:       true,
		hosts:        []host.Host{hostm},
		processor:    procMock,
		taskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.run(map[string]string{}, []string{}, []string{taskFile})

	require.NoError(t, err)
	assert.Equal(t, cacheLastExecutionBefore, cache.GetLastExecutionAt(), "Does not update the last execution time because dryRun is true")
	assert.Equal(t, cacheTasks, cache.GetCachedTasks(), "Does not update the cached tasks because dryRun is true")
}

func TestExecuteRunner_Run_RepositoriesCLI(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().FullName().Return("git.local/unittest/repo").AnyTimes()
	repo.EXPECT().Host().Return("git.local").AnyTimes()
	repo.EXPECT().Owner().Return("unittest").AnyTimes()
	repo.EXPECT().Name().Return("repo").AnyTimes()
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
	procMock := mock.NewMockRepositoryTaskProcessor(ctrl)
	var ctx = reflect.TypeOf((*context.Context)(nil)).Elem()
	var anyTask task.Task = &task.Wrapper{}
	procMock.EXPECT().
		Process(gomock.AssignableToTypeOf(ctx), false, repo, gomock.AssignableToTypeOf(anyTask), false).
		Return(processor.ResultNoChanges, nil)

	runner := &run{
		cache:        cache,
		dryRun:       false,
		hosts:        []host.Host{hostm},
		processor:    procMock,
		taskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.run(map[string]string{}, []string{"git.local/unittest/repo"}, []string{taskFile})

	require.NoError(t, err)
	assert.NotEqual(t, cacheLastExecutionBefore, cache.GetLastExecutionAt(), "Updates the lat execution time in the cache")
}
