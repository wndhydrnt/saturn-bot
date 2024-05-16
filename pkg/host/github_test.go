package host

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v59/github"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			DefaultBranch: github.String("main"),
		},
	}

	assert.Equal(t, "main", repo.BaseBranch())
}

func TestGitHubRepository_CanMergePullRequest(t *testing.T) {
	pr := &github.PullRequest{
		Mergeable: github.Bool(true),
	}

	repo := &GitHubRepository{}
	result, err := repo.CanMergePullRequest(pr)

	require.NoError(t, err)
	assert.True(t, result)
}

func TestGitHubRepository_CanMergePullRequest_MergeableNil(t *testing.T) {
	pr := &github.PullRequest{}

	repo := &GitHubRepository{}
	result, err := repo.CanMergePullRequest(pr)

	require.NoError(t, err)
	assert.True(t, result)
}

func TestGitHubRepository_CloneUrlHttp(t *testing.T) {
	repo := &GitHubRepository{
		repo: &github.Repository{
			CloneURL: github.String("https://github.com/unit/test.git"),
		},
	}

	assert.Equal(t, "https://github.com/unit/test.git", repo.CloneUrlHttp())
}

func TestGitHubRepository_CloneUrlSsh(t *testing.T) {
	repo := &GitHubRepository{
		repo: &github.Repository{
			SSHURL: github.String("git@github.com:unit/test.git"),
		},
	}

	assert.Equal(t, "git@github.com:unit/test.git", repo.CloneUrlSsh())
}

