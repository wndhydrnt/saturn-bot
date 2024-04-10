package pkg

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	proto "github.com/wndhydrnt/saturn-sync-go/protocol/v1"
	"github.com/wndhydrnt/saturn-sync/pkg/cache"
	"github.com/wndhydrnt/saturn-sync/pkg/git"
	"github.com/wndhydrnt/saturn-sync/pkg/host"
	"github.com/wndhydrnt/saturn-sync/pkg/mock"
	"github.com/wndhydrnt/saturn-sync/pkg/task"
	"go.uber.org/mock/gomock"
)

func setupRepoMock(ctrl *gomock.Controller) *mock.MockRepository {
	r := mock.NewMockRepository(ctrl)
	r.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	r.EXPECT().Host().Return("git.local").AnyTimes()
	r.EXPECT().Name().Return("test").AnyTimes()
	r.EXPECT().Owner().Return("unit").AnyTimes()
	r.EXPECT().WebUrl().Return("http://git.local/unit/test").AnyTimes()
	return r
}

func TestApplyTaskToRepository_CreatePullRequestLocalChanges(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(nil, nil)
	repo.EXPECT().IsPullRequestClosed(nil).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(nil).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(nil).Return("").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(nil).Return(false).AnyTimes()
	repo.EXPECT().CreatePullRequest("saturn-sync--unittest", gomock.Any()).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("commit test").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(true, nil)
	gitc.EXPECT().Push("saturn-sync--unittest").Return(nil)
	tw := &task.Wrapper{Task: &proto.Task{CommitMessage: ptr("commit test"), Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrCreated, result)
}

func TestApplyTaskToRepository_CreatePullRequestRemoteChanges(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(nil, nil)
	repo.EXPECT().IsPullRequestClosed(nil).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(nil).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(nil).Return("").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(nil).Return(false).AnyTimes()
	repo.EXPECT().CreatePullRequest("saturn-sync--unittest", gomock.Any()).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("commit test").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: &proto.Task{CommitMessage: ptr("commit test"), Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrCreated, result)
}

func TestApplyTaskToRepository_PullRequestClosedAndMergeOnceActive(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(true)
	gitc := mock.NewMockGitClient(ctrl)
	tw := &task.Wrapper{Task: &proto.Task{MergeOnce: ptr(true), Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, "")

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrClosedBefore, result)
}

func TestApplyTaskToRepository_PullRequestMergedAndMergeOnceActive(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(true)
	gitc := mock.NewMockGitClient(ctrl)
	tw := &task.Wrapper{Task: &proto.Task{MergeOnce: ptr(true), Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, "")

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrMergedBefore, result)
}

func TestApplyTaskToRepository_CreateOnly(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	gitc := mock.NewMockGitClient(ctrl)
	tw := &task.Wrapper{Task: &proto.Task{CreateOnly: ptr(true), Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, "")

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrOpen, result)
}

func TestApplyTaskToRepository_ClosePullRequestIfChangesExistInBaseBranch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true)
	repo.EXPECT().ClosePullRequest("Everything up-to-date. Closing.", prID)
	repo.EXPECT().DeleteBranch(prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	tw := &task.Wrapper{Task: &proto.Task{Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrClosed, result)
}

func TestApplyTaskToRepository_MergePullRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(true, nil)
	repo.EXPECT().GetPullRequestCreationTime(prID).Return(time.Now().AddDate(0, 0, -1))
	repo.EXPECT().CanMergePullRequest(prID).Return(true, nil)
	repo.EXPECT().MergePullRequest(true, prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: &proto.Task{AutoMerge: ptr(true), Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrMerged, result)
}

func TestApplyTaskToRepository_MergePullRequest_FailedMergeChecks(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(false, nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: &proto.Task{AutoMerge: ptr(true), Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultChecksFailed, result)
}

func TestApplyTaskToRepository_MergePullRequest_AutoMergeAfter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(true, nil)
	repo.EXPECT().GetPullRequestCreationTime(prID).Return(time.Now().AddDate(0, 0, -1))
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: &proto.Task{
		AutoMerge:             ptr(true),
		AutoMergeAfterSeconds: ptr(int32(172800)),
		Name:                  "unittest",
	}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultAutoMergeTooEarly, result)
}

func TestApplyTaskToRepository_MergePullRequest_MergeConflict(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(true, nil)
	repo.EXPECT().GetPullRequestCreationTime(prID).Return(time.Now().AddDate(0, 0, -1))
	repo.EXPECT().CanMergePullRequest(prID).Return(false, nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: &proto.Task{
		AutoMerge: ptr(true),
		Name:      "unittest",
	}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultConflict, result)
}

