package filter

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	gsContext "github.com/wndhydrnt/saturn-sync/pkg/context"
	"github.com/wndhydrnt/saturn-sync/pkg/host"
)

type FilterRepository interface {
	GetFile(fileName string) (string, error)
	FullName() string
	HasFile(path string) (bool, error)
}

type Filter interface {
	Do(context.Context) (bool, error)
	Name() string
	String() string
}

type FileExists struct {
	fileName string
}

func NewFileExists(fileName string) *FileExists {
	return &FileExists{fileName: fileName}
}

func (fe *FileExists) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(gsContext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context passed to filter fileExists does not contain a repository")
	}

	return repo.HasFile(fe.fileName)
}

func (fe *FileExists) Name() string {
	return "hasFile"
}

func (fe *FileExists) String() string {
	return fmt.Sprintf("hasFile(fileName=%s)", fe.fileName)
}

type FileContainsLine struct {
	path   string
	search *regexp.Regexp
}

func NewFileContainsLine(path, search string) (*FileContainsLine, error) {
	re, err := regexp.Compile(search)
	if err != nil {
		return nil, fmt.Errorf("lineInFile: compile search %s to regexp: %w", search, err)
	}

	return &FileContainsLine{path: path, search: re}, nil
}

func (fcl *FileContainsLine) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(gsContext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context passed to filter lineInFile does not contain a repository")
	}

	content, err := repo.GetFile(fcl.path)
	if err != nil {
		if errors.Is(err, host.ErrFileNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("lineInFile: get file from repository: %w", err)
	}

	for _, line := range strings.Split(content, "\n") {
		match := fcl.search.MatchString(line)
		if match {
			return true, nil
		}
	}

	return false, nil
}

func (fcl *FileContainsLine) Name() string {
	return "fileContainsLine"
}

func (fcl *FileContainsLine) String() string {
	return fmt.Sprintf("lineInFile(path=%s, search=%s)", fcl.path, fcl.search.String())
}

type RepositoryNames struct {
	matchers []*regexp.Regexp
}

func NewRepositoryNames(names []string) (*RepositoryNames, error) {
	var nameRegexp []*regexp.Regexp
	for _, name := range names {
		if strings.HasPrefix(name, "https://") || strings.HasPrefix(name, "http://") {
			u, err := url.Parse(name)
			if err != nil {
				return nil, fmt.Errorf("name looks like a url but could ne be parsed %s: %w", name, err)
			}

			path := strings.TrimSuffix(u.Path, path.Ext(u.Path))
			name = fmt.Sprintf("%s%s", u.Host, path)
		}

		if !strings.HasPrefix(name, "^") {
			name = "^" + name
		}

		if !strings.HasSuffix(name, "$") {
			name = name + "$"
		}

		r, err := regexp.Compile(name)
		if err != nil {
			return nil, fmt.Errorf("compile regex of repository name: %w", err)
		}

		nameRegexp = append(nameRegexp, r)
	}

	return &RepositoryNames{matchers: nameRegexp}, nil
}

func (r *RepositoryNames) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(gsContext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context passed to filter repositoryName does not contain a repository")
	}

	for _, matcher := range r.matchers {
		match := matcher.MatchString(repo.FullName())
		if match {
			return true, nil
		}
	}

	return false, nil
}

func (r *RepositoryNames) Name() string {
	return "repositoryNames"
}

func (r *RepositoryNames) String() string {
	var matcherStrings []string
	for _, matcher := range r.matchers {
		matcherStrings = append(matcherStrings, matcher.String())
	}
	return fmt.Sprintf("repositoryNames(names=%s)", strings.Join(matcherStrings, ","))
}

// Reverse takes the result of a wrapped filter and returns the opposite.
type Reverse struct {
	wrapped Filter
}

// NewReverse returns a new Reverse filter.
func NewReverse(wrapped Filter) *Reverse {
	return &Reverse{wrapped: wrapped}
}

// Do implements Filter
func (r *Reverse) Do(ctx context.Context) (bool, error) {
	result, err := r.wrapped.Do(ctx)
	if err != nil {
		return false, err
	}

	return !result, nil
}

// Name implements Filter
func (r *Reverse) Name() string {
	return fmt.Sprintf("reverse(%s)", r.wrapped.Name())
}

// String implements Filter
func (r *Reverse) String() string {
	return fmt.Sprintf("reverse(%s)", r.wrapped.String())
}
