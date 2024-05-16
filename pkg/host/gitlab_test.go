package host

import (
	"errors"
	"net/http"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

const (
	gitlabMergeRequestBody = "Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n"
)

func TestGitLabRepository_GetBaseBranch(t *testing.T) {
	project := &gitlab.Project{DefaultBranch: "main"}

	underTest := &GitLabRepository{project: project}

	require.Equal(t, "main", underTest.BaseBranch())
}

func TestGitLabRepository_GetCloneURL(t *testing.T) {
	project := &gitlab.Project{HTTPURLToRepo: "https://gitlab.com/wandhydrant/unittest.git"}

	underTest := &GitLabRepository{project: project}

	require.Equal(t, "https://gitlab.com/wandhydrant/unittest.git", underTest.CloneUrlHttp())
}

func TestGitLabRepository_GetFullName(t *testing.T) {
	project := &gitlab.Project{WebURL: "https://gitlab.com/wandhydrant/unittest"}

	underTest := &GitLabRepository{project: project}

	require.Equal(t, "gitlab.com/wandhydrant/unittest", underTest.FullName())
}

func TestGitLabRepository_GetPullRequestBody(t *testing.T) {
	mr := &gitlab.MergeRequest{Description: "Unit Test"}

	underTest := &GitLabRepository{}

	require.Equal(t, "Unit Test", underTest.GetPullRequestBody(mr))
}

func TestGitLabRepository_GetPullRequestCreationTime(t *testing.T) {
	createdAt := time.Now()
	mr := &gitlab.MergeRequest{CreatedAt: &createdAt}

	underTest := &GitLabRepository{}

	require.Equal(t, createdAt, underTest.GetPullRequestCreationTime(mr))
}

func TestGitLabRepository_CanMergePullRequest(t *testing.T) {
	mr := &gitlab.MergeRequest{}

	underTest := &GitLabRepository{}

	result, err := underTest.CanMergePullRequest(mr)
	require.NoError(t, err)
	require.True(t, result)
}

func TestGitLabRepository_ClosePullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests/987/notes").
		MatchType("json").
		JSON(map[string]string{"body": "Unit Test"}).
		Reply(200).
		JSON(map[string]string{})
	gock.New("http://gitlab.local").
		Put("/api/v4/projects/123/merge_requests/987").
		MatchType("json").
		JSON(map[string]string{"state_event": "close"}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{IID: 987}

	underTest := &GitLabRepository{client: setupClient(), project: project}

	err := underTest.ClosePullRequest("Unit Test", mr)
	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequestComment(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests/987/notes").
		MatchType("json").
		JSON(map[string]string{"body": "Unit Test"}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{IID: 987}

	underTest := &GitLabRepository{client: setupClient(), project: project}

	err := underTest.CreatePullRequestComment("Unit Test", mr)
	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests").
		MatchType("json").
		JSON(map[string]any{
			"title":         "Unit Test Title",
			"description":   "Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n",
			"source_branch": "saturn-bot--unit-test",
			"target_branch": "main",
			"assignee_ids":  nil,
		}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}
	prData := PullRequestData{Body: "Unit Test Body", Title: "Unit Test Title"}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.CreatePullRequest("saturn-bot--unit-test", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequest_WithAssignees(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "jane").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 975},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "joe").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 357},
		})
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests").
		MatchType("json").
		JSON(map[string]any{
			"title":         "Unit Test Title",
			"description":   "Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n",
			"source_branch": "saturn-bot--unit-test",
			"target_branch": "main",
			"assignee_ids":  []int{975, 357},
		}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}
	prData := PullRequestData{Assignees: []string{"jane", "joe"}, Body: "Unit Test Body", Title: "Unit Test Title"}

	client := setupClient()
	uc := &userCache{
		client: client,
		data:   map[string]*gitlab.User{},
	}
	underTest := &GitLabRepository{client: client, project: project, userCache: uc}
	err := underTest.CreatePullRequest("saturn-bot--unit-test", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequest_WithLabels(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests").
		MatchType("json").
		JSON(map[string]any{
			"title":         "Unit Test Title",
			"description":   "Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n",
			"labels":        "unit,test",
			"source_branch": "saturn-bot--unit-test",
			"target_branch": "main",
			"assignee_ids":  nil,
		}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}
	prData := PullRequestData{Body: "Unit Test Body", Labels: []string{"unit", "test"}, Title: "Unit Test Title"}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.CreatePullRequest("saturn-bot--unit-test", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_DeleteBranch(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Delete("/api/v4/projects/123/repository/branches/saturn-bot--unit-test").
		Reply(200)
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{IID: 987, SourceBranch: "saturn-bot--unit-test"}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.DeleteBranch(mr)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_DeleteBranch_NoDeleteIfGitLabDeletesMR(t *testing.T) {
	defer gock.Off()
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{IID: 987, SourceBranch: "saturn-bot--unit-test", ShouldRemoveSourceBranch: true}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.DeleteBranch(mr)

	require.NoError(t, err)
	require.True(t, gock.IsDone(), "not all mocks of gock have been called")
}

func TestGitLabRepository_DeletePullRequestComment(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Delete("/api/v4/projects/123/merge_requests/987/notes/456").
		Reply(200).
		JSON(map[string]string{})
	comment := PullRequestComment{ID: 456}
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{IID: 987}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.DeletePullRequestComment(comment, mr)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_FindPullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/merge_requests").
		MatchParams(map[string]string{"source_branch": "saturn-bot--unit-test", "state": "all"}).
		Reply(200).
		JSON([]*gitlab.MergeRequest{
			{SourceBranch: "saturn-bot--unit-test"},
		})
	project := &gitlab.Project{ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.FindPullRequest("saturn-bot--unit-test")

	require.NoError(t, err)
	require.IsType(t, &gitlab.MergeRequest{}, result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_FindPullRequest_NotFound(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/merge_requests").
		MatchParams(map[string]string{"source_branch": "saturn-bot--unit-test", "state": "all"}).
		Reply(200).
		JSON([]*gitlab.MergeRequest{})
	project := &gitlab.Project{ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	_, err := underTest.FindPullRequest("saturn-bot--unit-test")

	require.ErrorIs(t, err, ErrPullRequestNotFound)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_GetFile(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/repository/files/src/test.txt").
		MatchParam("ref", "main").
		Reply(200).
		JSON(&gitlab.File{
			Content: "Unit Test",
		})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.GetFile("src/test.txt")

	require.NoError(t, err)
	require.Equal(t, "Unit Test", result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_GetFile_base64(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/repository/files/src/test.txt").
		MatchParam("ref", "main").
		Reply(200).
		JSON(&gitlab.File{
			Content:  "VW5pdCBUZXN0",
			Encoding: "base64",
		})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.GetFile("src/test.txt")

	require.NoError(t, err)
	require.Equal(t, "Unit Test", result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_GetFile_NotFound(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/repository/files/src/test.txt").
		MatchParam("ref", "main").
		Reply(404)
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	_, err := underTest.GetFile("src/test.txt")

	require.ErrorIs(t, err, ErrFileNotFound)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_HasFile(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/repository/tree").
		MatchParam("path", "src").
		Reply(200).
		JSON([]*gitlab.TreeNode{
			{Type: "other", Path: "src/other.txt"},
			{Type: "blob", Path: "src/test.yaml"},
			{Type: "blob", Path: "src/test.txt"},
		})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.HasFile("src/*.txt")

	require.NoError(t, err)
	require.True(t, result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_HasFile_NoMatch(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/repository/tree").
		MatchParam("path", "src").
		Reply(200).
		JSON([]*gitlab.TreeNode{
			{Type: "blob", Path: "src/test.yaml"},
			{Type: "blob", Path: "src/test.txt"},
		})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.HasFile("src/test.json")

	require.NoError(t, err)
	require.False(t, result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_HasFile_EmptyRepository(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/repository/tree").
		MatchParam("path", ".").
		Reply(404)
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.HasFile("test.txt")

	require.NoError(t, err)
	require.False(t, result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_HasSuccessfulPullRequestBuild(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/merge_requests/987/approval_state").
		Reply(200).
		JSON(&gitlab.MergeRequestApprovalState{
			Rules: []*gitlab.MergeRequestApprovalRule{
				{Approved: true},
			},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/repository/commits/f7g8/statuses").
		Reply(200).
		JSON([]*gitlab.CommitStatus{
			{AllowFailure: true, Status: "failed"},
			{AllowFailure: false, Status: "success"},
		})
	mr := &gitlab.MergeRequest{IID: 987, SHA: "f7g8"}
	project := &gitlab.Project{ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.HasSuccessfulPullRequestBuild(mr)

	require.NoError(t, err)
	require.True(t, result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_HasSuccessfulPullRequestBuild_RuleNotApproved(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/merge_requests/987/approval_state").
		Reply(200).
		JSON(&gitlab.MergeRequestApprovalState{
			Rules: []*gitlab.MergeRequestApprovalRule{
				{Approved: true},
				{Approved: false},
			},
		})
	mr := &gitlab.MergeRequest{IID: 987}
	project := &gitlab.Project{ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.HasSuccessfulPullRequestBuild(mr)

	require.NoError(t, err)
	require.False(t, result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_HasSuccessfulPullRequestBuild_FailedBuild(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/merge_requests/987/approval_state").
		Reply(200).
		JSON(&gitlab.MergeRequestApprovalState{
			Rules: []*gitlab.MergeRequestApprovalRule{
				{Approved: true},
			},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/repository/commits/f7g8/statuses").
		Reply(200).
		JSON([]*gitlab.CommitStatus{
			{AllowFailure: false, Status: "failed"},
		})
	mr := &gitlab.MergeRequest{IID: 987, SHA: "f7g8"}
	project := &gitlab.Project{ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.HasSuccessfulPullRequestBuild(mr)

	require.NoError(t, err)
	require.False(t, result)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_IsPullRequestClosed(t *testing.T) {
	testCases := []struct {
		state string
		want  bool
	}{
		{
			state: "closed",
			want:  true,
		},
		{
			state: "locked",
			want:  true,
		},
		{
			state: "opened",
			want:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.state, func(t *testing.T) {
			mr := &gitlab.MergeRequest{State: tc.state}

			underTest := &GitLabRepository{client: setupClient()}
			result := underTest.IsPullRequestClosed(mr)

			require.Equal(t, tc.want, result)
		})
	}
}

func TestGitLabRepository_IsPullRequestMerged(t *testing.T) {
	testCases := []struct {
		state string
		want  bool
	}{
		{
			state: "merged",
			want:  true,
		},
		{
			state: "opened",
			want:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.state, func(t *testing.T) {
			mr := &gitlab.MergeRequest{State: tc.state}

			underTest := &GitLabRepository{client: setupClient()}
			result := underTest.IsPullRequestMerged(mr)

			require.Equal(t, tc.want, result)
		})
	}
}

func TestGitLabRepository_IsPullRequestOpen(t *testing.T) {
	testCases := []struct {
		state string
		want  bool
	}{
		{
			state: "closed",
			want:  false,
		},
		{
			state: "merged",
			want:  false,
		},
		{
			state: "opened",
			want:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.state, func(t *testing.T) {
			mr := &gitlab.MergeRequest{State: tc.state}

			underTest := &GitLabRepository{client: setupClient()}
			result := underTest.IsPullRequestOpen(mr)

			require.Equal(t, tc.want, result)
		})
	}
}

func TestGitLabRepository_ListPullRequestComments(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123/merge_requests/987/notes").
		Reply(200).
		JSON([]*gitlab.Note{
			{Body: "First note", ID: 1},
			{Body: "Second note", ID: 2},
		})
	mr := &gitlab.MergeRequest{IID: 987}
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.ListPullRequestComments(mr)

	require.NoError(t, err)
	wantComments := []PullRequestComment{
		{Body: "First note", ID: 1},
		{Body: "Second note", ID: 2},
	}
	require.True(t, slices.Equal(wantComments, result))
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_MergePullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Put("/api/v4/projects/123/merge_requests/987/merge").
		MatchType("json").
		JSON(map[string]interface{}{"should_remove_source_branch": true, "squash": false}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{IID: 987}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.MergePullRequest(true, mr)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_UpdatePullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Put("/api/v4/projects/123/merge_requests/987").
		MatchType("json").
		JSON(map[string]any{
			"title":        "New PR Title",
			"description":  "New PR Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n",
			"assignee_ids": nil,
		}).
		Reply(200).
		JSON(map[string]string{})
	prData := PullRequestData{
		Body:  "New PR Body",
		Title: "New PR Title",
	}
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{Description: "PR Body", IID: 987, Title: "PR Title"}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.UpdatePullRequest(prData, mr)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_UpdatePullRequest_NoUpdateRequired(t *testing.T) {
	defer gock.Off()
	prData := PullRequestData{
		Body:  "PR Body",
		Title: "PR Title",
	}
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{
		Description: "PR Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n",
		IID:         987,
		Title:       "PR Title",
	}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.UpdatePullRequest(prData, mr)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_UpdatePullRequest_UpdatedAssignees(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "y").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 25},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "z").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 26},
		})
	gock.New("http://gitlab.local").
		Put("/api/v4/projects/123/merge_requests/987").
		MatchType("json").
		JSON(map[string]any{
			"title":        "PR Title",
			"description":  gitlabMergeRequestBody,
			"assignee_ids": []int{25, 26, 1},
		}).
		Reply(200).
		JSON(map[string]string{})
	prData := PullRequestData{
		Assignees: []string{"y", "z", "a"},
		Body:      "Unit Test Body",
		Title:     "PR Title",
	}
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{
		Assignees: []*gitlab.BasicUser{
			{ID: 1, Username: "a"},
			{ID: 2, Username: "b"},
			{ID: 3, Username: "c"},
		},
		Description: gitlabMergeRequestBody,
		IID:         987,
		Title:       "PR Title",
	}
	client := setupClient()
	uc := &userCache{
		client: client,
		data:   map[string]*gitlab.User{},
	}

	underTest := &GitLabRepository{client: setupClient(), project: project, userCache: uc}
	err := underTest.UpdatePullRequest(prData, mr)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabHost_CreateFromName(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{
			input: "gitlab.local/unit/test",
			want:  "gitlab.local/unit/test",
		},
		{
			input: "https://gitlab.local/unit/test",
			want:  "gitlab.local/unit/test",
		},
		{
			input: "http://gitlab.local/unit/test",
			want:  "gitlab.local/unit/test",
		},
		{
			input: "https://gitlab.local/unit/test.git",
			want:  "gitlab.local/unit/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			defer gock.Off()
			gock.New("http://gitlab.local").
				Get("/api/v4/projects/unit/test$").
				Reply(200).
				JSON(&gitlab.Project{ID: 1, WebURL: "http://gitlab.local/unit/test"})

			underTest := &GitLabHost{client: setupClient()}
			repo, err := underTest.CreateFromName(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.want, repo.FullName())
		})
	}
}

func TestGitLabHost_ListRepositories(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects").
		MatchParams(map[string]string{
			"archived": "false",
			// Use regexp.QuoteMeta() because gock parses matchers as regular expressions
			"last_activity_after": regexp.QuoteMeta(time.Unix(0, 0).Format("2006-01-02T15:04:05Z07:00")),
			"min_access_level":    "30",
			"per_page":            "20",
		}).
		Reply(200).
		JSON([]*gitlab.Project{
			{ID: 1, WebURL: "http://gitlab.local/unittest/first"},
			{ID: 2, WebURL: "http://gitlab.local/unittest/second"},
		})
	since := time.Unix(0, 0)
	resultChan := make(chan []Repository)
	errChan := make(chan error)

	underTest := &GitLabHost{client: setupClient()}
	go underTest.ListRepositories(&since, resultChan, errChan)

	result := []Repository{}
	var wantErr error
	done := false
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case repo := <-resultChan:
			result = append(result, repo...)
		case err := <-errChan:
			wantErr = err
			done = true
		case <-timeout:
			wantErr = errors.New("test timeout")
			done = true
		}

		if done {
			break
		}
	}

	require.NoError(t, wantErr)
	require.Len(t, result, 2)
	require.Equal(t, "gitlab.local/unittest/first", result[0].FullName())
	require.Equal(t, "gitlab.local/unittest/second", result[1].FullName())
}

func TestGitLabHost_ListRepositoriesWithOpenPullRequests(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/user").
		Reply(200).
		JSON(&gitlab.User{ID: 4321})
	gock.New("http://gitlab.local").
		Get("/api/v4/merge_requests").
		MatchParam("author_id", "4321").
		MatchParam("per_page", "20").
		Reply(200).
		JSON([]*gitlab.MergeRequest{
			{IID: 1, ProjectID: 123},
			{IID: 2, ProjectID: 123},
			{IID: 3, ProjectID: 456},
			{IID: 4, ProjectID: 789},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/123").
		Reply(200).
		JSON(&gitlab.Project{ID: 123, WebURL: "http://gitlab.local/unittest/first"})
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/456").
		Reply(200).
		JSON(&gitlab.Project{ID: 456, WebURL: "http://gitlab.local/unittest/second"})
	gock.New("http://gitlab.local").
		Get("/api/v4/projects/789").
		Reply(200).
		JSON(&gitlab.Project{Archived: true, ID: 789, WebURL: "http://gitlab.local/unittest/third"})
	resultChan := make(chan []Repository)
	errChan := make(chan error)

	underTest := &GitLabHost{client: setupClient()}
	go underTest.ListRepositoriesWithOpenPullRequests(resultChan, errChan)

	result := []Repository{}
	var wantErr error
	done := false
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case repo := <-resultChan:
			result = append(result, repo...)
		case err := <-errChan:
			wantErr = err
			done = true
		case <-timeout:
			wantErr = errors.New("test timeout")
			done = true
		}

		if done {
			break
		}
	}

	require.NoError(t, wantErr)
	require.Len(t, result, 2)
	require.Equal(t, "gitlab.local/unittest/first", result[0].FullName())
	require.Equal(t, "gitlab.local/unittest/second", result[1].FullName())
}

func setupClient() *gitlab.Client {
	httpClient := &http.Client{}
	gock.InterceptClient(httpClient)
	client, _ := gitlab.NewClient("", gitlab.WithBaseURL("http://gitlab.local"), gitlab.WithHTTPClient(httpClient))
	return client
}
