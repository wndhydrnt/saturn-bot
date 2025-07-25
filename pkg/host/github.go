package host

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/gregjones/httpcache"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/metrics"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

var (
	ctx = context.Background()
)

type GitHubRepository struct {
	client *github.Client
	host   *GitHubHost
	repo   *github.Repository
}

func (g *GitHubRepository) BaseBranch() string {
	return g.repo.GetDefaultBranch()
}

func (g *GitHubRepository) CanMergePullRequest(pr *PullRequest) (bool, error) {
	gpr := pr.Raw.(*github.PullRequest)
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

func (g *GitHubRepository) ClosePullRequest(msg string, pr *PullRequest) (*PullRequest, error) {
	gpr := pr.Raw.(*github.PullRequest)
	comment := &github.IssueComment{Body: github.Ptr(msg)}
	_, _, err := g.client.Issues.CreateComment(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), comment)
	if err != nil {
		return nil, fmt.Errorf("create comment before closing pull request: %w", err)
	}

	gpr.State = github.Ptr("closed")
	gprUpdated, _, err := g.client.PullRequests.Edit(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), gpr)
	if err != nil {
		return nil, fmt.Errorf("close pull request: %w", err)
	}

	return convertGithubPullRequestToPullRequest(gprUpdated), nil
}

func (g *GitHubRepository) CreatePullRequestComment(body string, pr *PullRequest) error {
	gpr := pr.Raw.(*github.PullRequest)
	comment := &github.IssueComment{Body: github.Ptr(body)}
	_, _, err := g.client.Issues.CreateComment(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), comment)
	if err != nil {
		return fmt.Errorf("create comment on pull request '%d': %w", gpr.GetNumber(), err)
	}

	return nil
}

func (g *GitHubRepository) CreatePullRequest(branch string, data PullRequestData) (*PullRequest, error) {
	body, err := data.GetBody()
	if err != nil {
		return nil, err
	}

	gpr := &github.NewPullRequest{
		Base:                g.repo.DefaultBranch,
		Body:                github.Ptr(body),
		Head:                github.Ptr(branch),
		MaintainerCanModify: github.Ptr(true),
		Title:               github.Ptr(data.Title),
	}
	pr, _, err := g.client.PullRequests.Create(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr)
	if err != nil {
		return nil, fmt.Errorf("create github pull request: %w", err)
	}

	if len(data.Assignees) > 0 {
		_, _, err := g.client.Issues.AddAssignees(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), pr.GetNumber(), data.Assignees)
		if err != nil {
			return nil, fmt.Errorf("add assignees to pull request: %w", err)
		}
	}

	if len(data.Reviewers) > 0 {
		_, _, err := g.client.PullRequests.RequestReviewers(
			ctx,
			g.repo.GetOwner().GetLogin(),
			g.repo.GetName(),
			pr.GetNumber(),
			github.ReviewersRequest{Reviewers: data.Reviewers},
		)
		if err != nil {
			return nil, fmt.Errorf("request review for pull request: %w", err)
		}
	}

	return convertGithubPullRequestToPullRequest(pr), nil
}

func (g *GitHubRepository) FindPullRequest(branch string) (*PullRequest, error) {
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
				return convertGithubPullRequestToPullRequest(pr), nil
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

func (g *GitHubRepository) GetPullRequestBody(pr *PullRequest) string {
	gpr := pr.Raw.(*github.PullRequest)
	return gpr.GetBody()
}

func (g *GitHubRepository) DeleteBranch(pr *PullRequest) error {
	// GitHub handles deletion
	if g.repo.GetDeleteBranchOnMerge() {
		return nil
	}

	gpr := pr.Raw.(*github.PullRequest)
	_, err := g.client.Git.DeleteRef(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), "heads/"+gpr.GetHead().GetRef())
	if err != nil {
		return fmt.Errorf("delete GitHub branch %s: %w", gpr.GetHead().GetRef(), err)
	}

	return nil
}

func (g *GitHubRepository) DeletePullRequestComment(comment PullRequestComment, _ *PullRequest) error {
	_, err := g.client.Issues.DeleteComment(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), comment.ID)
	if err != nil {
		return fmt.Errorf("delete pull request comment with ID %d: %w", comment.ID, err)
	}

	return nil
}

