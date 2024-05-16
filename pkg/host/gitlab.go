package host

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/xanzy/go-gitlab"
)

// userCache caches GitLab users.
// The cache prevents frequent requests for users, for example when finding assignees, from triggering rate limits.
type userCache struct {
	client *gitlab.Client
	data   map[string]*gitlab.User
	dataMu sync.RWMutex
}

func (u *userCache) get(name string) (*gitlab.User, error) {
	u.dataMu.RLock()
	user, ok := u.data[name]
	u.dataMu.RUnlock()
	if ok {
		return user, nil
	}

	users, _, err := u.client.Users.ListUsers(&gitlab.ListUsersOptions{
		Username: gitlab.Ptr(name),
	})
	if err != nil {
		return nil, fmt.Errorf("get user from GitLab: %w", err)
	}

	u.dataMu.Lock()
	u.data[name] = users[0]
	u.dataMu.Unlock()
	return users[0], nil
}

type GitLabRepository struct {
	client    *gitlab.Client
	fullName  string
	project   *gitlab.Project
	userCache *userCache
}

func (g *GitLabRepository) BaseBranch() string {
	return g.project.DefaultBranch
}

func (g *GitLabRepository) CloneUrlHttp() string {
	return g.project.HTTPURLToRepo
}

func (g *GitLabRepository) CloneUrlSsh() string {
	return g.project.SSHURLToRepo
}

func (g *GitLabRepository) FullName() string {
	if g.fullName == "" {
		u, _ := url.Parse(g.project.WebURL)
		g.fullName = fmt.Sprintf("%s%s", u.Host, u.Path)
	}

	return g.fullName
}

func (g *GitLabRepository) GetPullRequestBody(pr interface{}) string {
	mr := pr.(*gitlab.MergeRequest)
	return mr.Description
}

func (g *GitLabRepository) GetPullRequestCreationTime(pr interface{}) time.Time {
	mr := pr.(*gitlab.MergeRequest)
	return *mr.CreatedAt
}

func (g *GitLabRepository) WebUrl() string {
	return g.project.WebURL
}

func (g *GitLabRepository) CanMergePullRequest(pr interface{}) (bool, error) {
	return true, nil
}

func (g *GitLabRepository) ClosePullRequest(msg string, pr interface{}) error {
	mr := pr.(*gitlab.MergeRequest)
	_, _, err := g.client.Notes.CreateMergeRequestNote(
		g.project.ID,
		mr.IID,
		&gitlab.CreateMergeRequestNoteOptions{
			Body: gitlab.Ptr(msg),
		},
	)
	if err != nil {
		return fmt.Errorf("create note for gitlab merge request %d: %w", mr.IID, err)
	}

	_, _, err = g.client.MergeRequests.UpdateMergeRequest(
		g.project.ID,
		mr.IID,
		&gitlab.UpdateMergeRequestOptions{
			StateEvent: gitlab.Ptr("close"),
		},
	)
	if err != nil {
		return fmt.Errorf("close gitlab merge request %d: %w", mr.IID, err)
	}

	return nil
}

func (g *GitLabRepository) CreatePullRequestComment(body string, pr interface{}) error {
	mr := pr.(*gitlab.MergeRequest)
	_, _, err := g.client.Notes.CreateMergeRequestNote(
		g.project.ID,
		mr.IID,
		&gitlab.CreateMergeRequestNoteOptions{
			Body: gitlab.Ptr(body),
		},
	)
	if err != nil {
		return fmt.Errorf("create note for gitlab merge request %d: %w", mr.IID, err)
	}

	return nil
}

