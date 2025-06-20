package host

import (
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

const (
	githubPullRequestBody = `pull request body

---

**Auto-merge:** Disabled. Merge this manually.

**Ignore:** This PR will be recreated if closed.

---

- [ ] If you want to rebase this PR, check this box

---

_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.
`
)

func TestGitHubRepository_BaseBranch(t *testing.T) {
	repo := &GitHubRepository{
		repo: &github.Repository{
			DefaultBranch: github.Ptr("main"),
		},
	}

	assert.Equal(t, "main", repo.BaseBranch())
}

func TestGitHubRepository_CanMergePullRequest(t *testing.T) {
	pr := &github.PullRequest{
		Mergeable: github.Ptr(true),
	}

	repo := &GitHubRepository{}
	result, err := repo.CanMergePullRequest(toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, result)
}

func TestGitHubRepository_CanMergePullRequest_MergeableNil(t *testing.T) {
	pr := &github.PullRequest{}

	repo := &GitHubRepository{}
	result, err := repo.CanMergePullRequest(toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, result)
}

func TestGitHubRepository_CloneUrlHttp(t *testing.T) {
	repo := &GitHubRepository{
		repo: &github.Repository{
			CloneURL: github.Ptr("https://github.com/unit/test.git"),
		},
	}

	assert.Equal(t, "https://github.com/unit/test.git", repo.CloneUrlHttp())
}

func TestGitHubRepository_CloneUrlSsh(t *testing.T) {
	repo := &GitHubRepository{
		repo: &github.Repository{
			SSHURL: github.Ptr("git@github.com:unit/test.git"),
		},
	}

	assert.Equal(t, "git@github.com:unit/test.git", repo.CloneUrlSsh())
}

func TestGitHubRepository_ClosePullRequest(t *testing.T) {
	defer gock.Off()
	pr := &github.PullRequest{
		Number: github.Ptr(987),
	}
	gock.New("https://api.github.com").
		Post("/repos/unit/test/issues/987/comments").
		MatchType("json").
		JSON(map[string]interface{}{
			"body": "close pull request",
		}).
		Reply(200).
		JSON(map[string]string{})
	gock.New("https://api.github.com").
		Patch("/repos/unit/test/pulls/987").
		MatchType("json").
		JSON(map[string]interface{}{
			"state": "closed",
		}).
		Reply(200).
		JSON(github.PullRequest{
			Number: github.Ptr(987),
			State:  github.Ptr("closed"),
		})

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	prUpdated, err := repo.ClosePullRequest("close pull request", toSbPr(pr))

	require.NoError(t, err)
	require.Equal(t, PullRequestStateClosed, prUpdated.State)
	require.True(t, gock.IsDone())
}

func TestGitHubRepository_CreatePullRequestComment(t *testing.T) {
	defer gock.Off()
	pr := &github.PullRequest{
		Number: github.Ptr(987),
	}
	gock.New("https://api.github.com").
		Post("/repos/unit/test/issues/987/comments").
		MatchType("json").
		JSON(map[string]interface{}{
			"body": "pull request comment",
		}).
		Reply(200).
		JSON(map[string]string{})

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.CreatePullRequestComment("pull request comment", toSbPr(pr))

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

var createPullRequestRespBody = github.PullRequest{
	CreatedAt: &github.Timestamp{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
	Number:    github.Ptr(1),
	HTMLURL:   github.Ptr("https://github.com/unit/test/pull/1"),
	Head: &github.PullRequestBranch{
		Ref: github.Ptr("saturn-bot--unittest"),
		Repo: &github.Repository{
			HTMLURL: github.Ptr("https://github.com/unit/test"),
		},
	},
	State: github.Ptr("open"),
}

func TestGitHubRepository_CreatePullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Post("/repos/unit/test/pulls").
		MatchType("json").
		JSON(&github.NewPullRequest{
			Base:                github.Ptr("main"),
			Body:                github.Ptr(githubPullRequestBody),
			Head:                github.Ptr("unittest"),
			MaintainerCanModify: github.Ptr(true),
			Title:               github.Ptr("pull request title"),
		}).
		Reply(200).
		JSON(createPullRequestRespBody)
	prData := PullRequestData{
		Body:     "pull request body",
		TaskName: "Unit Test",
		Title:    "pull request title",
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	pr, err := repo.CreatePullRequest("unittest", prData)

	require.NoError(t, err)
	wantPr := &PullRequest{
		CreatedAt:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Number:         1,
		WebURL:         "https://github.com/unit/test/pull/1",
		Raw:            &createPullRequestRespBody,
		State:          PullRequestStateOpen,
		HostName:       "github.com",
		BranchName:     "saturn-bot--unittest",
		RepositoryName: "github.com/unit/test",
		Type:           GitHubType,
	}
	require.Equal(t, wantPr, pr, "returns the pull request struct")
	require.True(t, gock.IsDone())
}

func TestGitHubRepository_CreatePullRequest_WithAssignees(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Post("/repos/unit/test/pulls").
		MatchType("json").
		JSON(&github.NewPullRequest{
			Base:                github.Ptr("main"),
			Body:                github.Ptr(githubPullRequestBody),
			Head:                github.Ptr("unittest"),
			MaintainerCanModify: github.Ptr(true),
			Title:               github.Ptr("pull request title"),
		}).
		Reply(200).
		JSON(createPullRequestRespBody)
	gock.New("https://api.github.com").
		Post("/repos/unit/test/issues/1/assignees").
		MatchType("json").
		JSON(map[string]any{
			"assignees": []string{"jane", "joe"},
		}).
		Reply(200)
	prData := PullRequestData{
		Assignees: []string{"jane", "joe"},
		Body:      "pull request body",
		TaskName:  "Unit Test",
		Title:     "pull request title",
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	_, err := repo.CreatePullRequest("unittest", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitHubRepository_CreatePullRequest_WithReviewers(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Post("/repos/unit/test/pulls").
		MatchType("json").
		JSON(&github.NewPullRequest{
			Base:                github.Ptr("main"),
			Body:                github.Ptr(githubPullRequestBody),
			Head:                github.Ptr("unittest"),
			MaintainerCanModify: github.Ptr(true),
			Title:               github.Ptr("pull request title"),
		}).
		Reply(200).
		JSON(createPullRequestRespBody)
	gock.New("https://api.github.com").
		Post("/repos/unit/test/pulls/1/requested_reviewers").
		MatchType("json").
		JSON(github.ReviewersRequest{Reviewers: []string{"abby", "owen"}}).
		Reply(200).
		JSON(&github.PullRequest{
			Number: github.Ptr(1),
		})
	prData := PullRequestData{
		Body:      "pull request body",
		Reviewers: []string{"abby", "owen"},
		TaskName:  "Unit Test",
		Title:     "pull request title",
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	_, err := repo.CreatePullRequest("unittest", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitHubRepository_FindPullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/unit/test/pulls").
		MatchParams(map[string]string{
			"page":     "1",
			"per_page": "20",
			"state":    "all",
		}).
		Reply(200).
		JSON([]*github.PullRequest{
			{Head: &github.PullRequestBranch{Ref: github.Ptr("other")}},
			{
				CreatedAt: &github.Timestamp{Time: time.Now()},
				HTMLURL:   github.Ptr("https://github.com/unit/test/pulls/1"),
				Head:      &github.PullRequestBranch{Ref: github.Ptr("unittest")},
				Number:    github.Ptr(123),
				State:     github.Ptr("open"),
			},
		})

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	prId, err := repo.FindPullRequest("unittest")

	require.NoError(t, err)
	assert.False(t, prId.CreatedAt.IsZero())
	assert.Equal(t, int64(123), prId.Number)
	assert.Equal(t, "https://github.com/unit/test/pulls/1", prId.WebURL)
	assert.IsType(t, &github.PullRequest{}, prId.Raw)
	assert.Equal(t, PullRequestStateOpen, prId.State)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_FindPullRequest_PullRequestNotFound(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/unit/test/pulls").
		MatchParams(map[string]string{
			"page":     "1",
			"per_page": "20",
			"state":    "all",
		}).
		Reply(200).
		JSON([]*github.PullRequest{
			{Head: &github.PullRequestBranch{Ref: github.Ptr("other")}},
			{Head: &github.PullRequestBranch{Ref: github.Ptr("some")}},
		})

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	prId, err := repo.FindPullRequest("unittest")

	assert.ErrorIs(t, err, ErrPullRequestNotFound)
	assert.Nil(t, prId)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_FullName(t *testing.T) {
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	fullName := repo.FullName()

	assert.Equal(t, "github.com/unit/test", fullName)
}

func TestGitHubRepository_GetPullRequestBody(t *testing.T) {
	pr := &github.PullRequest{
		Body: github.Ptr("pull request body"),
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	body := repo.GetPullRequestBody(toSbPr(pr))

	assert.Equal(t, "pull request body", body)
}

func TestGitHubRepository_DeleteBranch(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Delete("/repos/unit/test/git/refs/heads/unittest").
		Reply(200)

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	pr := &github.PullRequest{
		Number: github.Ptr(987),
		Head: &github.PullRequestBranch{
			Ref: github.Ptr("unittest"),
		},
	}
	err := repo.DeleteBranch(toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_DeleteBranch_RepoAutoDelete(t *testing.T) {
	defer gock.Off()
	githubRepo := setupGitHubRepository()
	githubRepo.DeleteBranchOnMerge = github.Ptr(true)
	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   githubRepo,
	}
	pr := &github.PullRequest{
		Number: github.Ptr(987),
		Head: &github.PullRequestBranch{
			Ref: github.Ptr("unittest"),
		},
	}
	err := repo.DeleteBranch(toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_DeletePullRequestComment(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Delete("/repos/unit/test/issues/comments/765").
		Reply(200)
	pr := &github.PullRequest{
		Number: github.Ptr(987),
	}
	comment := PullRequestComment{ID: 765}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.DeletePullRequestComment(comment, toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_HasSuccessfulPullRequestBuild_Success(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/unit/test/commits/a1b2c3d4/check-runs").
		MatchParam("page", "1").
		MatchParam("per_page", "20").
		Reply(200).
		JSON(&github.ListCheckRunsResults{
			CheckRuns: []*github.CheckRun{
				{Conclusion: github.Ptr("success")},
				{Conclusion: github.Ptr("success")},
			},
		})
	pr := &github.PullRequest{
		Number: github.Ptr(987),
		Head: &github.PullRequestBranch{
			SHA: github.Ptr("a1b2c3d4"),
		},
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	result, err := repo.HasSuccessfulPullRequestBuild(toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, result)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_HasSuccessfulPullRequestBuild_Failed(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/unit/test/commits/a1b2c3d4/check-runs").
		MatchParam("page", "1").
		MatchParam("per_page", "20").
		Reply(200).
		JSON(&github.ListCheckRunsResults{
			CheckRuns: []*github.CheckRun{
				{Conclusion: github.Ptr("success")},
				{Conclusion: github.Ptr("failed")},
			},
		})
	pr := &github.PullRequest{
		Number: github.Ptr(987),
		Head: &github.PullRequestBranch{
			SHA: github.Ptr("a1b2c3d4"),
		},
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	result, err := repo.HasSuccessfulPullRequestBuild(toSbPr(pr))

	require.NoError(t, err)
	assert.False(t, result)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_IsPullRequestClosed(t *testing.T) {
	pr := &github.PullRequest{
		MergedAt: nil,
		State:    github.Ptr("closed"),
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	result := repo.IsPullRequestClosed(pr)

	assert.True(t, result)

	pr = &github.PullRequest{
		MergedAt: nil,
		State:    github.Ptr("open"),
	}
	repo = &GitHubRepository{repo: setupGitHubRepository()}
	result = repo.IsPullRequestClosed(pr)

	assert.False(t, result)
}

func TestGitHubRepository_IsPullRequestMerged(t *testing.T) {
	pr := &github.PullRequest{
		MergedAt: &github.Timestamp{Time: time.Now()},
		State:    github.Ptr("closed"),
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	result := repo.IsPullRequestMerged(pr)

	assert.True(t, result)

	pr = &github.PullRequest{
		MergedAt: nil,
		State:    github.Ptr("open"),
	}
	repo = &GitHubRepository{repo: setupGitHubRepository()}
	result = repo.IsPullRequestMerged(pr)

	assert.False(t, result)
}

func TestGitHubRepository_IsPullRequestOpen(t *testing.T) {
	pr := &github.PullRequest{
		MergedAt: nil,
		State:    github.Ptr("open"),
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	result := repo.IsPullRequestOpen(pr)

	assert.True(t, result)

	pr = &github.PullRequest{
		MergedAt: &github.Timestamp{Time: time.Now()},
		State:    github.Ptr("closed"),
	}
	repo = &GitHubRepository{repo: setupGitHubRepository()}
	result = repo.IsPullRequestOpen(pr)

	assert.False(t, result)
}

func TestGitHubRepository_ListPullRequestComments(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/unit/test/issues/987/comments").
		MatchParam("page", "1").
		MatchParam("per_page", "20").
		Reply(200).
		JSON([]*github.IssueComment{
			{Body: github.Ptr("comment body"), ID: github.Ptr(int64(357))},
		})
	pr := &github.PullRequest{
		Number: github.Ptr(987),
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	result, err := repo.ListPullRequestComments(toSbPr(pr))

	require.NoError(t, err)
	comment := PullRequestComment{Body: "comment body", ID: 357}
	assert.Equal(t, []PullRequestComment{comment}, result)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_MergePullRequest_NoDeleteBranch(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Put("/repos/unit/test/pulls/987/merge").
		JSON(map[string]string{
			"commit_message": "Auto-merge by saturn-bot",
		}).
		Reply(200)
	pr := &github.PullRequest{
		Number: github.Ptr(987),
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.MergePullRequest(false, toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_MergePullRequest_DeleteBranch(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Put("/repos/unit/test/pulls/987/merge").
		JSON(map[string]string{
			"commit_message": "Auto-merge by saturn-bot",
		}).
		Reply(200)
	gock.New("https://api.github.com").
		Delete("/repos/unit/test/git/refs/heads/unittest").
		Reply(200)
	pr := &github.PullRequest{
		Number: github.Ptr(987),
		Head: &github.PullRequestBranch{
			Ref: github.Ptr("unittest"),
		},
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.MergePullRequest(true, toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_MergePullRequest_DeleteBranchByGitHub(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Put("/repos/unit/test/pulls/987/merge").
		JSON(map[string]string{
			"commit_message": "Auto-merge by saturn-bot",
		}).
		Reply(200)
	pr := &github.PullRequest{
		Number: github.Ptr(987),
		Head: &github.PullRequestBranch{
			Ref: github.Ptr("unittest"),
		},
	}

	ghRepo := setupGitHubRepository()
	ghRepo.DeleteBranchOnMerge = github.Ptr(true)
	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   ghRepo,
	}
	err := repo.MergePullRequest(true, toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_MergePullRequest_MergeMethods(t *testing.T) {
	testCases := []struct {
		method string
		repo   *github.Repository
	}{
		{
			method: "merge",
			repo: func() *github.Repository {
				ghRepo := setupGitHubRepository()
				ghRepo.AllowMergeCommit = github.Ptr(true)
				return ghRepo
			}(),
		},

		{
			method: "rebase",
			repo: func() *github.Repository {
				ghRepo := setupGitHubRepository()
				ghRepo.AllowRebaseMerge = github.Ptr(true)
				return ghRepo
			}(),
		},

		{
			method: "squash",
			repo: func() *github.Repository {
				ghRepo := setupGitHubRepository()
				ghRepo.AllowSquashMerge = github.Ptr(true)
				return ghRepo
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			defer gock.Off()
			gock.New("https://api.github.com").
				Put("/repos/unit/test/pulls/987/merge").
				JSON(map[string]string{
					"commit_message": "Auto-merge by saturn-bot",
					"merge_method":   tc.method,
				}).
				Reply(200)
			pr := &github.PullRequest{
				Number: github.Ptr(987),
			}

			repo := &GitHubRepository{
				client: setupGitHubTestClient(),
				repo:   tc.repo,
			}
			err := repo.MergePullRequest(false, toSbPr(pr))

			require.NoError(t, err)
			assert.True(t, gock.IsDone())
		})
	}
}

func TestGitHubRepository_UpdatePullRequest_Update(t *testing.T) {
	body := `new body

---

**Auto-merge:** Disabled. Merge this manually.

**Ignore:** This PR will be recreated if closed.

---

- [ ] If you want to rebase this PR, check this box

---

_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.
`
	defer gock.Off()
	gock.New("https://api.github.com").
		Patch("/repos/unit/test/pulls/987").
		JSON(map[string]string{
			"title": "new title",
			"body":  body,
		}).
		Reply(200)
	pr := &github.PullRequest{
		Body:   github.Ptr("old body"),
		Number: github.Ptr(987),
		Title:  github.Ptr("old title"),
	}
	prData := PullRequestData{
		Body:     "new body",
		TaskName: "Unit Test",
		Title:    "new title",
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.UpdatePullRequest(prData, toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_UpdatePullRequest_NoUpdate(t *testing.T) {
	body := `old body

---

**Auto-merge:** Disabled. Merge this manually.

**Ignore:** This PR will be recreated if closed.

---

- [ ] If you want to rebase this PR, check this box

---

_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.
`
	defer gock.Off()
	pr := &github.PullRequest{
		Body:   github.Ptr(body),
		Number: github.Ptr(987),
		Title:  github.Ptr("old title"),
	}
	prData := PullRequestData{
		Body:     "old body",
		TaskName: "Unit Test",
		Title:    "old title",
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.UpdatePullRequest(prData, toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_UpdatePullRequest_UpdatedAssignees(t *testing.T) {
	body := `PR body

---

**Auto-merge:** Disabled. Merge this manually.

**Ignore:** This PR will be recreated if closed.

---

- [ ] If you want to rebase this PR, check this box

---

_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.
`
	defer gock.Off()
	gock.New("https://api.github.com").
		Delete("/repos/unit/test/issues/987/assignees").
		JSON(map[string]any{
			"assignees": []string{"b", "c"},
		}).
		Reply(200)
	gock.New("https://api.github.com").
		Post("/repos/unit/test/issues/987/assignees").
		JSON(map[string]any{
			"assignees": []string{"y", "z"},
		}).
		Reply(200)
	pr := &github.PullRequest{
		Assignees: []*github.User{
			{Login: github.Ptr("a")},
			{Login: github.Ptr("b")},
			{Login: github.Ptr("c")},
		},
		Body:   github.Ptr(body),
		Number: github.Ptr(987),
		Title:  github.Ptr("PR title"),
	}
	prData := PullRequestData{
		Assignees: []string{"y", "z", "a"},
		Body:      "PR body",
		TaskName:  "Unit Test",
		Title:     "PR title",
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.UpdatePullRequest(prData, toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_UpdatePullRequest_UpdatedReviewers(t *testing.T) {
	body := `PR body

---

**Auto-merge:** Disabled. Merge this manually.

**Ignore:** This PR will be recreated if closed.

---

- [ ] If you want to rebase this PR, check this box

---

_This pull request has been created by [saturn-bot](https://github.com/wndhydrnt/saturn-bot)_ 🪐🤖.
`
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/unit/test/pulls/987/reviews").
		MatchParams(map[string]string{
			"page":     "1",
			"per_page": "20",
		}).
		Reply(200).
		JSON([]*github.PullRequestReview{
			{User: &github.User{Login: github.Ptr("ellie")}},
			{User: &github.User{Login: github.Ptr("jesse")}},
			{User: &github.User{Login: github.Ptr("ellie")}}, // user ellie approved twice for some reason
		})
	gock.New("https://api.github.com").
		Post("/repos/unit/test/pulls/987/requested_reviewers").
		JSON(github.ReviewersRequest{Reviewers: []string{"dina"}}).
		Reply(200)
	gock.New("https://api.github.com").
		Delete("/repos/unit/test/pulls/987/requested_reviewers").
		JSON(github.ReviewersRequest{Reviewers: []string{"tommy"}}).
		Reply(200)
	pr := &github.PullRequest{
		Body:   github.Ptr(body),
		Number: github.Ptr(987),
		RequestedReviewers: []*github.User{
			{Login: github.Ptr("tommy")},
			{Login: github.Ptr("joel")},
		},
		Title: github.Ptr("PR title"),
	}
	prData := PullRequestData{
		Body:      "PR body",
		Reviewers: []string{"ellie", "joel", "dina"},
		TaskName:  "Unit Test",
		Title:     "PR title",
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.UpdatePullRequest(prData, toSbPr(pr))

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_WebUrl(t *testing.T) {
	githubRepo := setupGitHubRepository()
	githubRepo.HTMLURL = github.Ptr("https://github.com/unit/test")
	repo := &GitHubRepository{repo: githubRepo}
	url := repo.WebUrl()

	assert.Equal(t, "https://github.com/unit/test", url)
}

func TestGitHubHost_CreateFromName(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{
			input: "github.com/unit/test",
			want:  "github.com/unit/test",
		},
		{
			input: "https://github.com/unit/test",
			want:  "github.com/unit/test",
		},
		{
			input: "http://github.com/unit/test",
			want:  "github.com/unit/test",
		},
		{
			input: "https://github.com/unit/test.git",
			want:  "github.com/unit/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			defer gock.Off()
			gock.New("https://api.github.com").
				Get("/repos/unit/test").
				Reply(200).
				JSON(&github.Repository{
					Owner: &github.User{Login: github.Ptr("unit")},
					Name:  github.Ptr("test"),
				})

			gh := &GitHubHost{
				client: setupGitHubTestClient(),
			}
			repo, err := gh.CreateFromName(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.want, repo.FullName())
			assert.IsType(t, &GitHubRepository{}, repo)
		})
	}
}

func TestGitHubHost_PullRequestIterator_FullUpdate(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/user").
		Reply(200).
		JSON(&github.User{
			Login: github.Ptr("unittest"),
		})
	gock.New("https://api.github.com").
		Get("/search/issues").
		MatchParams(map[string]string{
			"page":     "1",
			"per_page": "20",
			"q":        `is:pr author:unittest archived:false`,
		}).
		Reply(200).
		SetHeader("Link", `<https://api.github.com/repositories/1300192/issues?page=2>; rel="next"`).
		JSON(&github.IssuesSearchResult{
			Issues: []*github.Issue{
				// Regular issue, not a pull request
				{
					PullRequestLinks: nil,
				},
			},
		})
	gock.New("https://api.github.com").
		Get("/search/issues").
		MatchParams(map[string]string{
			"page":     "2",
			"per_page": "20",
			"q":        `is:pr author:unittest archived:false`,
		}).
		Reply(200).
		JSON(&github.IssuesSearchResult{
			Issues: []*github.Issue{
				{
					PullRequestLinks: &github.PullRequestLinks{URL: github.Ptr("https://api.github.com/repos/unittest/first/pulls/5")},
				},
			},
		})

	githubPr := &github.PullRequest{
		CreatedAt: &github.Timestamp{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		Number:    github.Ptr(5),
		HTMLURL:   github.Ptr("https://github.com/unittest/first/pull/5"),
		State:     github.Ptr("open"),
		Head: &github.PullRequestBranch{
			Ref: github.Ptr("saturn-bot--unittest"),
			Repo: &github.Repository{
				HTMLURL: github.Ptr("https://github.com/unittest/first"),
			},
		},
	}
	gock.New("https://api.github.com").
		Get("/repos/unittest/first/pulls/5").
		Reply(200).
		JSON(githubPr)

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	iterator := gh.PullRequestIterator()
	result := slices.Collect(iterator.ListPullRequests(nil))

	require.NoError(t, iterator.Error(), "iterator does not error")
	require.Len(t, result, 1, "returns one pull request")
	wantPr := &PullRequest{
		CreatedAt:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Number:         5,
		WebURL:         "https://github.com/unittest/first/pull/5",
		Raw:            githubPr,
		State:          PullRequestStateOpen,
		HostName:       "github.com",
		BranchName:     "saturn-bot--unittest",
		RepositoryName: "github.com/unittest/first",
		Type:           GitHubType,
	}
	require.Equal(t, wantPr, result[0], "returns the expected pull request")
	assert.True(t, gock.IsDone())
}

func TestGitHubHost_PullRequestIterator_PartialUpdate(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/user").
		Reply(200).
		JSON(&github.User{
			Login: github.Ptr("unittest"),
		})
	gock.New("https://api.github.com").
		Get("/search/issues").
		MatchParams(map[string]string{
			"page":     "1",
			"per_page": "20",
			"q":        `is:pr author:unittest archived:false`,
		}).
		Reply(200).
		JSON(&github.IssuesSearchResult{
			Issues: []*github.Issue{
				{
					UpdatedAt:        &github.Timestamp{Time: time.Date(2000, 1, 3, 0, 0, 0, 0, time.UTC)},
					PullRequestLinks: &github.PullRequestLinks{URL: github.Ptr("https://api.github.com/repos/unittest/first/pulls/5")},
				},
				{
					UpdatedAt:        &github.Timestamp{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
					PullRequestLinks: &github.PullRequestLinks{URL: github.Ptr("https://api.github.com/repos/unittest/first/pulls/7")},
				},
			},
		})

	githubPr := &github.PullRequest{
		CreatedAt: &github.Timestamp{Time: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		Number:    github.Ptr(5),
		HTMLURL:   github.Ptr("https://github.com/unittest/first/pull/5"),
		State:     github.Ptr("open"),
		Head: &github.PullRequestBranch{
			Ref: github.Ptr("saturn-bot--unittest"),
			Repo: &github.Repository{
				HTMLURL: github.Ptr("https://github.com/unittest/first"),
			},
		},
	}
	gock.New("https://api.github.com").
		Get("/repos/unittest/first/pulls/5").
		Reply(200).
		JSON(githubPr)

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	iterator := gh.PullRequestIterator()
	since := time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
	result := slices.Collect(iterator.ListPullRequests(ptr.To(since)))

	require.NoError(t, iterator.Error(), "iterator does not error")
	require.Len(t, result, 1, "returns one pull request")
	wantPr := &PullRequest{
		CreatedAt:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Number:         5,
		WebURL:         "https://github.com/unittest/first/pull/5",
		Raw:            githubPr,
		State:          PullRequestStateOpen,
		HostName:       "github.com",
		BranchName:     "saturn-bot--unittest",
		RepositoryName: "github.com/unittest/first",
		Type:           GitHubType,
	}
	require.Equal(t, wantPr, result[0], "returns the expected pull request")
	assert.True(t, gock.IsDone())
}

func TestGitHubHost_RepositoryIterator_FullUpdate(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/user/repos").
		MatchParams(map[string]string{
			"direction":   "desc",
			"page":        "1",
			"per_page":    "20",
			"sort":        "updated",
			"visibility":  "all",
			"affiliation": "owner,collaborator",
		}).
		Reply(200).
		SetHeader("Link", `<https://api.github.com/user/repos?page=2>; rel="next"`).
		JSON([]*github.Repository{
			{Owner: &github.User{Login: github.Ptr("unittest")}, Name: github.Ptr("first"), UpdatedAt: &github.Timestamp{Time: time.Now()}},
		})
	gock.New("https://api.github.com").
		Get("/user/repos").
		MatchParams(map[string]string{
			"direction":   "desc",
			"page":        "2",
			"per_page":    "20",
			"sort":        "updated",
			"visibility":  "all",
			"affiliation": "owner,collaborator",
		}).
		Reply(200).
		JSON([]*github.Repository{
			{Owner: &github.User{Login: github.Ptr("unittest")}, Name: github.Ptr("second"), UpdatedAt: &github.Timestamp{Time: time.Now()}},
		})

	gock.New("https://api.github.com").
		Get("/repos/unittest/first").
		Reply(200).
		JSON(&github.Repository{
			Owner: &github.User{Login: github.Ptr("unittest")}, Name: github.Ptr("first"), UpdatedAt: &github.Timestamp{Time: time.Now()},
		})
	gock.New("https://api.github.com").
		Get("/repos/unittest/second").
		Reply(200).
		JSON(&github.Repository{
			Owner: &github.User{Login: github.Ptr("unittest")}, Name: github.Ptr("second"), UpdatedAt: &github.Timestamp{Time: time.Now()},
		})

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	iterator := gh.RepositoryIterator()
	result := slices.Collect(iterator.ListRepositories(nil))

	require.NoError(t, iterator.Error())
	assert.Len(t, result, 2)
	assert.Equal(t, "github.com/unittest/first", result[0].FullName())
	assert.IsType(t, &GitHubRepository{}, result[0])
	assert.Equal(t, "github.com/unittest/second", result[1].FullName())
	assert.IsType(t, &GitHubRepository{}, result[1])
	assert.True(t, gock.IsDone())
}

func TestGitHubHost_RepositoryIterator_PartialUpdate(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/user/repos").
		MatchParams(map[string]string{
			"direction":   "desc",
			"page":        "1",
			"per_page":    "20",
			"sort":        "updated",
			"visibility":  "all",
			"affiliation": "owner,collaborator",
		}).
		Reply(200).
		JSON([]*github.Repository{
			{
				Owner:     &github.User{Login: github.Ptr("unittest")},
				Name:      github.Ptr("second"),
				UpdatedAt: &github.Timestamp{Time: time.Date(2000, 1, 1, 1, 0, 0, 0, time.UTC)},
			},
			{
				Owner:     &github.User{Login: github.Ptr("unittest")},
				Name:      github.Ptr("first"),
				UpdatedAt: &github.Timestamp{Time: time.Date(1999, 12, 31, 23, 59, 59, 0, time.UTC)},
			},
		})
	gock.New("https://api.github.com").
		Get("/repos/unittest/second").
		Reply(200).
		JSON(&github.Repository{
			Owner: &github.User{Login: github.Ptr("unittest")}, Name: github.Ptr("second"), UpdatedAt: &github.Timestamp{Time: time.Now()},
		})

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	iterator := gh.RepositoryIterator()
	since := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	result := slices.Collect(iterator.ListRepositories(ptr.To(since)))

	require.NoError(t, iterator.Error())
	assert.Len(t, result, 1)
	assert.Equal(t, "github.com/unittest/second", result[0].FullName())
	assert.IsType(t, &GitHubRepository{}, result[0])
	assert.True(t, gock.IsDone())
}

func TestGitHubHost_RepositoryIterator_ErrorList(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/user/repos").
		MatchParams(map[string]string{
			"direction":   "desc",
			"page":        "1",
			"per_page":    "20",
			"sort":        "updated",
			"visibility":  "all",
			"affiliation": "owner,collaborator",
		}).
		Reply(502).
		SetError(errors.New("bad gateway"))

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	iterator := gh.RepositoryIterator()
	result := slices.Collect(iterator.ListRepositories(nil))

	require.ErrorContains(t, iterator.Error(), "bad gateway")
	require.Len(t, result, 0)
	assert.True(t, gock.IsDone())
}

func TestGitHubHost_RepositoryIterator_ErrorGetRepository(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/user/repos").
		MatchParams(map[string]string{
			"direction":   "desc",
			"page":        "1",
			"per_page":    "20",
			"sort":        "updated",
			"visibility":  "all",
			"affiliation": "owner,collaborator",
		}).
		Reply(200).
		JSON([]*github.Repository{
			{
				Owner:     &github.User{Login: github.Ptr("unittest")},
				Name:      github.Ptr("second"),
				UpdatedAt: &github.Timestamp{Time: time.Date(2000, 1, 1, 1, 0, 0, 0, time.UTC)},
			},
		})
	gock.New("https://api.github.com").
		Get("/repos/unittest/second").
		Reply(502).
		SetError(errors.New("bad gateway"))

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	iterator := gh.RepositoryIterator()
	result := slices.Collect(iterator.ListRepositories(nil))

	require.ErrorContains(t, iterator.Error(), "bad gateway")
	require.Len(t, result, 0)
	assert.True(t, gock.IsDone())
}

func setupGitHubTestClient() *github.Client {
	httpClient := &http.Client{}
	gock.InterceptClient(httpClient)
	return github.NewClient(httpClient)
}

func setupGitHubRepository() *github.Repository {
	return &github.Repository{
		DefaultBranch: github.Ptr("main"),
		Name:          github.Ptr("test"),
		Owner: &github.User{
			Login: github.Ptr("unit"),
		},
	}
}