func (g *GitHubRepository) HasSuccessfulPullRequestBuild(pr *PullRequest) (bool, error) {
	gpr := pr.Raw.(*github.PullRequest)
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

func (g *GitHubRepository) Host() HostDetail {
	return g.host
}

// ID implements [Repository].
func (g *GitHubRepository) ID() int64 {
	return g.repo.GetID()
}

// IsPullRequestClosed implements [Repository].
// Note: uses MergedAt attribute instead of Merged
// because Merge attribute isn't set when listing pull requests via the GitHub API.
func (g *GitHubRepository) IsPullRequestClosed(pr PullRequestRaw) bool {
	gpr := pr.(*github.PullRequest)
	return isGithubPullRequestClosed(gpr)
}

// IsPullRequestMerged implements [Repository].
// Note: uses MergedAt attribute instead of Merged
// because Merge attribute isn't set when listing pull requests via the GitHub API.
func (g *GitHubRepository) IsPullRequestMerged(pr PullRequestRaw) bool {
	gpr := pr.(*github.PullRequest)
	return isGithubPullRequestMerged(gpr)
}

// IsPullRequestOpen implements [Repository].
func (g *GitHubRepository) IsPullRequestOpen(pr PullRequestRaw) bool {
	gpr := pr.(*github.PullRequest)
	return isGithubPullRequestOpen(gpr)
}

func (g *GitHubRepository) ListPullRequestComments(pr *PullRequest) ([]PullRequestComment, error) {
	gpr := pr.Raw.(*github.PullRequest)
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

func (g *GitHubRepository) MergePullRequest(deleteBranch bool, pr *PullRequest) error {
	gpr := pr.Raw.(*github.PullRequest)
	opts := &github.PullRequestOptions{
		MergeMethod: g.determineMergeMethod(),
	}
	mergeResult, _, err := g.client.PullRequests.Merge(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), "Auto-merge by saturn-bot", opts)
	if err != nil {
		return fmt.Errorf("merge github pull request %d: %w", gpr.GetNumber(), err)
	}

	gpr.Merged = mergeResult.Merged

	// Don't delete if DeleteBranchOnMerge == true.
	// GitHub deletes the branch on its own.
	if deleteBranch && !g.repo.GetDeleteBranchOnMerge() {
		return g.DeleteBranch(pr)
	}

	return nil
}

func (g *GitHubRepository) determineMergeMethod() string {
	if g.repo.GetAllowSquashMerge() {
		return "squash"
	}

	if g.repo.GetAllowRebaseMerge() {
		return "rebase"
	}

	if g.repo.GetAllowMergeCommit() {
		return "merge"
	}

	return ""
}

func (g *GitHubRepository) Name() string {
	return g.repo.GetName()
}

func (g *GitHubRepository) Owner() string {
	return g.repo.GetOwner().GetLogin()
}

func (g *GitHubRepository) UpdatePullRequest(data PullRequestData, pr *PullRequest) error {
	gpr := pr.Raw.(*github.PullRequest)
	needsUpdate := false
	if gpr.GetTitle() != data.Title {
		needsUpdate = true
		gpr.Title = github.Ptr(data.Title)
	}

	body, err := data.GetBody()
	if err != nil {
		return err
	}

	if gpr.GetBody() != body {
		needsUpdate = true
		gpr.Body = github.Ptr(body)
	}

	if needsUpdate {
		_, _, err = g.client.PullRequests.Edit(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), gpr)
		if err != nil {
			return fmt.Errorf("edit github pull request %d: %w", gpr.GetNumber(), err)
		}
	}

	if len(data.Assignees) > 0 {
		assigneesToAdd, assigneesToRemove := diffAssignees(gpr.Assignees, data.Assignees)
		if len(assigneesToRemove) > 0 {
			_, _, err = g.client.Issues.RemoveAssignees(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), assigneesToRemove)
			if err != nil {
				return fmt.Errorf("remove assignees from updated pull request %d: %w", gpr.GetNumber(), err)
			}
		}

		if len(assigneesToAdd) > 0 {
			_, _, err = g.client.Issues.AddAssignees(ctx, g.repo.GetOwner().GetLogin(), g.repo.GetName(), gpr.GetNumber(), assigneesToAdd)
			if err != nil {
				return fmt.Errorf("add assignees to updated pull request %d: %w", gpr.GetNumber(), err)
			}
		}
	}

	if len(data.Reviewers) > 0 {
		reviews, err := g.listAllReviews(gpr.GetNumber())
		if err != nil {
			return err
		}

		var submittedReviewers []*github.User
		for _, review := range reviews {
			submittedReviewers = append(submittedReviewers, review.User)
		}

		reviewersToAdd, reviewersToRemove := diffReviewers(gpr.RequestedReviewers, submittedReviewers, data.Reviewers)
		if len(reviewersToAdd) > 0 {
			_, _, err := g.client.PullRequests.RequestReviewers(
				ctx,
				g.repo.GetOwner().GetLogin(),
				g.repo.GetName(),
				gpr.GetNumber(),
				github.ReviewersRequest{Reviewers: reviewersToAdd},
			)
			if err != nil {
				return fmt.Errorf("update to add requested reviewers on pull request %d: %w", gpr.GetNumber(), err)
			}
		}

		if len(reviewersToRemove) > 0 {
			_, err := g.client.PullRequests.RemoveReviewers(
				ctx,
				g.repo.GetOwner().GetLogin(),
				g.repo.GetName(),
				gpr.GetNumber(),
				github.ReviewersRequest{Reviewers: reviewersToRemove},
			)
			if err != nil {
				return fmt.Errorf("update to remove requested reviewers from pull request %d: %w", gpr.GetNumber(), err)
			}
		}
	}

	return nil
}

