package processor_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/git"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/processor"
	"github.com/wndhydrnt/saturn-bot/pkg/task"
	"github.com/wndhydrnt/saturn-bot/pkg/task/schema"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
	gitmock "github.com/wndhydrnt/saturn-bot/test/mock/git"
	hostmock "github.com/wndhydrnt/saturn-bot/test/mock/host"
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

func setupRepoMock(ctrl *gomock.Controller) *hostmock.MockRepository {
	hostMock := hostmock.NewMockHostDetail(ctrl)
	hostMock.EXPECT().Name().Return("git.local").AnyTimes()
	r := hostmock.NewMockRepository(ctrl)
	r.EXPECT().FullName().Return("git.local/unit/test").AnyTimes()
	r.EXPECT().Host().Return(hostMock).AnyTimes()
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
	repo.EXPECT().PullRequest(nil).Return(nil).AnyTimes()
	repo.EXPECT().IsPullRequestClosed(nil).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(nil).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(nil).Return("").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(nil).Return(false).AnyTimes()
	prCreate := &host.PullRequest{Number: 1}
	repo.EXPECT().
		CreatePullRequest("saturn-bot--unittest", gomock.Any()).
		Return(prCreate, nil)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("commit test").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(true, nil)
	gitc.EXPECT().Push("saturn-bot--unittest").Return(nil)
	tw := &task.Task{Task: schema.Task{CommitMessage: "commit test", Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.Equal(t, processor.ResultPrCreated, results[0].Result)
	assert.Equal(t, prCreate, results[0].PullRequest)
	assert.NoError(t, results[0].Error)
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
	repo.EXPECT().PullRequest(nil).Return(nil).AnyTimes()
	repo.EXPECT().IsPullRequestClosed(nil).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(nil).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(nil).Return("").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(nil).Return(false).AnyTimes()
	prCreate := &host.PullRequest{Number: 1}
	repo.EXPECT().
		CreatePullRequest("saturn-bot--unittest", gomock.Any()).
		Return(prCreate, nil)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("commit test").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{CommitMessage: "commit test", Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrCreated, results[0].Result)
	assert.Equal(t, prCreate, results[0].PullRequest)
}

func TestProcessor_Process_CreatePullRequestPreviouslyClosed(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().PullRequest(prID).Return(nil).AnyTimes()
	repo.EXPECT().IsPullRequestClosed(prID).Return(true).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(nil).Return("").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(nil).Return(false).AnyTimes()
	prCreate := &host.PullRequest{Number: 1}
	repo.EXPECT().
		CreatePullRequest("saturn-bot--unittest", gomock.Any()).
		Return(prCreate, nil)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{CommitMessage: "commit test", Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrCreated, results[0].Result)
	assert.Equal(t, prCreate, results[0].PullRequest)
}

func TestProcessor_Process_PullRequestClosedAndMergeOnceActive(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().PullRequest(prID).Return(nil).AnyTimes()
	repo.EXPECT().IsPullRequestClosed(prID).Return(true)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/tmp", nil)
	tw := &task.Task{Task: schema.Task{MergeOnce: true, Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrClosedBefore, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}

func TestProcessor_Process_PullRequestMergedAndMergeOnceActive(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().PullRequest(prID).Return(nil).AnyTimes()
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(true)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/tmp", nil)
	tw := &task.Task{Task: schema.Task{MergeOnce: true, Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrMergedBefore, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}

func TestProcessor_Process_CreateOnly(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().PullRequest(prID).Return(nil).AnyTimes()
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/tmp", nil)
	tw := &task.Task{Task: schema.Task{CreateOnly: true, Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrOpen, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
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
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	tw := &task.Task{Task: schema.Task{Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrClosed, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
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
	pr := &host.PullRequest{Number: 579, WebURL: "https://git.localhost/unit/test"}
	repo.EXPECT().PullRequest(prID).Return(pr)
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().HasSuccessfulPullRequestBuild(prID).Return(true, nil)
	repo.EXPECT().GetPullRequestCreationTime(prID).Return(time.Now().AddDate(0, 0, -1))
	repo.EXPECT().CanMergePullRequest(prID).Return(true, nil)
	repo.EXPECT().MergePullRequest(true, prID).Return(nil)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{AutoMerge: true, Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrMerged, results[0].Result)
	assert.Equal(t, pr, results[0].PullRequest)
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
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{AutoMerge: true, Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultChecksFailed, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
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
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{
		AutoMerge:      true,
		AutoMergeAfter: "48h",
		Name:           "unittest",
	}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultAutoMergeTooEarly, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
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
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{
		AutoMerge: true,
		Name:      "unittest",
	}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultConflict, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
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
	prData := host.PullRequestData{
		Assignees: []string{"dina"},
		Reviewers: []string{"joel"},
		TaskName:  "unittest",
		TemplateData: template.Data{
			Run: map[string]string{
				"inputOne": "iValueOne",
				"inputTwo": "iValueTwo",
			},
			Repository: template.DataRepository{
				FullName: "git.local/unit/test",
				Host:     "git.local",
				Name:     "test",
				Owner:    "unit",
				WebUrl:   "http://git.local/unit/test",
			},
			TaskName: "unittest",
		},
		Title: "saturn-bot: task unittest",
	}
	repo.EXPECT().UpdatePullRequest(prData, prID).Return(nil)
	repo.EXPECT().PullRequest(prID).Return(nil).AnyTimes()
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().Push("saturn-bot--unittest").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(true, nil)
	tw := &task.Task{Task: schema.Task{
		Name: "unittest",
		Inputs: []schema.Input{
			{Name: "inputOne"},
			{Name: "inputTwo"},
		},
		Assignees: []string{"dina"},
		Reviewers: []string{"joel"},
	}}
	tw.AddPreCloneFilters(&trueFilter{})
	err = tw.SetInputs(map[string]string{"inputOne": "iValueOne", "inputTwo": "iValueTwo"})
	require.NoError(t, err)

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrOpen, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
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
	repo.EXPECT().PullRequest(prID).Return(nil).AnyTimes()
	repo.EXPECT().IsPullRequestClosed(prID).Return(false).AnyTimes()
	repo.EXPECT().IsPullRequestMerged(prID).Return(false).AnyTimes()
	repo.EXPECT().GetPullRequestBody(prID).Return("")
	repo.EXPECT().BaseBranch().Return("main")
	repo.EXPECT().IsPullRequestOpen(prID).Return(false).AnyTimes()
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultNoChanges, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
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
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().
		UpdateTaskBranch("saturn-bot--unittest", false, repo).
		Return(false, &git.BranchModifiedError{Checksums: []string{"abc", "def"}})
	tw := &task.Task{Task: schema.Task{Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultBranchModified, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
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
	repo.EXPECT().PullRequest(prID).Return(nil).AnyTimes()
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", true, repo).Return(false, nil)
	gitc.EXPECT().HasLocalChanges().Return(true, nil)
	gitc.EXPECT().CommitChanges("").Return(nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrOpen, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}

func TestProcessor_Process_ChangeLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	gitc := gitmock.NewMockGitClient(ctrl)
	repo := setupRepoMock(ctrl)
	tw := &task.Task{Task: schema.Task{ChangeLimit: 1, Name: "unittest"}}
	tw.IncChangeLimitCount()

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultSkip, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}

func TestProcessor_Process_MaxOpenPRs(t *testing.T) {
	ctrl := gomock.NewController(t)
	gitc := gitmock.NewMockGitClient(ctrl)
	repo := setupRepoMock(ctrl)
	tw := &task.Task{Task: schema.Task{MaxOpenPRs: 1, Name: "unittest"}}
	tw.IncOpenPRsCount()

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultSkip, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}

func TestProcessor_Process_FilterNotMatching(t *testing.T) {
	ctrl := gomock.NewController(t)
	gitc := gitmock.NewMockGitClient(ctrl)
	repo := setupRepoMock(ctrl)
	tw := &task.Task{Task: schema.Task{Name: "unittest"}}
	tw.AddPreCloneFilters(&falseFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultNoMatch, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}

func TestProcessor_Process_NoFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	gitc := gitmock.NewMockGitClient(ctrl)
	repo := setupRepoMock(ctrl)
	tw := &task.Task{Task: schema.Task{Name: "unittest"}}

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultNoMatch, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}

func TestProcessor_Process_AutoCloseAfter_Close(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().IsPullRequestOpen(prID).Return(true)
	createdAt := time.Now().Add(-1 * time.Hour)
	pr := &host.PullRequest{CreatedAt: &createdAt}
	repo.EXPECT().PullRequest(prID).Return(pr)
	msg := "Pull request has been open for longer than 30s. Closing automatically."
	repo.EXPECT().ClosePullRequest(msg, prID).Return(nil)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/tmp", nil)
	tw := &task.Task{Task: schema.Task{AutoCloseAfter: 30, Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrClosed, results[0].Result)
	assert.Equal(t, pr, results[0].PullRequest)
}

func TestProcessor_Process_AutoCloseAfter_NotTimeYet(t *testing.T) {
	prID := "prID"
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	repo.EXPECT().FindPullRequest("saturn-bot--unittest").Return(prID, nil)
	repo.EXPECT().IsPullRequestClosed(prID).Return(false)
	repo.EXPECT().IsPullRequestMerged(prID).Return(false)
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).Times(2)
	createdAt := time.Now().Add(-1 * time.Hour)
	pr := &host.PullRequest{CreatedAt: &createdAt}
	repo.EXPECT().PullRequest(prID).Return(pr)
	repo.EXPECT().GetPullRequestBody(prID).Return("").AnyTimes()
	repo.EXPECT().BaseBranch().Return("main").AnyTimes()
	repo.EXPECT().UpdatePullRequest(gomock.Any(), prID).Return(nil)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return("/tmp", nil)
	gitc.EXPECT().UpdateTaskBranch("saturn-bot--unittest", false, repo)
	gitc.EXPECT().HasLocalChanges().Return(false, nil)
	gitc.EXPECT().HasRemoteChanges("main").Return(true, nil)
	gitc.EXPECT().HasRemoteChanges("saturn-bot--unittest").Return(false, nil)
	tw := &task.Task{Task: schema.Task{AutoCloseAfter: 86_400, Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultPrOpen, results[0].Result)
	assert.Equal(t, pr, results[0].PullRequest)
}

func TestProcessor_Process_EmptyRepository(t *testing.T) {
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
	repo.EXPECT().IsPullRequestOpen(prID).Return(true).AnyTimes()
	repo.EXPECT().PullRequest(prID).Return(nil)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().Prepare(repo, false).Return(tempDir, nil)
	gitc.EXPECT().
		UpdateTaskBranch("saturn-bot--unittest", false, repo).
		Return(false, git.EmptyRepositoryError{})
	tw := &task.Task{Task: schema.Task{Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultNoMatch, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}

func TestProcessor_Process_CleanupOnPrepareError(t *testing.T) {
	tempDir := t.TempDir()
	ctrl := gomock.NewController(t)
	repo := setupRepoMock(ctrl)
	gitc := gitmock.NewMockGitClient(ctrl)
	gitc.EXPECT().
		Prepare(repo, false).
		Return(tempDir, fmt.Errorf("prepare error"))
	gitc.EXPECT().
		Cleanup(repo).
		Return(nil)

	tw := &task.Task{Task: schema.Task{Name: "unittest"}}
	tw.AddPreCloneFilters(&trueFilter{})

	p := &processor.Processor{Git: gitc}
	results := p.Process(false, repo, []*task.Task{tw}, true)

	assert.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, processor.ResultNoMatch, results[0].Result)
	assert.Nil(t, results[0].PullRequest)
}