func TestApplyTaskToRepository_UpdatePullRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().UpdatePullRequest(gomock.AssignableToTypeOf(host.PullRequestData{}), prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().Push("saturn-sync--unittest").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(true, nil)
	tw := &task.Wrapper{Task: &proto.Task{Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrOpen, result)
}

func TestApplyTaskToRepository_NoChanges(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(false).AnyTimes()
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: &proto.Task{Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultNoChanges, result)
}

func TestApplyTaskToRepository_BranchModified(t *testing.T) {
	prCommentBody := `<!-- saturn-sync::{branch-modified} -->
:warning: **This pull request has been modified.**

This is a safety mechanism to prevent saturn-sync from accidentally overriding custom commits.

saturn-sync will not be able to resolve merge conflicts with ` + "`main`" + ` automatically.
It will not update this pull request or auto-merge it.

Check the box in the description of this PR to force a rebase. This will remove all commits not made by saturn-sync.

The commit(s) that modified the pull request:

- abc

- def

`

	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := mock.NewMockRepository(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().ListPullRequestComments(prID).Return([]host.PullRequestComment{}, nil)
	repo.EXPECT().FullName().Return("git.local/unit/test")
	repo.EXPECT().CreatePullRequestComment(prCommentBody, prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().
		UpdateTaskBranch("saturn-sync--unittest", false, repo).
		Return(false, &git.BranchModifiedError{Checksums: []string{"abc", "def"}})
	tw := &task.Wrapper{Task: &proto.Task{Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultBranchModified, result)
}

func TestApplyTaskToRepository_ForceRebaseByUser(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-sync--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("some text\n[x] If you want to rebase this PR\nsome text")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	prComment := host.PullRequestComment{Body: "<!-- saturn-sync::{branch-modified} -->\nsome text", ID: 123}
	repo.EXPECT().
		ListPullRequestComments(prID).
		Return([]host.PullRequestComment{prComment}, nil)
	repo.EXPECT().DeletePullRequestComment(prComment, prID).Return(nil)
	repo.EXPECT().UpdatePullRequest(gomock.AssignableToTypeOf(host.PullRequestData{}), prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().UpdateTaskBranch("saturn-sync--unittest", true, repo).Return(false, nil)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-sync--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: &proto.Task{Name: "unittest"}}

	result, err := applyTaskToRepository(context.Background(), false, gitc, slog.Default(), repo, tw, tempDir)

	require.NoError(t, err)
	assert.Equal(t, ApplyResultPrOpen, result)
}

type mockHost struct {
	repositories                     []host.Repository
	repositoriesWithOpenPullRequests []host.Repository
}

func (m *mockHost) CreateFromName(name string) (host.Repository, error) {
	return nil, fmt.Errorf("Not Implemented")
}

func (m *mockHost) ListRepositories(_ *time.Time, result chan []host.Repository, errChan chan error) {
	result <- m.repositories
	errChan <- nil
}

func (m *mockHost) ListRepositoriesWithOpenPullRequests(result chan []host.Repository, errChan chan error) {
	result <- m.repositoriesWithOpenPullRequests
	errChan <- nil
}

func mockApplyTaskFunc(ctx context.Context, dryRun bool, gitc git.GitClient, logger *slog.Logger, repo host.Repository, task task.Task, workDir string) (ApplyResult, error) {
	return ApplyResultNoChanges, nil
}

func createTestCache(taskFilePath string) string {
	f, err := os.Open(taskFilePath)
	if err != nil {
		panic(fmt.Sprintf("createTestCache: open task file: %s", err))
	}

	defer f.Close()
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		panic(fmt.Sprintf("createTestCache: calc checksum of task: %s", err))
	}

	checksum := fmt.Sprintf("%x", h.Sum(nil))
	tpl := `{
		"LastExecutionAt":1709382940609266,
		"Tasks":[{"Checksum":"%s","Name":"Unit Test"}]
	}`
	content := fmt.Sprintf(tpl, checksum)
	return createTempFile(content, "*.json")
}

func createTestTask(nameFilter string) string {
	tpl := `name: "Unit Test"
filters:
  repositoryNames:
    - names: ["%s"]
`
	content := fmt.Sprintf(tpl, nameFilter)
	return createTempFile(content, "*.yaml")
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
	repoWithPr := mock.NewMockRepository(ctrl)
	repoWithPr.EXPECT().FullName().Return("git.local/unittest/repoWithPr").AnyTimes()
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/work/repo", nil)
	gitc.EXPECT().Prepare(repoWithPr, false).Return("/work/repoWithPr", nil)
	hostm := &mockHost{
		repositories:                     []host.Repository{repo},
		repositoriesWithOpenPullRequests: []host.Repository{repoWithPr},
	}
	taskFile := createTestTask("git.local/unittest/repo.*")
	cacheFile := createTestCache(taskFile)
	defer func() {
		if err := os.Remove(cacheFile); err != nil {
			panic(err)
		}

		if err := os.Remove(taskFile); err != nil {
			panic(err)
		}
	}()

	runner := &executeRunner{
		applyTaskFunc: mockApplyTaskFunc,
		cache:         cache.NewJsonFile(cacheFile),
		dryRun:        false,
		git:           gitc,
		hosts:         []host.Host{hostm},
		taskRegistry:  &task.Registry{},
	}
	err := runner.run([]string{taskFile})

	require.NoError(t, err)
}

func ptr[T any](t T) *T {
	return &t
}
