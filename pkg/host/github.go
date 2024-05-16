package host

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/google/go-github/v59/github"
	"github.com/gregjones/httpcache"
)

var (
	ctx = context.Background()
)

type GitHubRepository struct {
	client *github.Client
	repo   *github.Repository
}

func (g *GitHubRepository) BaseBranch() string {
	return g.repo.GetDefaultBranch()
}

func (g *GitHubRepository) CanMergePullRequest(pr interface{}) (bool, error) {
	gpr := pr.(*github.PullRequest)
	if gpr.Mergeable == nil {
		return true, nil
	}

	return gpr.GetMergeable(), nil
}

func (g *GitHubRepository) CloneUrlHttp() string {
	return g.repo.GetCloneURL()
}

func (g *GitHubRepository) CloneUrlSsh() string {
	return g.repo.GetSSHURL()
}

func (g *GitHubRepository) ClosePullRequest(msg string, pr interface{}) error {
	gpr := pr.(*github.PullRequest)
	comment := &github.IssueComment{Body: github.String(msg)}
	_, _, err := g.client.Issues.CreateComment(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), comment)
	if err != nil {
		return fmt.Errorf("create comment before closing pull request: %w", err)
	}

	gpr.State = github.String("closed")
	_, _, err = g.client.PullRequests.Edit(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), gpr)
	if err != nil {
		return fmt.Errorf("close pull request: %w", err)
	}

	return nil
}

func (g *GitHubRepository) CreatePullRequestComment(body string, pr interface{}) error {
	gpr := pr.(*github.PullRequest)
	comment := &github.IssueComment{Body: github.String(body)}
	_, _, err := g.client.Issues.CreateComment(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), comment)
	if err != nil {
		return fmt.Errorf("create comment on pull request '%d': %w", gpr.GetNumber(), err)
	}

	return nil
}

func (g *GitHubRepository) CreatePullRequest(branch string, data PullRequestData) error {
	body, err := data.GetBody()
	if err != nil {
		return err
	}

	title, err := data.GetTitle()
	if err != nil {
		return err
	}

	gpr := &github.NewPullRequest{
		Base:                g.repo.DefaultBranch,
		Body:                github.String(body),
		Head:                github.String(branch),
		MaintainerCanModify: github.Bool(true),
		Title:               github.String(title),
	}
	pr, _, err := g.client.PullRequests.Create(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr)
	if err != nil {
		return fmt.Errorf("create github pull request: %w", err)
	}

	if len(data.Assignees) > 0 {
		_, _, err := g.client.Issues.AddAssignees(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), int(pr.GetID()), data.Assignees)
		if err != nil {
			return fmt.Errorf("add assignees to pull request: %w", err)
		}
	}

	return nil
}

func (g *GitHubRepository) FindPullRequest(branch string) (any, error) {
	opts := &github.PullRequestListOptions{
		State: "all",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}
	for {
		prs, resp, err := g.client.PullRequests.List(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), opts)
		if err != nil {
			return nil, fmt.Errorf("list pull requests: %w", err)
		}

		for _, pr := range prs {
			if pr.GetHead().GetRef() == branch {
				return pr, nil
			}
		}

		if resp.NextPage == 0 {
			return nil, ErrPullRequestNotFound
		}

		opts.Page = resp.NextPage
	}
}

func (g *GitHubRepository) FullName() string {
	return fmt.Sprintf("github.com/%s/%s", g.repo.GetOwner().GetLogin(), g.repo.GetName())
}

func (g *GitHubRepository) GetFile(fileName string) (string, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: g.BaseBranch(),
	}
	content, _, _, err := g.client.Repositories.GetContents(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), fileName, opts)
	if err != nil {
		return "", fmt.Errorf("get content of file '%s': %w", fileName, err)
	}

	payload, err := content.GetContent()
	if err != nil {
		return "", fmt.Errorf("decode content of file '%s': %w", fileName, err)
	}

	return payload, nil
}

func (g *GitHubRepository) GetPullRequestBody(pr interface{}) string {
	gpr := pr.(*github.PullRequest)
	return gpr.GetBody()
}

func (g *GitHubRepository) GetPullRequestCreationTime(pr interface{}) time.Time {
	gpr := pr.(*github.PullRequest)
	return gpr.GetCreatedAt().Time
}

func (g *GitHubRepository) DeleteBranch(pr interface{}) error {
	// GitHub handles deletion
	if g.repo.GetDeleteBranchOnMerge() {
		return nil
	}

	gpr := pr.(*github.PullRequest)
	_, err := g.client.Git.DeleteRef(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), "heads/"+gpr.GetHead().GetRef())
	if err != nil {
		return fmt.Errorf("delete GitHub branch %s: %w", gpr.GetHead().GetRef(), err)
	}

	return nil
}

