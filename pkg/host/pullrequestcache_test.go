package host

import (
	"iter"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
)

type hostMock struct {
	Host

	iter PullRequestIterator
}

func (h *hostMock) PullRequestIterator() PullRequestIterator {
	return h.iter
}

type iterMock struct {
	err          error
	pullRequests []*PullRequest
}

func (i *iterMock) ListPullRequests(since *time.Time) iter.Seq[*PullRequest] {
	return nil
}

func (i *iterMock) ListPullRequestsError() error {
	return i.err
}

func TestUpdatePullRequestCache_FullUpdate(t *testing.T) {
	iterm := &iterMock{}
	hostm := &hostMock{iter: iterm}
	options.ToOptions()
	cacher, err := cache.New()
	require.NoError(t, err, "creates the cache db")
	prCache := NewPullRequestCache(cacher)
	err := UpdatePullRequestCache(clock.NewFakeDefault(), []Host{hostm}, prCache)

	require.NoError(t, err)
}

func TestUpdatePullRequestCache_PartialUpdate(t *testing.T) {

}

func TestUpdatePullRequestCache_UpdatesTheLatest(t *testing.T) {

}

func TestUpdatePullRequestCache_Error(t *testing.T) {

}
