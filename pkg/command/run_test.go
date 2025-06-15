package command_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/action"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/command"
	"github.com/wndhydrnt/saturn-bot/pkg/filter"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	hostmock "github.com/wndhydrnt/saturn-bot/test/mock/host"
	processormock "github.com/wndhydrnt/saturn-bot/test/mock/processor"
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
	pullRequestIterator host.PullRequestIterator
	repositoryIterator  host.RepositoryIterator
	repositories        []host.Repository
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

	return nil, fmt.Errorf("repository not found")
}

func (m *mockHost) CreateFromName(name string) (host.Repository, error) {
	for _, repo := range m.repositories {
		if repo.FullName() == name {
			return repo, nil
		}
	}

	return nil, fmt.Errorf("Repository not found")
}

func (m *mockHost) AuthenticatedUser() (*host.UserInfo, error) {
	return nil, nil
}

func (m *mockHost) Name() string {
	return "git.local"
}

func (m *mockHost) Type() host.Type {
	return "mock"
}

func (m *mockHost) PullRequestFactory() host.PullRequestFactory {
	return func() any { return map[string]interface{}{} }
}

func (m *mockHost) PullRequestIterator() host.PullRequestIterator {
	return m.pullRequestIterator
}

func (m *mockHost) RepositoryIterator() host.RepositoryIterator {
	return m.repositoryIterator
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

func setupRunRepoMock(ctrl *gomock.Controller, name string) *hostmock.MockRepository {
	hostDetailMock := hostmock.NewMockHostDetail(ctrl)
	hostDetailMock.EXPECT().Name().Return("git.local").AnyTimes()
	repo := hostmock.NewMockRepository(ctrl)
	fullName := "git.local/unittest/" + name
	repo.EXPECT().FullName().Return(fullName).AnyTimes()
	repo.EXPECT().Host().Return(hostDetailMock).AnyTimes()
	repo.EXPECT().Owner().Return("unittest").AnyTimes()
	repo.EXPECT().Name().Return(name).AnyTimes()
	repo.EXPECT().Raw().Return(repoCacheItem{Name: fullName}).AnyTimes()
	repo.EXPECT().IsArchived().Return(false).AnyTimes()
	return repo
}

func setupRepositoryIteratorMock(ctrl *gomock.Controller, since *time.Time, listErr error, repos ...host.Repository) host.RepositoryIterator {
	iterFunc := func(yield func(request host.Repository) bool) {
		for _, repo := range repos {
			if !yield(repo) {
				return
			}
		}
	}

	repoIterMock := hostmock.NewMockRepositoryIterator(ctrl)
	repoIterMock.EXPECT().ListRepositories(since).Return(iterFunc)
	repoIterMock.EXPECT().Error().Return(listErr)
	return repoIterMock
}

func setupPullRequestIterator(ctrl *gomock.Controller, listErr error, prs ...*host.PullRequest) host.PullRequestIterator {
	iterFunc := func(yield func(request *host.PullRequest) bool) {
		for _, pr := range prs {
			if !yield(pr) {
				return
			}
		}
	}
	iterMock := hostmock.NewMockPullRequestIterator(ctrl)
	iterMock.EXPECT().ListPullRequests(nil).Return(iterFunc)
	iterMock.EXPECT().Error().Return(listErr)
	return iterMock
}

func setupCacher(t *testing.T) host.Cacher {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "cache.db")
	cacher, err := cache.New(dbPath)
	require.NoError(t, err)
	return cacher
}