func (g *GitLabRepository) CreatePullRequest(branch string, data PullRequestData) error {
	description, err := data.GetBody()
	if err != nil {
		return fmt.Errorf("create merge request for gitlab: %w", err)
	}

	title, err := data.GetTitle()
	if err != nil {
		return fmt.Errorf("create merge request for gitlab: %w", err)
	}

	var assigneeIDs []int
	for _, assignee := range data.Assignees {
		user, err := g.userCache.get(assignee)
		if err != nil {
			slog.Warn("Cannot find assignee in GitLab to add to new merge request", "assignee", assignee, "err", err)
			continue
		}

		assigneeIDs = append(assigneeIDs, user.ID)
	}

	var labels *gitlab.LabelOptions
	if len(data.Labels) > 0 {
		labels = gitlab.Ptr(gitlab.LabelOptions(data.Labels))
	}
	_, _, err = g.client.MergeRequests.CreateMergeRequest(
		g.project.ID,
		&gitlab.CreateMergeRequestOptions{
			AssigneeIDs:  &assigneeIDs,
			Description:  gitlab.Ptr(description),
			Labels:       labels,
			SourceBranch: gitlab.Ptr(branch),
			TargetBranch: gitlab.Ptr(g.project.DefaultBranch),
			Title:        gitlab.Ptr(title),
		},
	)
	if err != nil {
		return fmt.Errorf("create merge request for project %d: %w", g.project.ID, err)
	}

	return nil
}

func (g *GitLabRepository) DeleteBranch(pr interface{}) error {
	mr := pr.(*gitlab.MergeRequest)
	if mr.ShouldRemoveSourceBranch {
		return nil
	}

	_, err := g.client.Branches.DeleteBranch(g.project.ID, mr.SourceBranch)
	if err != nil {
		return fmt.Errorf("delete branch %s in gitlab of project %d: %w", mr.SourceBranch, g.project.ID, err)
	}

	return nil
}

func (g *GitLabRepository) DeletePullRequestComment(comment PullRequestComment, pr interface{}) error {
	mr := pr.(*gitlab.MergeRequest)
	_, err := g.client.Notes.DeleteMergeRequestNote(g.project.ID, mr.IID, int(comment.ID))
	if err != nil {
		return fmt.Errorf("delete note of gitlab merge request %d project %d: %w", mr.IID, g.project.ID, err)
	}

	return nil
}

func (g *GitLabRepository) FindPullRequest(branch string) (any, error) {
	mrs, _, err := g.client.MergeRequests.ListProjectMergeRequests(
		g.project.ID,
		&gitlab.ListProjectMergeRequestsOptions{
			State:        gitlab.Ptr("all"),
			SourceBranch: gitlab.Ptr(branch),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list merge requests: %w", err)
	}

	if len(mrs) == 0 {
		return nil, ErrPullRequestNotFound
	}

	return mrs[0], nil
}

func (g *GitLabRepository) GetFile(fileName string) (string, error) {
	file, _, err := g.client.RepositoryFiles.GetFile(g.project.ID, fileName, &gitlab.GetFileOptions{Ref: gitlab.Ptr(g.project.DefaultBranch)})
	if err != nil {
		var errResp *gitlab.ErrorResponse
		if errors.As(err, &errResp) && errResp.Response.StatusCode == http.StatusNotFound {
			return "", ErrFileNotFound
		}

		return "", fmt.Errorf("get file %s: %w", fileName, err)
	}

	if file.Encoding == "base64" {
		b, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return "", fmt.Errorf("decode base64-encoded content of file %s: %w", file.FileName, err)
		}

		return string(b), nil
	}

	return file.Content, nil
}