func (g *GitHubRepository) WebUrl() string {
	return g.repo.GetHTMLURL()
}

// Raw implements [Repository].
func (g *GitHubRepository) Raw() any {
	return g.repo
}

// IsArchived implements [Repository].
func (g *GitHubRepository) IsArchived() bool {
	return g.repo.GetArchived()
}

// UpdatedAt implements [Repository].
func (g *GitHubRepository) UpdatedAt() time.Time {
	return g.repo.UpdatedAt.Time
}

// listAllReviews lists all reviews done for a pull request.
// The function is necessary because the GitHub API removes a user from the list of "requested reviewers"
// and adds the user to the list of reviews.
// The function is used as part of the feature to set reviewers of a pull request.
func (g *GitHubRepository) listAllReviews(prNumber int) ([]*github.PullRequestReview, error) {
	opts := &github.ListOptions{
		Page:    1,
		PerPage: 20,
	}
	var reviews []*github.PullRequestReview
	for {
		batch, resp, err := g.client.PullRequests.ListReviews(
			ctx,
			g.repo.GetOwner().GetLogin(),
			g.repo.GetName(),
			prNumber,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("list reviews of pull request %d: %w", prNumber, err)
		}

		reviews = append(reviews, batch...)

		if resp.NextPage == 0 {
			return reviews, nil
		}

		opts.Page = resp.NextPage
	}
}

func diffAssignees(current []*github.User, want []string) (toAdd, toRemove []string) {
	var currentLogins []string
	for _, user := range current {
		currentLogins = append(currentLogins, user.GetLogin())
		if !slices.Contains(want, user.GetLogin()) {
			toRemove = append(toRemove, user.GetLogin())
		}
	}

	for _, assignee := range want {
		if !slices.Contains(currentLogins, assignee) {
			toAdd = append(toAdd, assignee)
		}
	}

	// sort to ensure that slices.Compact() catches all duplicates
	slices.Sort(toAdd)
	slices.Sort(toRemove)
	return slices.Compact(toAdd), slices.Compact(toRemove)
}

func diffReviewers(requested, submitted []*github.User, want []string) (toAdd, toRemove []string) {
	// Normalize the list of requested reviewers by adding the users that have already
	// submitted a review.
	// This is done to not request a review from users again if they have submitted one already.
	for _, user := range submitted {
		if slices.Contains(want, user.GetLogin()) {
			requested = append(requested, user)
		}
	}

	return diffAssignees(requested, want)
}

type GitHubHost struct {
	authenticatedUser *UserInfo
	client            *github.Client
}