func TestGitHubRepository_ClosePullRequest(t *testing.T) {
	defer gock.Off()
	pr := &github.PullRequest{
		Number: github.Int(987),
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
		JSON(map[string]string{})

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.ClosePullRequest("close pull request", pr)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitHubRepository_CreatePullRequestComment(t *testing.T) {
	defer gock.Off()
	pr := &github.PullRequest{
		Number: github.Int(987),
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
	err := repo.CreatePullRequestComment("pull request comment", pr)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitHubRepository_CreatePullRequest(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Post("/repos/unit/test/pulls").
		MatchType("json").
		JSON(&github.NewPullRequest{
			Base:                github.String("main"),
			Body:                github.String(githubPullRequestBody),
			Head:                github.String("unittest"),
			MaintainerCanModify: github.Bool(true),
			Title:               github.String("pull request title"),
		}).
		Reply(200).
		JSON(map[string]string{})
	prData := PullRequestData{
		Body:     "pull request body",
		TaskName: "Unit Test",
		Title:    "pull request title",
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.CreatePullRequest("unittest", prData)

	require.NoError(t, err)
	require.True(t, gock.IsDone())
}

func TestGitHubRepository_CreatePullRequest_WithAssignees(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Post("/repos/unit/test/pulls").
		MatchType("json").
		JSON(&github.NewPullRequest{
			Base:                github.String("main"),
			Body:                github.String(githubPullRequestBody),
			Head:                github.String("unittest"),
			MaintainerCanModify: github.Bool(true),
			Title:               github.String("pull request title"),
		}).
		Reply(200).
		JSON(&github.PullRequest{
			Number: github.Int(1),
		})
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
	err := repo.CreatePullRequest("unittest", prData)

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
			{Head: &github.PullRequestBranch{Ref: github.String("other")}},
			{Head: &github.PullRequestBranch{Ref: github.String("unittest")}},
		})

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	prId, err := repo.FindPullRequest("unittest")

	require.NoError(t, err)
	assert.IsType(t, &github.PullRequest{}, prId)
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
			{Head: &github.PullRequestBranch{Ref: github.String("other")}},
			{Head: &github.PullRequestBranch{Ref: github.String("some")}},
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

func TestGitHubRepository_GetFile(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/unit/test/contents/test/test.txt").
		MatchParam("ref", "main").
		Reply(200).
		JSON(&github.RepositoryContent{
			Content: github.String("test content"),
		})

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	content, err := repo.GetFile("test/test.txt")

	require.NoError(t, err)
	assert.Equal(t, "test content", content)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_GetPullRequestBody(t *testing.T) {
	pr := &github.PullRequest{
		Body: github.String("pull request body"),
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	body := repo.GetPullRequestBody(pr)

	assert.Equal(t, "pull request body", body)
}

func TestGitHubRepository_GetPullRequestCreationTime(t *testing.T) {
	now := time.Now()
	pr := &github.PullRequest{
		CreatedAt: &github.Timestamp{Time: now},
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	createdAt := repo.GetPullRequestCreationTime(pr)

	assert.Equal(t, now, createdAt)
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
		Number: github.Int(987),
		Head: &github.PullRequestBranch{
			Ref: github.String("unittest"),
		},
	}
	err := repo.DeleteBranch(pr)

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_DeleteBranch_RepoAutoDelete(t *testing.T) {
	defer gock.Off()
	githubRepo := setupGitHubRepository()
	githubRepo.DeleteBranchOnMerge = github.Bool(true)
	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   githubRepo,
	}
	pr := &github.PullRequest{
		Number: github.Int(987),
		Head: &github.PullRequestBranch{
			Ref: github.String("unittest"),
		},
	}
	err := repo.DeleteBranch(pr)

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_DeletePullRequestComment(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Delete("/repos/unit/test/issues/comments/765").
		Reply(200)
	pr := &github.PullRequest{
		Number: github.Int(987),
	}
	comment := PullRequestComment{ID: 765}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.DeletePullRequestComment(comment, pr)

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_HasFile(t *testing.T) {
	defer gock.Off()
	gock.New("https://api.github.com").
		Get("/repos/unit/test/git/trees/main").
		MatchParam("recursive", "1").
		Reply(200).
		JSON(&github.Tree{
			Entries: []*github.TreeEntry{
				{Path: github.String("test.txt")},
				{Path: github.String("config/prod.yaml")},
			},
		})

	gock.New("https://api.github.com").
		Get("/repos/unit/test/git/trees/main").
		MatchParam("recursive", "1").
		Reply(200).
		JSON(&github.Tree{
			Entries: []*github.TreeEntry{
				{Path: github.String("test.txt")},
				{Path: github.String("config/prod.yaml")},
			},
		})

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	success, err := repo.HasFile("**/*.yaml")

	require.NoError(t, err)
	assert.True(t, success)

	fail, err := repo.HasFile("config/staging.yaml")
	require.NoError(t, err)
	assert.False(t, fail)

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
				{Conclusion: github.String("success")},
				{Conclusion: github.String("success")},
			},
		})
	pr := &github.PullRequest{
		Number: github.Int(987),
		Head: &github.PullRequestBranch{
			SHA: github.String("a1b2c3d4"),
		},
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	result, err := repo.HasSuccessfulPullRequestBuild(pr)

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
				{Conclusion: github.String("success")},
				{Conclusion: github.String("failed")},
			},
		})
	pr := &github.PullRequest{
		Number: github.Int(987),
		Head: &github.PullRequestBranch{
			SHA: github.String("a1b2c3d4"),
		},
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	result, err := repo.HasSuccessfulPullRequestBuild(pr)

	require.NoError(t, err)
	assert.False(t, result)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_IsPullRequestClosed(t *testing.T) {
	pr := &github.PullRequest{
		Merged: github.Bool(false),
		State:  github.String("closed"),
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	result := repo.IsPullRequestClosed(pr)

	assert.True(t, result)

	pr = &github.PullRequest{
		Merged: github.Bool(false),
		State:  github.String("open"),
	}
	repo = &GitHubRepository{repo: setupGitHubRepository()}
	result = repo.IsPullRequestClosed(pr)

	assert.False(t, result)
}

func TestGitHubRepository_IsPullRequestMerged(t *testing.T) {
	pr := &github.PullRequest{
		Merged: github.Bool(true),
		State:  github.String("closed"),
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	result := repo.IsPullRequestMerged(pr)

	assert.True(t, result)

	pr = &github.PullRequest{
		Merged: github.Bool(false),
		State:  github.String("open"),
	}
	repo = &GitHubRepository{repo: setupGitHubRepository()}
	result = repo.IsPullRequestMerged(pr)

	assert.False(t, result)
}

func TestGitHubRepository_IsPullRequestOpen(t *testing.T) {
	pr := &github.PullRequest{
		Merged: github.Bool(false),
		State:  github.String("open"),
	}
	repo := &GitHubRepository{repo: setupGitHubRepository()}
	result := repo.IsPullRequestOpen(pr)

	assert.True(t, result)

	pr = &github.PullRequest{
		Merged: github.Bool(true),
		State:  github.String("closed"),
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
			{Body: github.String("comment body"), ID: github.Int64(357)},
		})
	pr := &github.PullRequest{
		Number: github.Int(987),
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	result, err := repo.ListPullRequestComments(pr)

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
		Number: github.Int(987),
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.MergePullRequest(false, pr)

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
		Number: github.Int(987),
		Head: &github.PullRequestBranch{
			Ref: github.String("unittest"),
		},
	}

	repo := &GitHubRepository{
		client: setupGitHubTestClient(),
		repo:   setupGitHubRepository(),
	}
	err := repo.MergePullRequest(true, pr)

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
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
		Body:   github.String("old body"),
		Number: github.Int(987),
		Title:  github.String("old title"),
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
	err := repo.UpdatePullRequest(prData, pr)

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
		Body:   github.String(body),
		Number: github.Int(987),
		Title:  github.String("old title"),
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
	err := repo.UpdatePullRequest(prData, pr)

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
			{Login: github.String("a")},
			{Login: github.String("b")},
			{Login: github.String("c")},
		},
		Body:   github.String(body),
		Number: github.Int(987),
		Title:  github.String("PR title"),
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
	err := repo.UpdatePullRequest(prData, pr)

	require.NoError(t, err)
	assert.True(t, gock.IsDone())
}

func TestGitHubRepository_WebUrl(t *testing.T) {
	githubRepo := setupGitHubRepository()
	githubRepo.HTMLURL = github.String("https://github.com/unit/test")
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
					Owner: &github.User{Login: github.String("unit")},
					Name:  github.String("test"),
				})

			gh := &GitHubHost{
				client: setupGitHubTestClient(),
			}
			repo, err := gh.CreateFromName(tc.input)

			require.NoError(t, err)
			assert.Equal(t, tc.want, repo.FullName())
		})
	}
}