func (g *GitLabRepository) HasFile(p string) (bool, error) {
	dir := path.Dir(p)
	opts := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
		Path: gitlab.Ptr(dir),
	}

	for {
		tree, resp, err := g.client.Repositories.ListTree(
			g.project.ID,
			opts,
		)
		if err != nil {
			var errResp *gitlab.ErrorResponse
			if errors.As(err, &errResp) {
				if errResp.Response.StatusCode == http.StatusNotFound {
					slog.Warn("Tree not found - empty repository?")
					return false, nil
				}
			}
			return false, fmt.Errorf("list tree of repository %d: %w", g.project.ID, err)
		}

		for _, entry := range tree {
			if entry.Type != "blob" {
				continue
			}

			matched, err := path.Match(p, entry.Path)
			if err != nil {
				return false, fmt.Errorf("match file pattern %s: %w", p, err)
			}

			if matched {
				return true, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return false, nil
}

func (g *GitLabRepository) HasSuccessfulPullRequestBuild(pr interface{}) (bool, error) {
	mr := pr.(*gitlab.MergeRequest)
	state, _, err := g.client.MergeRequestApprovals.GetApprovalState(g.project.ID, mr.IID)
	if err != nil {
		return false, fmt.Errorf("get approval state of gitlab merge request %d project %d: %w", mr.IID, g.project.ID, err)
	}

	for _, rule := range state.Rules {
		if !rule.Approved {
			return false, nil
		}
	}

	opts := &gitlab.GetCommitStatusesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 10,
		},
	}

	failed := false
	for {
		statuses, resp, err := g.client.Commits.GetCommitStatuses(g.project.ID, mr.SHA, opts)
		if err != nil {
			return false, fmt.Errorf("get commit status for sha %s of gitlab merge request %d project %d: %w", mr.SHA, mr.IID, g.project.ID, err)
		}

		for _, status := range statuses {
			if status.AllowFailure {
				continue
			}

			if status.Status != "success" {
				failed = true
			}
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return !failed, nil
}

func (g *GitLabRepository) Host() string {
	return g.client.BaseURL().Host
}

func (g *GitLabRepository) IsPullRequestClosed(pr interface{}) bool {
	mr := pr.(*gitlab.MergeRequest)
	return mr.State == "closed" || mr.State == "locked"
}

func (g *GitLabRepository) IsPullRequestMerged(pr interface{}) bool {
	mr := pr.(*gitlab.MergeRequest)
	return mr.State == "merged"
}

func (g *GitLabRepository) IsPullRequestOpen(pr interface{}) bool {
	mr := pr.(*gitlab.MergeRequest)
	return mr.State == "opened"
}

func (g *GitLabRepository) ListPullRequestComments(pr interface{}) ([]PullRequestComment, error) {
	var result []PullRequestComment
	if pr == nil {
		return result, nil
	}

	mr := pr.(*gitlab.MergeRequest)
	opts := &gitlab.ListMergeRequestNotesOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 10,
		},
	}
	for {
		notes, resp, err := g.client.Notes.ListMergeRequestNotes(g.project.ID, mr.IID, opts)
		if err != nil {
			return result, fmt.Errorf("list notes of gitlab merge request %d project %d: %w", mr.IID, g.project.ID, err)
		}

		for _, note := range notes {
			result = append(result, PullRequestComment{Body: note.Body, ID: int64(note.ID)})
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return result, nil
}

func (g *GitLabRepository) MergePullRequest(deleteBranch bool, pr interface{}) error {
	mr := pr.(*gitlab.MergeRequest)
	_, _, err := g.client.MergeRequests.AcceptMergeRequest(
		g.project.ID,
		mr.IID,
		&gitlab.AcceptMergeRequestOptions{ShouldRemoveSourceBranch: gitlab.Ptr(deleteBranch), Squash: &mr.Squash},
	)
	if err != nil {
		return fmt.Errorf("merge merge request %d: %w", mr.IID, err)
	}

	return nil
}

func (g *GitLabRepository) Name() string {
	return g.project.Name
}

func (g *GitLabRepository) Owner() string {
	return g.project.Namespace.FullPath
}

func (g *GitLabRepository) UpdatePullRequest(data PullRequestData, pr interface{}) error {
	needsUpdate := false
	mr := pr.(*gitlab.MergeRequest)
	title, err := data.GetTitle()
	if err != nil {
		return fmt.Errorf("get title to update gitlab merge request: %w", err)
	}

	if mr.Title != title {
		needsUpdate = true
	}

	body, err := data.GetBody()
	if err != nil {
		return fmt.Errorf("get body to update gitlab merge request: %w", err)
	}

	if mr.Description != body {
		needsUpdate = true
	}

	assignedAssignees := map[string]*gitlab.BasicUser{}
	for _, user := range mr.Assignees {
		assignedAssignees[user.Username] = user
	}

	var assigneeIDs []int
	for _, assignee := range data.Assignees {
		assignedAssignee, ok := assignedAssignees[assignee]
		if ok {
			assigneeIDs = append(assigneeIDs, assignedAssignee.ID)
		} else {
			user, err := g.userCache.get(assignee)
			if err != nil {
				slog.Warn("Cannot find assignee in GitLab to update merge request", "assignee", assignee, "err", err)
				continue
			}

			needsUpdate = true
			assigneeIDs = append(assigneeIDs, user.ID)
		}
	}

	if needsUpdate {
		_, _, err = g.client.MergeRequests.UpdateMergeRequest(
			g.project.ID,
			mr.IID,
			&gitlab.UpdateMergeRequestOptions{
				AssigneeIDs: gitlab.Ptr(assigneeIDs),
				Description: gitlab.Ptr(body),
				Title:       gitlab.Ptr(title),
			},
		)
		if err != nil {
			return fmt.Errorf("update gitlab merge request %d project %d: %w", mr.IID, g.project.ID, err)
		}
	}

	return nil
}

type GitLabHost struct {
	client    *gitlab.Client
	userCache *userCache
}

func NewGitLabHost(token string) (*GitLabHost, error) {
	client, err := gitlab.NewClient(token)
	if err != nil {
		return nil, fmt.Errorf("initialize gitlab client: %w", err)
	}

	return &GitLabHost{
		client: client,
		userCache: &userCache{
			client: client,
			data:   map[string]*gitlab.User{},
		},
	}, nil
}

func (g *GitLabHost) CreateFromName(name string) (Repository, error) {
	if !strings.HasPrefix(name, "https://") && !strings.HasPrefix(name, "http://") {
		name = "https://" + name
	}

	nameURL, err := url.Parse(name)
	if err != nil {
		return nil, fmt.Errorf("name of repository could not be parsed: %w", err)
	}

	if nameURL.Host != g.client.BaseURL().Host {
		return nil, nil
	}

	projectID := strings.TrimPrefix(nameURL.Path, "/")
	ext := path.Ext(nameURL.Path)
	if ext != "" {
		projectID = strings.TrimSuffix(projectID, ext)
	}

	project, _, err := g.client.Projects.GetProject(projectID, &gitlab.GetProjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get gitlab project: %w", err)
	}

	return &GitLabRepository{client: g.client, project: project, userCache: g.userCache}, nil
}

func (g *GitLabHost) ListRepositories(since *time.Time, result chan []Repository, errChan chan error) {
	opts := &gitlab.ListProjectsOptions{
		Archived:          gitlab.Ptr(false),
		LastActivityAfter: since,
		MinAccessLevel:    gitlab.Ptr(gitlab.AccessLevelValue(30)),
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}
	for {
		projects, resp, err := g.client.Projects.ListProjects(opts)
		if err != nil {
			errChan <- fmt.Errorf("list gitlab projects: %w", err)
			return
		}

		var batch []Repository
		for _, project := range projects {
			batch = append(batch, &GitLabRepository{client: g.client, project: project, userCache: g.userCache})
		}

		result <- batch
		if resp.NextPage == 0 {
			errChan <- nil
			return
		}

		opts.Page = resp.NextPage
	}
}

func (g *GitLabHost) ListRepositoriesWithOpenPullRequests(result chan []Repository, errChan chan error) {
	user, _, err := g.client.Users.CurrentUser()
	if err != nil {
		errChan <- fmt.Errorf("list current gitlab user: %w", err)
		return
	}

	opts := &gitlab.ListMergeRequestsOptions{
		AuthorID: &user.ID,
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}
	visitedProjectsIDs := map[int]struct{}{}
	for {
		mergeRequests, resp, err := g.client.MergeRequests.ListMergeRequests(opts)
		if err != nil {
			errChan <- fmt.Errorf("list gitlab projects with open merge requests: %w", err)
			return
		}

		var batch []Repository
		for _, mergeRequest := range mergeRequests {
			_, exists := visitedProjectsIDs[mergeRequest.ProjectID]
			if exists {
				continue
			}

			visitedProjectsIDs[mergeRequest.ProjectID] = struct{}{}
			project, _, err := g.client.Projects.GetProject(mergeRequest.ProjectID, &gitlab.GetProjectOptions{})
			if err != nil {
				errChan <- fmt.Errorf("get gitlab project %d with open merge request %d: %w", mergeRequest.ProjectID, mergeRequest.IID, err)
				return
			}

			if project.Archived {
				slog.Debug("Ignore project because it has been archived", "project", project.ID)
				continue
			}

			batch = append(batch, &GitLabRepository{client: g.client, project: project, userCache: g.userCache})
		}

		result <- batch

		if resp.NextPage == 0 {
			errChan <- nil
			return
		}

		opts.Page = resp.NextPage
	}
}