func NewGitHubHost(address, token string, cacheDisabled bool) (*GitHubHost, error) {
	retryingClient := retryablehttp.NewClient()
	retryingClient.Backoff = githubBackoff
	retryingClient.CheckRetry = githubRetryPolicy
	retryingClient.RequestLogHook = func(_ retryablehttp.Logger, r *http.Request, i int) {
		// 0 is the initial request
		if i > 0 {
			log.Log().Infof("GitHub retry request attempt %d to %s", i, r.URL.String())
		}
	}

	httpClient := retryingClient.StandardClient()
	httpClient.Timeout = 2 * time.Second
	// Set up metrics first, then add the caching layer.
	// Makes the caching layer execute before the metrics.
	// If it reads from cache then those calls aren't counted
	// as requests.
	metrics.InstrumentHttpClient(httpClient)
	if !cacheDisabled {
		transport := httpcache.NewMemoryCacheTransport()
		transport.Transport = httpClient.Transport
		httpClient.Transport = transport
	}

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

func (g *GitHubHost) AuthenticatedUser() (*UserInfo, error) {
	if g.authenticatedUser != nil {
		return g.authenticatedUser, nil
	}

	user, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get current github user: %w", err)
	}

	var userEmail string
	opts := &github.ListOptions{
		Page:    1,
		PerPage: 20,
	}
	for {
		emails, resp, err := g.client.Users.ListEmails(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("list email of authenticated user: %w", err)
		}

		for _, email := range emails {
			if email.GetVisibility() == "public" {
				userEmail = email.GetEmail()
			}
		}

		if userEmail != "" {
			break
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	if userEmail == "" {
		return nil, fmt.Errorf("no public email address for user %s", user.GetLogin())
	}

	log.Log().Debug("Discovered authenticated user from GitHub")
	g.authenticatedUser = &UserInfo{
		Email: userEmail,
		Name:  user.GetLogin(),
	}
	return g.authenticatedUser, nil
}

// CreateFromJson implements [Host].
func (g *GitHubHost) CreateFromJson(dec *json.Decoder) (Repository, error) {
	ghRepo := &github.Repository{}
	err := dec.Decode(ghRepo)
	if err != nil {
		return nil, fmt.Errorf("decode GitHub repository from JSON: %w", err)
	}

	return &GitHubRepository{client: g.client, host: g, repo: ghRepo}, nil
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

	return &GitHubRepository{client: g.client, host: g, repo: repo}, nil
}

func (g *GitHubHost) Name() string {
	if g.client.BaseURL.Host == "api.github.com" {
		return "github.com"
	}

	return g.client.BaseURL.Host
}

// Type implements [Host].
func (g *GitHubHost) Type() Type {
	return GitHubType
}

func (g *GitHubHost) RepositoryIterator() RepositoryIterator {
	return &githubRepositoryIterator{host: g}
}

// PullRequestFactory implements [Host].
func (g *GitHubHost) PullRequestFactory() PullRequestFactory {
	return func() any {
		return &github.PullRequest{}
	}
}

// PullRequestIterator implements [Host].
func (g *GitHubHost) PullRequestIterator() PullRequestIterator {
	return &githubPullRequestIterator{client: g.client}
}

type githubPullRequestIterator struct {
	client *github.Client
	err    error
}

func (it *githubPullRequestIterator) ListPullRequests(since *time.Time) iter.Seq[*PullRequest] {
	return func(yield func(*PullRequest) bool) {
		user, _, err := it.client.Users.Get(ctx, "")
		if err != nil {
			it.err = fmt.Errorf("get authenticated user to list pull requests: %w", err)
			return
		}

		query := fmt.Sprintf("is:pr author:%s archived:false", user.GetLogin())
		opts := &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 20,
			},
			Sort:  "updated",
			Order: "desc",
		}
		for {
			results, resp, err := it.client.Search.Issues(ctx, query, opts)
			if err != nil {
				it.err = fmt.Errorf("list pull requests page %d: %w", opts.Page, err)
				return
			}

			for _, issue := range results.Issues {
				if !issue.IsPullRequest() {
					continue
				}

				if since != nil && issue.GetUpdatedAt().Before(ptr.From(since)) {
					return
				}

				prRequest, err := it.client.NewRequest(http.MethodGet, issue.PullRequestLinks.GetURL(), nil)
				if err != nil {
					it.err = fmt.Errorf("create request for pull request from url %s: %w", issue.PullRequestLinks.GetURL(), err)
					return
				}

				gpr := &github.PullRequest{}
				_, err = it.client.Do(context.Background(), prRequest, gpr)
				if err != nil {
					it.err = fmt.Errorf("get pull request from url %s: %w", issue.PullRequestLinks.GetURL(), err)
					return
				}

				pr := convertGithubPullRequestToPullRequest(gpr)
				if !yield(pr) {
					return
				}
			}

			if resp.NextPage == 0 {
				return
			}

			opts.Page = resp.NextPage
		}
	}
}