func TestGitHubHost_ListRepositories(t *testing.T) {
	defer gock.Off()
	resultChan := make(chan []Repository)
	errChan := make(chan error)
	gock.New("https://api.github.com").
		Get("/user/repos").
		MatchParams(map[string]string{
			"direction":  "desc",
			"page":       "1",
			"per_page":   "20",
			"sort":       "updated",
			"visibility": "all",
		}).
		Reply(200).
		JSON([]*github.Repository{
			{Owner: &github.User{Login: github.String("unittest")}, Name: github.String("first"), UpdatedAt: &github.Timestamp{Time: time.Now()}},
			{Owner: &github.User{Login: github.String("unittest")}, Name: github.String("second"), UpdatedAt: &github.Timestamp{Time: time.Now()}},
		})

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	go gh.ListRepositories(nil, resultChan, errChan)

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
	assert.Len(t, result, 2)
	assert.Equal(t, "github.com/unittest/first", result[0].FullName())
	assert.Equal(t, "github.com/unittest/second", result[1].FullName())
	assert.True(t, gock.IsDone())
}

func TestGitHubHost_ListRepositories_Since(t *testing.T) {
	defer gock.Off()
	since := time.Now().AddDate(0, 0, -1)
	resultChan := make(chan []Repository)
	errChan := make(chan error)
	gock.New("https://api.github.com").
		Get("/user/repos").
		MatchParams(map[string]string{
			"direction":  "desc",
			"page":       "1",
			"per_page":   "20",
			"sort":       "updated",
			"visibility": "all",
		}).
		Reply(200).
		JSON([]*github.Repository{
			{Owner: &github.User{Login: github.String("unittest")}, Name: github.String("first"), UpdatedAt: &github.Timestamp{Time: time.Now()}},
			{Owner: &github.User{Login: github.String("unittest")}, Name: github.String("second"), UpdatedAt: &github.Timestamp{Time: time.Now()}},
			{Owner: &github.User{Login: github.String("unittest")}, Name: github.String("old"), UpdatedAt: &github.Timestamp{Time: time.Now().AddDate(0, 0, -2)}},
		})

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	go gh.ListRepositories(&since, resultChan, errChan)

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
	assert.Equal(t, "github.com/unittest/first", result[0].FullName())
	assert.Equal(t, "github.com/unittest/second", result[1].FullName())
	assert.True(t, gock.IsDone())
}

func TestGitHubHost_ListRepositoriesWithOpenPullRequests(t *testing.T) {
	defer gock.Off()
	resultChan := make(chan []Repository)
	errChan := make(chan error)
	gock.New("https://api.github.com").
		Get("/user").
		Reply(200).
		JSON(&github.User{
			Login: github.String("unittest"),
		})
	gock.New("https://api.github.com").
		Get("/search/issues").
		MatchParams(map[string]string{
			"page":     "1",
			"per_page": "20",
			"q":        `is:open is:pr author:unittest archived:false`,
		}).
		Reply(200).
		JSON(&github.IssuesSearchResult{
			Issues: []*github.Issue{
				{
					PullRequestLinks: &github.PullRequestLinks{URL: github.String("https://api.github.com/repos/unittest/first/5")},
					RepositoryURL:    github.String("https://api.github.com/repos/unittest/first"),
				},
				{
					PullRequestLinks: &github.PullRequestLinks{URL: github.String("https://api.github.com/repos/unittest/second/10")},
					RepositoryURL:    github.String("https://api.github.com/repos/unittest/second"),
				},
			},
		})
	gock.New("https://api.github.com").
		Get("/repos/unittest/first").
		Reply(200).
		JSON(&github.Repository{Owner: &github.User{Login: github.String("unittest")}, Name: github.String("first")})
	gock.New("https://api.github.com").
		Get("/repos/unittest/second").
		Reply(200).
		JSON(&github.Repository{Owner: &github.User{Login: github.String("unittest")}, Name: github.String("second")})

	gh := &GitHubHost{
		client: setupGitHubTestClient(),
	}
	go gh.ListRepositoriesWithOpenPullRequests(resultChan, errChan)

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
	assert.Equal(t, "github.com/unittest/first", result[0].FullName())
	assert.Equal(t, "github.com/unittest/second", result[1].FullName())
	assert.True(t, gock.IsDone())
}

func setupGitHubTestClient() *github.Client {
	httpClient := &http.Client{}
	gock.InterceptClient(httpClient)
	return github.NewClient(httpClient)
}

func setupGitHubRepository() *github.Repository {
	return &github.Repository{
		DefaultBranch: github.String("main"),
		Name:          github.String("test"),
		Owner: &github.User{
			Login: github.String("unit"),
		},
	}
}
