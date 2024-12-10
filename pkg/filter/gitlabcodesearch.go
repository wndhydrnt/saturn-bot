package filter

import (
	"context"
	"errors"
	"fmt"

	sbcontext "github.com/wndhydrnt/saturn-bot/pkg/context"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
)

type GitlabCodeSearch struct {
	GitlabGroupID int
	Query         string
	GitlabHost    *host.GitLabHost

	searchResults []string
}

func (s *GitlabCodeSearch) Do(ctx context.Context) (bool, error) {
	repo, ok := ctx.Value(sbcontext.RepositoryKey{}).(FilterRepository)
	if !ok {
		return false, errors.New("context passed to filter search does not contain a repository")
	}

	return s.matchesQuery(repo.FullName())
}

func (s *GitlabCodeSearch) String() string {
	return fmt.Sprintf("gitlabCodeSearch(query=%s)", s.Query)
}

func (s *GitlabCodeSearch) matchesQuery(name string) (bool, error) {
	if s.searchResults == nil {
		var err error
		s.searchResults, err = s.Searcher.SearchCode(s.Query, host.CodeSearchArgs{GitLabGroupID: s.GitlabGroupID})
		if err != nil {
			return false, fmt.Errorf("failed to send query: %w", err)
		}
	}

	for _, sr := range s.searchResults {
		if sr == repoID {
			return true, nil
		}
	}

	return false, nil
}