func (g *GitHubRepository) DeletePullRequestComment(comment PullRequestComment, pr interface{}) error {
	_, err := g.client.Issues.DeleteComment(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), comment.ID)
	if err != nil {
		return fmt.Errorf("delete pull request comment with ID %d: %w", comment.ID, err)
	}

	return nil
}

func (g *GitHubRepository) HasFile(p string) (bool, error) {
	// TODO test what happens on empty repository
	tree, _, err := g.client.Git.GetTree(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), g.repo.GetDefaultBranch(), true)
	if err != nil {
		return false, fmt.Errorf("get github repository tree: %w", err)
	}

	for _, entry := range tree.Entries {
		matched, err := path.Match(p, entry.GetPath())
		if err != nil {
			return false, fmt.Errorf("malformed pattern '%s': %w", p, err)
		}

		if matched {
			return true, nil
		}
	}

	return false, nil
}

func (g *GitHubRepository) HasSuccessfulPullRequestBuild(pr interface{}) (bool, error) {
	gpr := pr.(*github.PullRequest)
	opts := &github.ListCheckRunsOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}
	for {
		checks, resp, err := g.client.Checks.ListCheckRunsForRef(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetHead().GetSHA(), opts)
		if err != nil {
			return false, fmt.Errorf("list check of github pull request: %w", err)
		}

		for _, run := range checks.CheckRuns {
			if run.GetConclusion() != "success" {
				return false, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return true, nil
}

func (g *GitHubRepository) Host() string {
	if g.client.BaseURL.Host == "api.github.com" {
		return "github.com"
	}

	return g.client.BaseURL.Host
}

func (g *GitHubRepository) IsPullRequestClosed(pr interface{}) bool {
	gpr := pr.(*github.PullRequest)
	return gpr.GetState() == "closed" && !gpr.GetMerged()
}

func (g *GitHubRepository) IsPullRequestMerged(pr interface{}) bool {
	gpr := pr.(*github.PullRequest)
	return gpr.GetState() == "closed" && gpr.GetMerged()
}

func (g *GitHubRepository) IsPullRequestOpen(pr interface{}) bool {
	gpr := pr.(*github.PullRequest)
	return gpr.GetState() == "open"
}

func (g *GitHubRepository) ListPullRequestComments(pr interface{}) ([]PullRequestComment, error) {
	gpr := pr.(*github.PullRequest)
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}
	var pullRequestComments []PullRequestComment
	for {
		comments, resp, err := g.client.Issues.ListComments(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), opts)
		if err != nil {
			return nil, fmt.Errorf("list comments of github pull request %d: %w", gpr.GetNumber(), err)
		}

		for _, comment := range comments {
			pullRequestComments = append(pullRequestComments, PullRequestComment{
				Body: comment.GetBody(),
				ID:   comment.GetID(),
			})
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return pullRequestComments, nil
}

func (g *GitHubRepository) MergePullRequest(deleteBranch bool, pr interface{}) error {
	gpr := pr.(*github.PullRequest)
	_, _, err := g.client.PullRequests.Merge(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), "Auto-merge by saturn-bot", &github.PullRequestOptions{})
	if err != nil {
		return fmt.Errorf("merge github pull request %d: %w", gpr.GetNumber(), err)
	}

	if deleteBranch {
		return g.DeleteBranch(pr)
	}

	return nil
}

func (g *GitHubRepository) Name() string {
	return g.repo.GetName()
}

func (g *GitHubRepository) Owner() string {
	return g.repo.GetOwner().GetLogin()
}

func (g *GitHubRepository) UpdatePullRequest(data PullRequestData, pr interface{}) error {
	gpr := pr.(*github.PullRequest)
	needsUpdate := false
	title, err := data.GetTitle()
	if err != nil {
		return err
	}

	if gpr.GetTitle() != title {
		needsUpdate = true
		gpr.Title = github.String(title)
	}

	body, err := data.GetBody()
	if err != nil {
		return err
	}

	if gpr.GetBody() != body {
		needsUpdate = true
		gpr.Body = github.String(body)
	}

	if needsUpdate {
		_, _, err = g.client.PullRequests.Edit(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), gpr)
		if err != nil {
			return fmt.Errorf("edit github pull request %d: %w", gpr.GetNumber(), err)
		}
	}

	var assigneesToDelete []string
	var currentAssignees []string
	for _, user := range gpr.Assignees {
		currentAssignees = append(currentAssignees, user.GetLogin())
		if !slices.Contains(data.Assignees, user.GetLogin()) {
			assigneesToDelete = append(assigneesToDelete, user.GetLogin())
		}
	}

	if len(assigneesToDelete) > 0 {
		_, _, err = g.client.Issues.RemoveAssignees(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), assigneesToDelete)
		if err != nil {
			return fmt.Errorf("remove assignees from updated pull request %d: %w", gpr.GetNumber(), err)
		}
	}

	var assigneesToAdd []string
	for _, assignee := range data.Assignees {
		if !slices.Contains(currentAssignees, assignee) {
			assigneesToAdd = append(assigneesToAdd, assignee)
		}
	}

	if len(assigneesToAdd) > 0 {
		_, _, err = g.client.Issues.AddAssignees(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), assigneesToAdd)
		if err != nil {
			return fmt.Errorf("add assignees to updated pull request %d: %w", gpr.GetNumber(), err)
		}
	}

	return nil
}

