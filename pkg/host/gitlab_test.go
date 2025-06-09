package host

import (
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	gitlab "gitlab.com/gitlab-org/api/client-go"
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

	require.Equal(t, "Unit Test", underTest.GetPullRequestBody(toSbPr(mr)))
}

func TestGitLabRepository_CanMergePullRequest(t *testing.T) {
	mr := &gitlab.MergeRequest{}

	underTest := &GitLabRepository{}

	result, err := underTest.CanMergePullRequest(toSbPr(mr))
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

	err := underTest.ClosePullRequest("Unit Test", toSbPr(mr))
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

	err := underTest.CreatePullRequestComment("Unit Test", toSbPr(mr))
	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests").
		MatchType("json").
		JSON(gitlab.CreateMergeRequestOptions{
			Title:              gitlab.Ptr("Unit Test Title"),
			Description:        gitlab.Ptr("Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n"),
			SourceBranch:       gitlab.Ptr("saturn-bot--unit-test"),
			TargetBranch:       gitlab.Ptr("main"),
			RemoveSourceBranch: gitlab.Ptr(false),
		}).
		Reply(200).
		JSON(gitlab.MergeRequest{
			CreatedAt:    ptr.To(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
			IID:          1,
			SourceBranch: "saturn-bot--unit-test",
			State:        "opened",
			WebURL:       "http://gitlab.local/unit/test/-/merge_requests/1",
		})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}
	prData := PullRequestData{Body: "Unit Test Body", Title: "Unit Test Title"}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	pr, err := underTest.CreatePullRequest("saturn-bot--unit-test", prData)

	require.NoError(t, err)
	require.IsType(t, &gitlab.MergeRequest{}, pr.Raw)
	// Set to nil after type check to make the next check for equality easier.
	pr.Raw = nil
	expectedPr := &PullRequest{
		CreatedAt:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Number:         1,
		WebURL:         "http://gitlab.local/unit/test/-/merge_requests/1",
		State:          PullRequestStateOpen,
		HostName:       "gitlab.local",
		BranchName:     "saturn-bot--unit-test",
		RepositoryName: "gitlab.local/unit/test",
		Type:           GitLabType,
	}
	require.Equal(t, expectedPr, pr)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequest_WithAssigneesReviewers(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "abby").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 975},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "joel").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 357},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "ellie").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 642},
		})
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests").
		MatchType("json").
		JSON(gitlab.CreateMergeRequestOptions{
			Title:              gitlab.Ptr("Unit Test Title"),
			Description:        gitlab.Ptr("Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n"),
			SourceBranch:       gitlab.Ptr("saturn-bot--unit-test"),
			TargetBranch:       gitlab.Ptr("main"),
			RemoveSourceBranch: gitlab.Ptr(false),
			AssigneeIDs:        gitlab.Ptr([]int{975}),
			ReviewerIDs:        gitlab.Ptr([]int{357, 642}),
		}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}
	prData := PullRequestData{
		Assignees: []string{"abby"},
		Body:      "Unit Test Body",
		Reviewers: []string{"joel", "ellie"},
		Title:     "Unit Test Title",
	}

	client := setupClient()
	uc := &userCache{
		client: client,
		data:   map[string]*gitlab.User{},
	}
	underTest := &GitLabRepository{client: client, project: project, userCache: uc}
	_, err := underTest.CreatePullRequest("saturn-bot--unit-test", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequest_WithLabels(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests").
		MatchType("json").
		JSON(gitlab.CreateMergeRequestOptions{
			Title:              gitlab.Ptr("Unit Test Title"),
			Description:        gitlab.Ptr("Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n"),
			SourceBranch:       gitlab.Ptr("saturn-bot--unit-test"),
			TargetBranch:       gitlab.Ptr("main"),
			RemoveSourceBranch: gitlab.Ptr(false),
			Labels:             &gitlab.LabelOptions{"unit", "test"},
		}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{DefaultBranch: "main", ID: 123}
	prData := PullRequestData{Body: "Unit Test Body", Labels: []string{"unit", "test"}, Title: "Unit Test Title"}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	_, err := underTest.CreatePullRequest("saturn-bot--unit-test", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequest_SquashOptionDefaultOn(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests").
		MatchType("json").
		JSON(gitlab.CreateMergeRequestOptions{
			Title:              gitlab.Ptr("Unit Test Title"),
			Description:        gitlab.Ptr("Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n"),
			SourceBranch:       gitlab.Ptr("saturn-bot--unit-test"),
			TargetBranch:       gitlab.Ptr("main"),
			RemoveSourceBranch: gitlab.Ptr(false),
			Squash:             gitlab.Ptr(true),
		}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{
		DefaultBranch: "main",
		ID:            123,
		SquashOption:  gitlab.SquashOptionDefaultOn,
	}
	prData := PullRequestData{
		Body:  "Unit Test Body",
		Title: "Unit Test Title",
	}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	_, err := underTest.CreatePullRequest("saturn-bot--unit-test", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_CreatePullRequest_SquashOptionAlways(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Post("/api/v4/projects/123/merge_requests").
		MatchType("json").
		JSON(gitlab.CreateMergeRequestOptions{
			Title:              gitlab.Ptr("Unit Test Title"),
			Description:        gitlab.Ptr("Unit Test Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n"),
			SourceBranch:       gitlab.Ptr("saturn-bot--unit-test"),
			TargetBranch:       gitlab.Ptr("main"),
			RemoveSourceBranch: gitlab.Ptr(false),
			Squash:             gitlab.Ptr(true),
		}).
		Reply(200).
		JSON(map[string]string{})
	project := &gitlab.Project{
		DefaultBranch: "main",
		ID:            123,
		SquashOption:  gitlab.SquashOptionAlways,
	}
	prData := PullRequestData{
		Body:  "Unit Test Body",
		Title: "Unit Test Title",
	}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	_, err := underTest.CreatePullRequest("saturn-bot--unit-test", prData)

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
	err := underTest.DeleteBranch(toSbPr(mr))

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_DeleteBranch_NoDeleteIfGitLabDeletesMR(t *testing.T) {
	defer gock.Off()
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{IID: 987, SourceBranch: "saturn-bot--unit-test", ShouldRemoveSourceBranch: true}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	err := underTest.DeleteBranch(toSbPr(mr))

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
	err := underTest.DeletePullRequestComment(comment, toSbPr(mr))

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
			{
				CreatedAt:    ptr.To(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
				IID:          123,
				SourceBranch: "saturn-bot--unit-test",
				State:        "opened",
				WebURL:       "http://gitlab.local/unit/test/-/merge_requests/123",
			},
		})
	project := &gitlab.Project{ID: 123}

	underTest := &GitLabRepository{client: setupClient(), project: project}
	result, err := underTest.FindPullRequest("saturn-bot--unit-test")

	require.NoError(t, err)
	require.IsType(t, &gitlab.MergeRequest{}, result.Raw)
	// Set to nil after type check to make the next check for equality easier.
	result.Raw = nil
	expectedPr := &PullRequest{
		CreatedAt:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Number:         123,
		WebURL:         "http://gitlab.local/unit/test/-/merge_requests/123",
		State:          PullRequestStateOpen,
		HostName:       "gitlab.local",
		BranchName:     "saturn-bot--unit-test",
		RepositoryName: "gitlab.local/unit/test",
		Type:           GitLabType,
	}
	require.Equal(t, expectedPr, result)
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
	result, err := underTest.HasSuccessfulPullRequestBuild(toSbPr(mr))

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
	result, err := underTest.HasSuccessfulPullRequestBuild(toSbPr(mr))

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
	result, err := underTest.HasSuccessfulPullRequestBuild(toSbPr(mr))

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
	result, err := underTest.ListPullRequestComments(toSbPr(mr))

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
	err := underTest.MergePullRequest(true, toSbPr(mr))

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_UpdatePullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Put("/api/v4/projects/123/merge_requests/987").
		MatchType("json").
		JSON(map[string]any{
			"title":       "New PR Title",
			"description": "New PR Body\n\n---\n\n**Auto-merge:** Disabled. Merge this manually.\n\n**Ignore:** This PR will be recreated if closed.\n\n---\n\n- [ ] If you want to rebase this PR, check this box\n\n---\n\n_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.\n",
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
	err := underTest.UpdatePullRequest(prData, toSbPr(mr))

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
	err := underTest.UpdatePullRequest(prData, toSbPr(mr))

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitLabRepository_UpdatePullRequest_UpdatedAssigneesReviewers(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "ellie").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 24},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "lev").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 25},
		})
	gock.New("http://gitlab.local").
		Get("/api/v4/users").
		MatchParam("username", "manny").
		Reply(200).
		JSON([]*gitlab.User{
			{ID: 26},
		})
	gock.New("http://gitlab.local").
		Put("/api/v4/projects/123/merge_requests/987").
		MatchType("json").
		JSON(map[string]any{
			"assignee_ids": []int{25, 26, 1},
			"reviewer_ids": []int{5, 24},
		}).
		Reply(200).
		JSON(map[string]string{})
	prData := PullRequestData{
		Assignees: []string{"lev", "manny", "abby"},
		Body:      "Unit Test Body",
		Reviewers: []string{"joel", "ellie"},
		Title:     "PR Title",
	}
	project := &gitlab.Project{ID: 123}
	mr := &gitlab.MergeRequest{
		Assignees: []*gitlab.BasicUser{
			{ID: 1, Username: "abby"},
			{ID: 2, Username: "owen"},
			{ID: 3, Username: "mel"},
		},
		Description: gitlabMergeRequestBody,
		IID:         987,
		Reviewers: []*gitlab.BasicUser{
			{ID: 4, Username: "tommy"},
			{ID: 5, Username: "joel"},
		},
		Title: "PR Title",
	}
	client := setupClient()
	uc := &userCache{
		client: client,
		data:   map[string]*gitlab.User{},
	}

	underTest := &GitLabRepository{client: setupClient(), project: project, userCache: uc}
	err := underTest.UpdatePullRequest(prData, toSbPr(mr))

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
			assert.IsType(t, &GitLabRepository{}, repo)
		})
	}
}

func TestGitLabHost_ListRepositories(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/projects").
		MatchParams(map[string]string{
			"min_access_level": "30",
			"order_by":         "updated_at",
			"per_page":         "20",
			"sort":             "desc",
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
	require.IsType(t, &GitLabRepository{}, result[0])
	require.Equal(t, "gitlab.local/unittest/second", result[1].FullName())
	require.IsType(t, &GitLabRepository{}, result[1])
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
	require.IsType(t, &GitLabRepository{}, result[0])
	require.Equal(t, "gitlab.local/unittest/second", result[1].FullName())
	require.IsType(t, &GitLabRepository{}, result[1])
	require.True(t, gock.IsDone())
}

func TestGitLabHost_PullRequestIterator_FullUpdate(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/user").
		Reply(200).
		JSON(&gitlab.User{ID: 4321})
	gitlabMr := &gitlab.MergeRequest{
		CreatedAt:    ptr.To(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
		IID:          133,
		SourceBranch: "saturn-bot--unittest",
		State:        "opened",
		WebURL:       "http://gitlab.local/unit/test/-/merge_requests/133",
	}
	gock.New("http://gitlab.local").
		Get("/api/v4/merge_requests").
		MatchParam("author_id", "4321").
		MatchParam("per_page", "100").
		MatchParam("order_by", "updated_at").
		Reply(200).
		JSON([]*gitlab.MergeRequest{gitlabMr})

	host := &GitLabHost{client: setupClient()}
	iterator := host.PullRequestIterator()
	result := slices.Collect(iterator.ListPullRequests(nil))

	require.NoError(t, iterator.Error())
	require.Len(t, result, 1)
	wantPr := &PullRequest{
		CreatedAt:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Number:         133,
		WebURL:         "http://gitlab.local/unit/test/-/merge_requests/133",
		Raw:            gitlabMr,
		State:          PullRequestStateOpen,
		HostName:       "gitlab.local",
		BranchName:     "saturn-bot--unittest",
		RepositoryName: "gitlab.local/unit/test",
		Type:           GitLabType,
	}
	require.Equal(t, wantPr, result[0])
	require.True(t, gock.IsDone())
}

func TestGitLabHost_PullRequestIterator_PartialUpdate(t *testing.T) {
	defer gock.Off()
	gock.New("http://gitlab.local").
		Get("/api/v4/user").
		Reply(200).
		JSON(&gitlab.User{ID: 4321})
	gitlabMr := &gitlab.MergeRequest{
		CreatedAt:    ptr.To(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
		IID:          133,
		SourceBranch: "saturn-bot--unittest",
		State:        "opened",
		WebURL:       "http://gitlab.local/unit/test/-/merge_requests/133",
	}
	gock.New("http://gitlab.local").
		Get("/api/v4/merge_requests").
		MatchParam("author_id", "^4321$").
		MatchParam("per_page", "^100$").
		MatchParam("order_by", "^updated_at$").
		MatchParam("sort", "^desc$").
		MatchParam("updated_after", "^2000-01-01T00:00:00Z$").
		Reply(200).
		JSON([]*gitlab.MergeRequest{gitlabMr})

	host := &GitLabHost{client: setupClient()}
	iterator := host.PullRequestIterator()
	since := ptr.To(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	result := slices.Collect(iterator.ListPullRequests(since))

	require.NoError(t, iterator.Error())
	require.Len(t, result, 1)
	wantPr := &PullRequest{
		CreatedAt:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Number:         133,
		WebURL:         "http://gitlab.local/unit/test/-/merge_requests/133",
		Raw:            gitlabMr,
		State:          PullRequestStateOpen,
		HostName:       "gitlab.local",
		BranchName:     "saturn-bot--unittest",
		RepositoryName: "gitlab.local/unit/test",
		Type:           GitLabType,
	}
	require.Equal(t, wantPr, result[0])
	require.True(t, gock.IsDone())
}

func setupClient() *gitlab.Client {
	httpClient := &http.Client{}
	gock.InterceptClient(httpClient)
	client, _ := gitlab.NewClient("", gitlab.WithBaseURL("http://gitlab.local"), gitlab.WithHTTPClient(httpClient))
	return client
}
