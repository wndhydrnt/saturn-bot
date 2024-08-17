package host

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/log"
	gsTemplate "github.com/wndhydrnt/saturn-bot/pkg/template"
)

var (
	ErrFileNotFound        = errors.New("file not found")
	ErrPullRequestNotFound = errors.New("pull request not found")

	tplPrBodyDefault  = template.Must(template.New("bodyDefault").Parse("Apply changes from task {{.TaskName}}."))
	tplPrTitleDefault = template.Must(template.New("titleDefault").Parse("saturn-bot: task {{.TaskName}}"))
)

// PullRequest holds data on an existing pull request.
type PullRequest struct {
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
	AutoMergeAfter *time.Duration
	Body           string
	Labels         []string
	MergeOnce      bool
	Reviewers      []string
	TaskName       string
	TemplateData   map[string]any
	Title          string
}

func (prd PullRequestData) GetBody() (string, error) {
	var bodyTpl *template.Template
	if prd.Body == "" {
		bodyTpl = tplPrBodyDefault
	} else {
		var err error
		bodyTpl, err = template.New("body").Parse(prd.Body)
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
		if prd.AutoMergeAfter == nil {
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

	content, err := gsTemplate.RenderPullRequestDescription(gsTemplate.PullRequestDescriptionInput{
		AutoMergeText: autoMergeText,
		Body:          buf.String(),
		IgnoreText:    ignoreText,
	})
	if err != nil {
		return "", err
	}

	return content, nil
}

func (prd PullRequestData) GetTitle() (string, error) {
	var tpl *template.Template
	if prd.Title == "" {
		tpl = tplPrTitleDefault
	} else {
		var err error
		tpl, err = template.New("title").Parse(prd.Title)
		if err != nil {
			return "", fmt.Errorf("parse title template of task %s: %w", prd.TaskName, err)
		}
	}

	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, prd.TemplateData)
	if err != nil {
		return "", fmt.Errorf("execute title template of task %s: %w", prd.TaskName, err)
	}

	return buf.String(), nil
}

type Repository interface {
	BaseBranch() string
	CanMergePullRequest(pr interface{}) (bool, error)
	CloneUrlHttp() string
	CloneUrlSsh() string
	ClosePullRequest(msg string, pr interface{}) error
	CreatePullRequestComment(body string, pr interface{}) error
	CreatePullRequest(branch string, data PullRequestData) error
	DeleteBranch(pr interface{}) error
	DeletePullRequestComment(comment PullRequestComment, pr interface{}) error
	FindPullRequest(branch string) (any, error)
	FullName() string
	GetFile(fileName string) (string, error)
	GetPullRequestBody(pr interface{}) string
	GetPullRequestCreationTime(pr interface{}) time.Time
	HasFile(path string) (bool, error)
	HasSuccessfulPullRequestBuild(pr interface{}) (bool, error)
	Host() string
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
}

type Host interface {
	CreateFromName(name string) (Repository, error)
	ListRepositories(since *time.Time, result chan []Repository, errChan chan error)
	ListRepositoriesWithOpenPullRequests(result chan []Repository, errChan chan error)
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
