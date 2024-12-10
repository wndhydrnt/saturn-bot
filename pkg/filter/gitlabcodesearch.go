package filter

import (
	"context"
	"errors"
	"fmt"
	"slices"

	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
	"github.com/wndhydrnt/saturn-bot/pkg/params"
)

// GitlabCodeSearch creates jq filters.
type GitlabCodeSearchFactory struct{}

// Create implements Factory.
func (f GitlabCodeSearchFactory) Create(opts CreateOptions, params params.Params) (Filter, error) {
	query, err := params.String("query", "")
	if err != nil {
		return nil, err
	}

	if query == "" {
		return nil, fmt.Errorf("required parameter `query` not set")
	}

	gcs := &GitlabCodeSearch{}
	for _, h := range opts.Hosts {
		gl, ok := h.(*host.GitLabHost)
		if ok {
			gcs.Host = gl
			break
		}
	}

	return gcs, nil
}

// Name implements Factory.
func (f GitlabCodeSearchFactory) Name() string {
	return "gitlabCodeSearch"
}

type GitlabCodeSearch struct {
	Host    *host.GitLabHost
	GroupID any
	Query   string

	searchResults []int64
}

func (s *GitlabCodeSearch) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(sbcontext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context passed to filter search does not contain a repository")
	}

	return s.matchesQuery(repo.ID())
}

func (s *GitlabCodeSearch) String() string {
	return fmt.Sprintf("gitlabCodeSearch(query=%s)", s.Query)
}

func (s *GitlabCodeSearch) matchesQuery(id int64) (bool, error) {
	if s.searchResults == nil {
		var err error
		s.searchResults, err = s.Host.SearchCode(s.GroupID, s.Query)
		if err != nil {
			return false, fmt.Errorf("search gitlab: %w", err)
		}
	}

	return slices.Contains(s.searchResults, id), nil
}
