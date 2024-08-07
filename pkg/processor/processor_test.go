package processor_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/mock"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"go.uber.org/mock/gomock"
)

type trueFilter struct{}

func (t *trueFilter) Do(_ context.Context) (bool, error) {
	return true, nil
}

func (t *trueFilter) String() string {
	return "true"
}

type falseFilter struct{}

func (t *falseFilter) Do(_ context.Context) (bool, error) {
	return false, nil
}

func (t *falseFilter) String() string {
	return "false"
}

func setupRepoMock(ctrl *gomock.Controller) *mock.MockRepository {
	r := mock.NewMockRepository(ctrl)
	r.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	r.EXPECT().Host().Return("git.local").AnyTimes()
	r.EXPECT().Name().Return("test").AnyTimes()
	r.EXPECT().Owner().Return("unit").AnyTimes()
	r.EXPECT().WebUrl().Return("http://git.local/unit/test").AnyTimes()
	return r
}

func TestProcessor_Process_CreatePullRequestLocalChanges(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(nil, nil)
	repo.EXPECT().IsPullRequestClosed(nil).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(nil).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(nil).Return("").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(nil).Return(false).AnyTimes()
	repo.EXPECT().CreatePullRequest("saturn-bot--unittest", gomock.Any()).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("commit test").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(true, nil)
	gitc.EXPECT().Push("saturn-bot--unittest").Return(nil)
	tw := &task.Wrapper{Task: schema.Task{CommitMessage: "commit test", Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrCreated, result)
}

func TestProcessor_Process_CreatePullRequestRemoteChanges(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(nil, nil)
	repo.EXPECT().IsPullRequestClosed(nil).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(nil).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(nil).Return("").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(nil).Return(false).AnyTimes()
	repo.EXPECT().CreatePullRequest("saturn-bot--unittest", gomock.Any()).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("commit test").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: schema.Task{CommitMessage: "commit test", Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrCreated, result)
}

func TestProcessor_Process_PullRequestClosedAndMergeOnceActive(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(true)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/tmp", nil)
	tw := &task.Wrapper{Task: schema.Task{MergeOnce: true, Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrClosedBefore, result)
}

func TestProcessor_Process_PullRequestMergedAndMergeOnceActive(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(true)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/tmp", nil)
	tw := &task.Wrapper{Task: schema.Task{MergeOnce: true, Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrMergedBefore, result)
}

func TestProcessor_Process_CreateOnly(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/tmp", nil)
	tw := &task.Wrapper{Task: schema.Task{CreateOnly: true, Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrOpen, result)
}

func TestProcessor_Process_ClosePullRequestIfChangesExistInBaseBranch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().IsPullRequestOpen(prID).Return(true)
	repo.EXPECT().PullRequest(prID).Return(nil)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true)
	repo.EXPECT().ClosePullRequest("Everything up-to-date. Closing.", prID)
	repo.EXPECT().DeleteBranch(prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	tw := &task.Wrapper{Task: schema.Task{Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrClosed, result)
}

func TestProcessor_Process_MergePullRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().PullRequest(prID).Return(&host.PullRequest{Number: 579, WebURL: "https://git.localhost/unit/test"})
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(true, nil)
	repo.EXPECT().GetPullRequestCreationTime(prID).Return(time.Now().AddDate(0, 0, -1))
	repo.EXPECT().CanMergePullRequest(prID).Return(true, nil)
	repo.EXPECT().MergePullRequest(true, prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: schema.Task{AutoMerge: true, Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrMerged, result)
}

func TestProcessor_Process_MergePullRequest_FailedMergeChecks(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(false, nil)
	repo.EXPECT().PullRequest(prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: schema.Task{AutoMerge: true, Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultChecksFailed, result)
}

func TestProcessor_Process_MergePullRequest_AutoMergeAfter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(true, nil)
	repo.EXPECT().GetPullRequestCreationTime(prID).Return(time.Now().AddDate(0, 0, -1))
	repo.EXPECT().PullRequest(prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: schema.Task{
		AutoMerge:      true,
		AutoMergeAfter: "48h",
		Name:           "unittest",
	}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultAutoMergeTooEarly, result)
}

func TestProcessor_Process_MergePullRequest_MergeConflict(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(true, nil)
	repo.EXPECT().GetPullRequestCreationTime(prID).Return(time.Now().AddDate(0, 0, -1))
	repo.EXPECT().CanMergePullRequest(prID).Return(false, nil)
	repo.EXPECT().PullRequest(prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: schema.Task{
		AutoMerge: true,
		Name:      "unittest",
	}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultConflict, result)
}

func TestProcessor_Process_UpdatePullRequest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	autoMergeAfter := time.Duration(0)
	prData := host.PullRequestData{
		AutoMergeAfter: &autoMergeAfter,
		TaskName:       "unittest",
		TemplateData: map[string]any{
			"RepositoryFullName": "git.local/unit/test",
			"RepositoryHost":     "git.local",
			"RepositoryName":     "test",
			"RepositoryOwner":    "unit",
			"RepositoryWebUrl":   "http://git.local/unit/test",
			"TaskName":           "unittest",
			"Greeting":           "Hello",
		},
	}
	repo.EXPECT().UpdatePullRequest(prData, prID).Return(nil)
	repo.EXPECT().PullRequest(prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().Push("saturn-bot--unittest").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(true, nil)
	tw := &task.Wrapper{Task: schema.Task{Name: "unittest"}}
	tw.AddFilters(&trueFilter{})
	ctx := context.Background()
	ctx = context.WithValue(ctx, sContext.RunDataKey{}, map[string]string{"Greeting": "Hello", "TaskName": "other"})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(ctx, false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrOpen, result)
}

func TestProcessor_Process_NoChanges(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(false).AnyTimes()
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: schema.Task{Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultNoChanges, result)
}

func TestProcessor_Process_BranchModified(t *testing.T) {
	prCommentBody := `<!-- saturn-bot::{branch-modified} -->
:warning: **This pull request has been modified.**

This is a safety mechanism to prevent saturn-bot from accidentally overriding custom commits.

saturn-bot will not be able to resolve merge conflicts with ` + "`main`" + ` automatically.
It will not update this pull request or auto-merge it.

Check the box in the description of this PR to force a rebase. This will remove all commits not made by saturn-bot.

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
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().ListPullRequestComments(prID).Return([]host.PullRequestComment{}, nil)
	repo.EXPECT().CreatePullRequestComment(prCommentBody, prID).Return(nil)
	repo.EXPECT().PullRequest(prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().
		UpdateTaskBranch("saturn-bot--unittest", false, repo).
		Return(false, &git.BranchModifiedError{Checksums: []string{"abc", "def"}})
	tw := &task.Wrapper{Task: schema.Task{Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultBranchModified, result)
}

func TestProcessor_Process_ForceRebaseByUser(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("some text\n[x] If you want to rebase this PR\nsome text")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	prComment := host.PullRequestComment{Body: "<!-- saturn-bot::{branch-modified} -->\nsome text", ID: 123}
	repo.EXPECT().
		ListPullRequestComments(prID).
		Return([]host.PullRequestComment{prComment}, nil)
	repo.EXPECT().DeletePullRequestComment(prComment, prID).Return(nil)
	repo.EXPECT().UpdatePullRequest(gomock.AssignableToTypeOf(host.PullRequestData{}), prID).Return(nil)
	repo.EXPECT().PullRequest(prID).Return(nil)
	gitc := mock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", true, repo).Return(false, nil)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Wrapper{Task: schema.Task{Name: "unittest"}}
	tw.AddFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultPrOpen, result)
}

func TestProcessor_Process_ChangeLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	gitc := mock.NewMockGitClient(ctrl)
	repo := setupRepoMock(ctrl)
	tw := &task.Wrapper{Task: schema.Task{ChangeLimit: 1, Name: "unittest"}}
	tw.IncChangeLimitCount()

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultSkip, result)
}

func TestProcessor_Process_MaxOpenPRs(t *testing.T) {
	ctrl := gomock.NewController(t)
	gitc := mock.NewMockGitClient(ctrl)
	repo := setupRepoMock(ctrl)
	tw := &task.Wrapper{Task: schema.Task{MaxOpenPRs: 1, Name: "unittest"}}
	tw.IncOpenPRsCount()

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultSkip, result)
}

func TestProcessor_Process_FilterNotMatching(t *testing.T) {
	ctrl := gomock.NewController(t)
	gitc := mock.NewMockGitClient(ctrl)
	repo := setupRepoMock(ctrl)
	tw := &task.Wrapper{Task: schema.Task{Name: "unittest"}}
	tw.AddFilters(&falseFilter{})

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultNoMatch, result)
}

func TestProcessor_Process_NoFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	gitc := mock.NewMockGitClient(ctrl)
	repo := setupRepoMock(ctrl)
	tw := &task.Wrapper{Task: schema.Task{Name: "unittest"}}

	p := &processor.Processor{Git: gitc}
	result, err := p.Process(context.Background(), false, repo, tw, true)

	require.NoError(t, err)
	assert.Equal(t, processor.ResultNoMatch, result)
}
