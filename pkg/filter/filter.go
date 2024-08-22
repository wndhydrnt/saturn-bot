package filter

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	gsContext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/str"
)

var (
	BuiltInFactories = []Factory{
		FileContentFactory{},
		FileFactory{},
		RepositoryFactory{},
	}
)

type FilterRepository interface {
	GetFile(fileName string) (string, error)
	HasFile(path string) (bool, error)
	Host() host.HostDetail
	Name() string
	Owner() string
}

type Filter interface {
	Do(context.Context) (bool, error)
	String() string
}

type Factory interface {
	Create(params map[string]any) (Filter, error)
	Name() string
}

type FileFactory struct{}

func (f FileFactory) Name() string {
	return "file"
}

func (f FileFactory) Create(params map[string]any) (Filter, error) {
	if params["paths"] == nil {
		return nil, fmt.Errorf("required parameter `paths` not set")
	}

	op, ok := params["op"].(string)
	if !ok {
		op = opAnd
	} else {
		if op != opAnd && op != opOr {
			return nil, fmt.Errorf("value of parameter `op` can be and,or not '%s'", params["op"])
		}
	}

	inPaths, ok := params["paths"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("required parameter `paths` is not a list")
	}

	var paths []string
	for _, item := range inPaths {
		p, ok := item.(string)
		if ok {
			paths = append(paths, p)
		}
	}

	return &File{Op: op, Paths: paths}, nil
}

const (
	opAnd = "and"
	opOr  = "or"
)

type File struct {
	Op    string
	Paths []string
}

func (fe *File) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(gsContext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context passed to filter fileExists does not contain a repository")
	}

	for _, p := range fe.Paths {
		result, err := repo.HasFile(p)
		if err != nil {
			return false, nil
		}

		if fe.Op == opAnd && !result {
			return false, nil
		}

		if fe.Op == opOr && result {
			return true, nil
		}
	}

	if fe.Op == opAnd {
		return true, nil
	} else {
		return false, nil
	}
}

func (fe *File) String() string {
	return fmt.Sprintf("file(op=%s,paths=%s)", fe.Op, fe.Paths)
}

type FileContentFactory struct{}

func (f FileContentFactory) Name() string {
	return "fileContent"
}

func (f FileContentFactory) Create(params map[string]any) (Filter, error) {
	if params["path"] == nil {
		return nil, fmt.Errorf("required parameter `path` not set")
	}
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter `path` is of type %T not string", params["path"])
	}

	if params["regexp"] == nil {
		return nil, fmt.Errorf("required parameter `regexp` not set")
	}
	reRaw, ok := params["regexp"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter `regexp` is of type %T not string", params["regexp"])
	}

	re, err := regexp.Compile(reRaw)
	if err != nil {
		return nil, fmt.Errorf("compile parameter `regexp` to regular expression: %w", err)
	}

	return &FileContent{Path: path, Regexp: re}, nil
}

type FileContent struct {
	Path   string
	Regexp *regexp.Regexp
}

func (fcl *FileContent) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(gsContext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context passed to filter lineInFile does not contain a repository")
	}

	content, err := repo.GetFile(fcl.Path)
	if err != nil {
		if errors.Is(err, host.ErrFileNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("fileContent: get file from repository: %w", err)
	}

	for _, line := range strings.Split(content, "\n") {
		match := fcl.Regexp.MatchString(line)
		if match {
			return true, nil
		}
	}

	return false, nil
}

func (fcl *FileContent) String() string {
	return fmt.Sprintf("fileContent(path=%s,regexp=%s)", fcl.Path, fcl.Regexp.String())
}

type Repository struct {
	Host  *regexp.Regexp
	Owner *regexp.Regexp
	Repo  *regexp.Regexp
}

type RepositoryFactory struct{}

func (f RepositoryFactory) Name() string {
	return "repository"
}

func (f RepositoryFactory) Create(params map[string]any) (Filter, error) {
	if params["host"] == nil {
		return nil, fmt.Errorf("required parameter `host` not set")
	}
	host, ok := params["host"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `host` is of type %T not string", params["host"])
	}

	hostRe, err := regexp.Compile(str.EncloseRegex(host))
	if err != nil {
		return nil, fmt.Errorf("compile parameter `host` to regular expression: %w", err)
	}

	if params["owner"] == nil {
		return nil, fmt.Errorf("required parameter `owner` not set")
	}
	owner, ok := params["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter `owner` is of type %T not string", params["owner"])
	}

	ownerRe, err := regexp.Compile(str.EncloseRegex(owner))
	if err != nil {
		return nil, fmt.Errorf("compile parameter `owner` to regular expression: %w", err)
	}

	if params["name"] == nil {
		return nil, fmt.Errorf("required parameter `name` not set")
	}
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("required parameter `name` is of type %T not string", params["name"])
	}

	nameRe, err := regexp.Compile(str.EncloseRegex(name))
	if err != nil {
		return nil, fmt.Errorf("compile parameter `name` to regular expression: %w", err)
	}

	return &Repository{Host: hostRe, Owner: ownerRe, Repo: nameRe}, nil
}

func (r *Repository) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(gsContext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context passed to filter repositoryName does not contain a repository")
	}

	if !r.Host.MatchString(repo.Host().Name()) {
		return false, nil
	}

	if !r.Owner.MatchString(repo.Owner()) {
		return false, nil
	}

	if !r.Repo.MatchString(repo.Name()) {
		return false, nil
	}

	return true, nil
}

func (r *Repository) String() string {
	return fmt.Sprintf("repository(host=%s,owner=%s,name=%s)", r.Host.String(), r.Owner.String(), r.Repo.String())
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

// String implements Filter
func (r *Reverse) String() string {
	return fmt.Sprintf("!%s", r.wrapped.String())
}
