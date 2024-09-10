package template

import (
	"bytes"
	"embed"
	"fmt"
	htmlTemplate "html/template"
)

var (
	//go:embed templates
	data      embed.FS
	templates *htmlTemplate.Template
)

func init() {
	templates = htmlTemplate.Must(htmlTemplate.ParseFS(data, "templates/*.tpl"))
}

type BranchModifiedInput struct {
	Checksums     []string
	DefaultBranch string
}

func RenderBranchModified(in BranchModifiedInput) (string, error) {
	buf := &bytes.Buffer{}
	err := templates.ExecuteTemplate(buf, "comment-branch-modified.tpl", in)
	if err != nil {
		return "", fmt.Errorf("render branch modified template: %w", err)
	}

	return buf.String(), nil
}

// Data is the root structure passed to templates.
type Data struct {
	Run        map[string]string
	Repository DataRepository
	TaskName   string
}

// DataRepository is the sub-resource in templates that exposes info about a repository.
type DataRepository struct {
	FullName string
	Host     string
	Name     string
	Owner    string
	WebUrl   string
}

type PullRequestDescriptionInput struct {
	AutoMergeText string
	Body          string
	IgnoreText    string
}

func RenderPullRequestDescription(in PullRequestDescriptionInput) (string, error) {
	buf := &bytes.Buffer{}
	err := templates.ExecuteTemplate(buf, "pull-request-description.tpl", in)
	if err != nil {
		return "", fmt.Errorf("render pull request description template: %w", err)
	}

	return buf.String(), nil
}