func TestExecuteRunner_Run(t *testing.T) {
	tmpDir := t.TempDir()
	ctrl := gomock.NewController(t)
	repoOne := setupRunRepoMock(ctrl, "repoOne")
	repoTwo := setupRunRepoMock(ctrl, "repoTwo")
	repositoryIterMock := setupRepositoryIteratorMock(ctrl, nil, nil, repoOne, repoTwo)
	pr := &host.PullRequest{BranchName: "saturn-bot--unittest", RepositoryName: "unittest"}
	pullRequestIterMock := setupPullRequestIterator(ctrl, nil, pr)
	hostm := &mockHost{
		pullRequestIterator: pullRequestIterMock,
		repositoryIterator:  repositoryIterMock,
		repositories:        []host.Repository{repoOne, repoTwo},
	}
	testTask := createTestTask("git.local/unittest/repo.*")
	taskFile := createTestTaskFile(testTask)
	defer func() {
		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := processormock.NewMockRepositoryTaskProcessor(ctrl)
	anyTask := []*task.Task{}
	procMock.EXPECT().
		Process(false, repoOne, gomock.AssignableToTypeOf(anyTask), true).
		Return([]processor.ProcessResult{
			{Result: processor.ResultNoChanges, Task: &task.Task{Task: testTask}},
		})
	procMock.EXPECT().
		Process(false, repoTwo, gomock.AssignableToTypeOf(anyTask), true).
		Return([]processor.ProcessResult{
			{Result: processor.ResultNoChanges, Task: &task.Task{Task: testTask}},
		})

	prCacheMock := hostmock.NewMockPullRequestCache(ctrl)
	prCacheMock.EXPECT().LastUpdatedAtFor(hostm).Return(nil)
	prCacheMock.EXPECT().Get("saturn-bot--unittest", "unittest").Return(nil)
	prCacheMock.EXPECT().Set("saturn-bot--unittest", "unittest", pr)
	prCacheMock.EXPECT().SetLastUpdatedAtFor(hostm, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

	defer gock.Off()
	gock.New("http://pgw.local").
		Put("/metrics/job/saturn-bot").
		Reply(200)
	pushGateway := push.New("http://pgw.local", "saturn-bot")
	taskRegistry := task.NewRegistry(runTestOpts)

	runner := &command.Run{
		Clock:            clock.NewFakeDefault(),
		DryRun:           false,
		Hosts:            []host.Host{hostm},
		Processor:        procMock,
		PullRequestCache: prCacheMock,
		PushGateway:      pushGateway,
		RepositoryLister: host.NewRepositoryCache(setupCacher(t), clock.Default, filepath.Join(tmpDir, "cache"), 0),
		TaskRegistry:     taskRegistry,
	}
	_, err := runner.Run([]string{}, []string{taskFile}, map[string]string{})

	require.NoError(t, err)
	require.True(t, gock.IsDone(), "All HTTP requests sent")
}

func TestExecuteRunner_Run_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	ctrl := gomock.NewController(t)
	repo := setupRunRepoMock(ctrl, "repo")
	repoIterator := setupRepositoryIteratorMock(ctrl, nil, nil, repo)
	hostm := &mockHost{
		repositories:       []host.Repository{repo},
		repositoryIterator: repoIterator,
	}
	testTask := createTestTask("git.local/unittest/repo.*")
	taskFile := createTestTaskFile(createTestTask("git.local/unittest/repo.*"))
	defer func() {
		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := processormock.NewMockRepositoryTaskProcessor(ctrl)
	anyTask := []*task.Task{}
	procMock.EXPECT().
		Process(true, repo, gomock.AssignableToTypeOf(anyTask), true).
		Return([]processor.ProcessResult{
			{Result: processor.ResultNoChanges, Task: &task.Task{Task: testTask}},
		})
	taskRegistry := task.NewRegistry(runTestOpts)

	runner := &command.Run{
		DryRun:           true,
		Hosts:            []host.Host{hostm},
		Processor:        procMock,
		RepositoryLister: host.NewRepositoryCache(setupCacher(t), clock.Default, filepath.Join(tmpDir, "cache"), 0),
		TaskRegistry:     taskRegistry,
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
	testTask := createTestTask("git.local/unittest/repo.*")
	taskFile := createTestTaskFile(testTask)
	defer func() {
		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()
	procMock := processormock.NewMockRepositoryTaskProcessor(ctrl)
	anyTask := []*task.Task{}
	procMock.EXPECT().
		Process(false, repo, gomock.AssignableToTypeOf(anyTask), false).
		Return([]processor.ProcessResult{
			{Result: processor.ResultNoChanges, Task: &task.Task{Task: testTask}},
		})
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
	procMock := processormock.NewMockRepositoryTaskProcessor(ctrl)

	// Verifies that procMock.Process() gets called with the expected value.
	isTask := func(x any) bool {
		t := x.([]*task.Task)
		return t[0].Name == "TaskOk"
	}

	procMock.EXPECT().
		Process(false, repo, gomock.Cond(isTask), false).
		Return([]processor.ProcessResult{
			{Result: processor.ResultNoChanges, Task: &task.Task{Task: taskOk}},
		})

	runner := &command.Run{
		DryRun:       false,
		Hosts:        []host.Host{hostm},
		Processor:    procMock,
		TaskRegistry: task.NewRegistry(runTestOpts),
	}
	_, err := runner.Run([]string{"git.local/unittest/repo"}, []string{taskOkFile, taskMissingInputFile}, map[string]string{"test": "unit"})

	require.NoError(t, err)
}
