package host

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"strings"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/template"
)

var (
	ErrFileNotFound        = errors.New("file not found")
	ErrPullRequestNotFound = errors.New("pull request not found")

	tplPrBodyDefault = htmlTemplate.Must(htmlTemplate.New("bodyDefault").Parse("Apply changes from task {{.TaskName}}."))
)

// PullRequest holds data on an existing pull request.
type PullRequest struct {
	// CreatedAt is the time and date at which the pull request has been created.
	CreatedAt *time.Time
	// Number is the identifier of the pull request.
	Number int64
	// WebURL is the URL humans visit to view the pull request.
	WebURL string
}

type PullRequestComment struct {
	Body string
	ID   int64
}

type PullRequestData struct {
	Assignees      []string
	AutoMerge      bool
	AutoMergeAfter time.Duration
	Body           string
	Labels         []string
	MergeOnce      bool
	Reviewers      []string
	TaskName       string
	TemplateData   template.Data
	Title          string
}

func (prd PullRequestData) GetBody() (string, error) {
	var bodyTpl *htmlTemplate.Template
	if prd.Body == "" {
		bodyTpl = tplPrBodyDefault
	} else {
		var err error
		bodyTpl, err = htmlTemplate.New("body").Parse(prd.Body)
		if err != nil {
			return "", fmt.Errorf("parse body template of task %s: %w", prd.TaskName, err)
		}
	}

	buf := &bytes.Buffer{}
	err := bodyTpl.Execute(buf, prd.TemplateData)
	if err != nil {
		return "", fmt.Errorf("execute body template of task %s: %w", prd.TaskName, err)
	}

	var autoMergeText string
	if prd.AutoMerge {
		if prd.AutoMergeAfter == 0 {
			autoMergeText = "Enabled. Saturn merges this automatically on its next run and if all checks have passed."
		} else {
			autoMergeText = fmt.Sprintf("Enabled. Saturn automatically merges this in %s and if all checks have passed.", prd.AutoMergeAfter.String())
		}
	} else {
		autoMergeText = "Disabled. Merge this manually."
	}

	var ignoreText string
	if prd.MergeOnce {
		ignoreText = "Close this PR and it will not be recreated again."
	} else {
		ignoreText = "This PR will be recreated if closed."
	}

	content, err := template.RenderPullRequestDescription(template.PullRequestDescriptionInput{
		AutoMergeText: autoMergeText,
		Body:          buf.String(),
		IgnoreText:    ignoreText,
	})
	if err != nil {
		return "", err
	}

	return content, nil
}

type Repository interface {
	BaseBranch() string
	CanMergePullRequest(pr interface{}) (bool, error)
	CloneUrlHttp() string
	CloneUrlSsh() string
	ClosePullRequest(msg string, pr interface{}) error
	CreatePullRequestComment(body string, pr interface{}) error
	CreatePullRequest(branch string, data PullRequestData) (*PullRequest, error)
	DeleteBranch(pr interface{}) error
	DeletePullRequestComment(comment PullRequestComment, pr interface{}) error
	FindPullRequest(branch string) (any, error)
	FullName() string
	GetPullRequestBody(pr interface{}) string
	GetPullRequestCreationTime(pr interface{}) time.Time
	HasSuccessfulPullRequestBuild(pr interface{}) (bool, error)
	Host() HostDetail
	// ID returns the global, unique identifier of the repository in the host.
	ID() int64
	// IsArchived returns true if the repository has been archived on the host.
	IsArchived() bool
	IsPullRequestClosed(pr interface{}) bool
	IsPullRequestMerged(pr interface{}) bool
	IsPullRequestOpen(pr interface{}) bool
	ListPullRequestComments(pr interface{}) ([]PullRequestComment, error)
	MergePullRequest(deleteBranch bool, pr interface{}) error
	Name() string
	PullRequest(pr any) *PullRequest
	Owner() string
	UpdatePullRequest(data PullRequestData, pr interface{}) error
	WebUrl() string
	Raw() any
}

type Host interface {
	HostDetail
	CreateFromName(name string) (Repository, error)
	CreateFromJson(dec *json.Decoder) (Repository, error)
	ListRepositories(since *time.Time, result chan []Repository, errChan chan error)
	ListRepositoriesWithOpenPullRequests(result chan []Repository, errChan chan error)
}

type UserInfo struct {
	Email string
	Name  string
}

type HostDetail interface {
	AuthenticatedUser() (*UserInfo, error)
	Name() string
}

// RepositoryLister lists all repositories from the cache.
//
// An implementation queries all hosts to gather the list of repositories.
// Every repository is then send to the result channel.
// If an error occurs, then the error is sent to the errChan channel.
type RepositoryLister interface {
	List(hosts []Host, result chan Repository, errChan chan error)
}

// RepositoryCacheRemover removes a repository from the cache.
type RepositoryCacheRemover interface {
	Remove(repo Repository) error
}

func CreatePullRequestCommentWithIdentifier(body string, identifier string, pr interface{}, repo Repository) error {
	if identifier == "" {
		return errors.New("identifier is empty")
	}

	comments, err := repo.ListPullRequestComments(pr)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("<!-- saturn-bot::{%s} -->", identifier)
	for _, comment := range comments {
		if strings.HasPrefix(comment.Body, prefix) {
			return nil
		}
	}

	log.Log().Debugf("Create comment on pull request of repository %s with identifier %s", repo.FullName(), identifier)
	return repo.CreatePullRequestComment(prefix+"\n"+body, pr)
}

func DeletePullRequestCommentByIdentifier(identifier string, pr interface{}, repo Repository) error {
	if identifier == "" {
		return errors.New("identifier is empty")
	}

	comments, err := repo.ListPullRequestComments(pr)
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("<!-- saturn-bot::{%s} -->", identifier)
	for _, comment := range comments {
		if strings.HasPrefix(comment.Body, prefix) {
			log.Log().Debugf("Delete comment on pull request of repository %s with identifier %s", repo.FullName(), identifier)
			err := repo.DeletePullRequestComment(comment, pr)
			if err != nil {
				return fmt.Errorf("delete comment by identifier %s: %w", identifier, err)
			}

			return nil
		}
	}

	return nil
}

// NewRepositoryFromName create a new [Repository] by finding the [Host] that serves
// the repository.
//
// Returns an error if no [Host] can be identified.
func NewRepositoryFromName(hosts []Host, repositoryName string) (Repository, error) {
	for _, h := range hosts {
		repo, err := h.CreateFromName(repositoryName)
		if err != nil {
			return nil, err
		}

		if repo != nil {
			return repo, nil
		}
	}

	var hostNameList []string
	for _, h := range hosts {
		hostNameList = append(hostNameList, "'"+h.Name()+"'")
	}

	hostNames := strings.Join(hostNameList, ",")
	return nil, fmt.Errorf("no host found for repository '%s'\n  available hosts: %s\n  set githubAddress or gitlabAddress - https://saturn-bot.readthedocs.io/en/latest/configuration/", repositoryName, hostNames)
}