func (g *GitHubRepository) WebUrl() string {
	return g.repo.GetHTMLURL()
}

type GitHubHost struct {
	client *github.Client
}

func NewGitHubHost(address, token string, cacheDisabled bool) (*GitHubHost, error) {
	var httpClient *http.Client
	if cacheDisabled {
		httpClient = &http.Client{}
	} else {
		httpClient = httpcache.NewMemoryCacheTransport().Client()
	}

	httpClient.Timeout = 2 * time.Second
	client := github.NewClient(httpClient)
	client = client.WithAuthToken(token)
	if address != "" {
		var err error
		client, err = client.WithEnterpriseURLs(address, address)
		if err != nil {
			return nil, fmt.Errorf("create github enterprise client: %w", err)
		}
	}

	return &GitHubHost{client: client}, nil
}

func (g *GitHubHost) CreateFromName(name string) (Repository, error) {
	if !strings.HasPrefix(name, "https://") && !strings.HasPrefix(name, "http://") {
		name = "https://" + name
	}

	nameURL, err := url.Parse(name)
	if err != nil {
		return nil, fmt.Errorf("name of repository could not be parsed: %w", err)
	}

	// TODO: How does enterprise server work?
	if g.client.BaseURL.Host != "api."+nameURL.Host {
		return nil, nil
	}

	ownerRepo := strings.TrimPrefix(nameURL.Path, "/")
	ext := path.Ext(ownerRepo)
	if ext != "" {
		ownerRepo = strings.TrimSuffix(ownerRepo, ext)
	}

	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return nil, nil
	}

	repo, _, err := g.client.Repositories.Get(ctx, parts[0], parts[1])
	if err != nil {
		return nil, fmt.Errorf("get github repository: %w", err)
	}

	return &GitHubRepository{repo: repo}, nil
}

func (g *GitHubHost) ListRepositories(since *time.Time, result chan []Repository, errChan chan error) {
	opts := &github.RepositoryListByAuthenticatedUserOptions{
		Affiliation: "owner,collaborator",
		Direction:   "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
		Sort:       "updated",
		Visibility: "all",
	}
	for {
		repos, resp, err := g.client.Repositories.ListByAuthenticatedUser(ctx, opts)
		if err != nil {
			errChan <- fmt.Errorf("list github repositories: %w", err)
			return
		}

		var batch []Repository
		for _, repo := range repos {
			// Return immediately because entries are sorted by "updated"
			if since != nil && repo.GetUpdatedAt().Before(*since) {
				result <- batch // Drain the remaining repositories
				errChan <- nil
				return
			}

			batch = append(batch, &GitHubRepository{client: g.client, repo: repo})
		}

		result <- batch
		if resp.NextPage == 0 {
			errChan <- nil
			return
		}

		opts.Page = resp.NextPage
	}
}

func (g *GitHubHost) ListRepositoriesWithOpenPullRequests(result chan []Repository, errChan chan error) {
	user, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		errChan <- fmt.Errorf("get authenticated user to list repositories: %w", err)
		return
	}

	query := fmt.Sprintf("is:open is:pr author:%s archived:false", user.GetLogin())
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}
	for {
		results, resp, err := g.client.Search.Issues(ctx, query, opts)
		if err != nil {
			errChan <- fmt.Errorf("list repositories with open pull requests: %w", err)
			return
		}

		var batch []Repository
		for _, issue := range results.Issues {
			if issue.IsPullRequest() {
				u, err := url.Parse(issue.GetRepositoryURL())
				if err != nil {
					errChan <- fmt.Errorf("parse repository url of issue %s: %w", issue.GetURL(), err)
					return
				}

				// Search result does not contain data about the associated repository, except for the URL.
				// go-github does not provide a built-in way to extract org/name data from a URL.
				// Need to parse the URL, extract the data and request the repository.
				parts := strings.Split(strings.TrimPrefix(u.Path, "/repos/"), "/")
				repo, _, err := g.client.Repositories.Get(ctx, parts[0], parts[1])
				if err != nil {
					errChan <- fmt.Errorf("get repository of issue %s: %w", issue.GetURL(), err)
					return
				}

				batch = append(batch, &GitHubRepository{client: g.client, repo: repo})
			}
		}

		result <- batch
		if resp.NextPage == 0 {
			errChan <- nil
			break
		}

		opts.Page = resp.NextPage
	}
}