func (it *githubPullRequestIterator) Error() error {
	return it.err
}

type githubRepositoryIterator struct {
	err  error
	host *GitHubHost
}

func (it *githubRepositoryIterator) ListRepositories(since *time.Time) iter.Seq[Repository] {
	return func(yield func(Repository) bool) {
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
			repos, resp, err := it.host.client.Repositories.ListByAuthenticatedUser(ctx, opts)
			if err != nil {
				it.err = fmt.Errorf("list github repositories: %w", err)
				return
			}

			for _, repo := range repos {
				// Return immediately because entries are sorted by "updated"
				if since != nil && repo.GetUpdatedAt().Before(*since) {
					return
				}

				// Get the repository again because ListByAuthenticatedUser doesn't return a full repository object.
				repoFull, _, err := it.host.client.Repositories.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName())
				if err != nil {
					it.err = fmt.Errorf("get github repository %s/%s: %w", repo.GetOwner().GetLogin(), repo.GetName(), err)
					return
				}

				sbRepo := &GitHubRepository{client: it.host.client, host: it.host, repo: repoFull}
				if !yield(sbRepo) {
					return
				}
			}

			if resp.NextPage == 0 {
				return
			}

			opts.Page = resp.NextPage
		}
	}
}

func (it *githubRepositoryIterator) Error() error {
	return it.err
}

func convertGithubPullRequestToPullRequest(gpr *github.PullRequest) *PullRequest {
	u, _ := url.Parse(gpr.GetHead().GetRepo().GetHTMLURL())

	return &PullRequest{
		CreatedAt:      gpr.GetCreatedAt().Time,
		Number:         int64(gpr.GetNumber()),
		WebURL:         gpr.GetHTMLURL(),
		State:          mapGithubPrToPullRequestState(gpr),
		Raw:            gpr,
		HostName:       u.Host,
		BranchName:     gpr.GetHead().GetRef(),
		RepositoryName: fmt.Sprintf("%s%s", u.Host, u.Path),
		Type:           GitHubType,
	}
}

// githubBackoff is a [github.com/hashicorp/go-retryablehttp.RetryPolicy].
func githubRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	ts := readGitHubRateLimitResetTime(resp)
	if ts != "" {
		return true, nil
	}

	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}

// githubBackoff is a [github.com/hashicorp/go-retryablehttp.Backoff].
func githubBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	tsRaw := readGitHubRateLimitResetTime(resp)
	if tsRaw != "" {
		i, err := strconv.ParseInt(tsRaw, 10, 64)
		if err != nil {
			return retryablehttp.DefaultBackoff(min, max, attemptNum, resp)
		}

		return time.Unix(i, 0).Sub(time.Now().UTC())
	}

	return retryablehttp.DefaultBackoff(min, max, attemptNum, resp)
}

func isGithubPullRequestClosed(pr *github.PullRequest) bool {
	return pr.GetState() == "closed" && pr.MergedAt == nil
}

func isGithubPullRequestMerged(pr *github.PullRequest) bool {
	return pr.GetState() == "closed" && pr.MergedAt != nil
}

func isGithubPullRequestOpen(pr *github.PullRequest) bool {
	return pr.GetState() == "open"
}

func mapGithubPrToPullRequestState(pr *github.PullRequest) PullRequestState {
	if isGithubPullRequestClosed(pr) {
		return PullRequestStateClosed
	}

	if isGithubPullRequestMerged(pr) {
		return PullRequestStateMerged
	}

	if isGithubPullRequestOpen(pr) {
		return PullRequestStateOpen
	}

	return PullRequestStateUnknown
}

// readGitHubRateLimitResetTime reads the timestamp returned by the GitHub API if a rate-limit is active.
// See https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28#exceeding-the-rate-limit
func readGitHubRateLimitResetTime(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		// Not checking header Retry-After because go-retryablehttp handles that.
		remaining := resp.Header.Get("x-ratelimit-remaining")
		if remaining != "0" {
			return ""
		}

		return resp.Header.Get("x-ratelimit-reset")
	}

	return ""
}
