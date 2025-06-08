package host

import (
	"errors"
	"iter"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wndhydrnt/saturn-bot/pkg/cache"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
)

type hostMock struct {
	Host

	iter PullRequestIterator
}

func (h *hostMock) Name() string {
	return "mock"
}

func (h *hostMock) PullRequestIterator() PullRequestIterator {
	return h.iter
}

type iterMock struct {
	err          error
	pullRequests []*PullRequest
	since        *time.Time
}

func (i *iterMock) ListPullRequests(since *time.Time) iter.Seq[*PullRequest] {
	i.since = since
	return func(yield func(*PullRequest) bool) {
		for _, pr := range i.pullRequests {
			if !yield(pr) {
				return
			}
		}
	}
}

func (i *iterMock) ListPullRequestsError() error {
	return i.err
}

func TestUpdatePullRequestCache_FullUpdate(t *testing.T) {
	iterm := &iterMock{
		pullRequests: []*PullRequest{
			{BranchName: "saturn-bot--unittest", RepositoryName: "unittest", Type: "mock"},
		},
	}
	hostm := &hostMock{iter: iterm}
	cacher, err := cache.New(filepath.Join(t.TempDir(), "cache.db"))
	require.NoError(t, err, "creates the cache db")
	prCache := NewPullRequestCache(cacher)
	prCache.SetRawFactory("mock", func() any { return map[string]interface{}{} })

	err = UpdatePullRequestCache(clock.NewFakeDefault(), []Host{hostm}, prCache)
	require.NoError(t, err, "call succeeds")

	require.Nil(t, iterm.since, "does a full update")
	require.Equal(t, iterm.pullRequests[0], prCache.Get("saturn-bot--unittest", "unittest"), "stores the pull request in the cache")
	ts, err := cacher.Get("mock_pr_ts")
	require.NoError(t, err, "reads pr timestamp from cache")
	require.Equal(t, "946684800", string(ts), "updates the pr timestamp for the host")
}

func TestUpdatePullRequestCache_PartialUpdate(t *testing.T) {
	iterm := &iterMock{
		pullRequests: []*PullRequest{
			{BranchName: "saturn-bot--unittest", RepositoryName: "unittest", Type: "mock"},
		},
	}
	hostm := &hostMock{iter: iterm}
	cacher, err := cache.New(filepath.Join(t.TempDir(), "cache.db"))
	require.NoError(t, err, "creates the cache db")
	prCache := NewPullRequestCache(cacher)
	lastUpdatedAt := time.Date(1999, 12, 31, 23, 59, 59, 0, time.UTC)
	prCache.SetLastUpdatedAtFor(hostm, lastUpdatedAt)
	prCache.SetRawFactory("mock", func() any { return map[string]interface{}{} })

	err = UpdatePullRequestCache(clock.NewFakeDefault(), []Host{hostm}, prCache)
	require.NoError(t, err, "call succeeds")

	require.True(t, iterm.since.Equal(lastUpdatedAt), "passes the time of the last update to the iterator")
	require.Equal(t, iterm.pullRequests[0], prCache.Get("saturn-bot--unittest", "unittest"))
	ts, err := cacher.Get("mock_pr_ts")
	require.NoError(t, err, "reads pr timestamp from cache")
	require.Equal(t, "946684800", string(ts), "updates the pr timestamp for the host")
}

func TestUpdatePullRequestCache_UpdatesTheLatest(t *testing.T) {
	iterm := &iterMock{
		pullRequests: []*PullRequest{
			{BranchName: "saturn-bot--unittest", RepositoryName: "unittest", Type: "mock", CreatedAt: time.Date(2010, 1, 1, 1, 0, 0, 0, time.UTC)},
			{BranchName: "saturn-bot--unittest", RepositoryName: "unittest", Type: "mock", CreatedAt: time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	hostm := &hostMock{iter: iterm}
	cacher, err := cache.New(filepath.Join(t.TempDir(), "cache.db"))
	require.NoError(t, err, "creates the cache db")
	prCache := NewPullRequestCache(cacher)
	prCache.SetRawFactory("mock", func() any { return map[string]interface{}{} })

	err = UpdatePullRequestCache(clock.NewFakeDefault(), []Host{hostm}, prCache)
	require.NoError(t, err, "call succeeds")

	require.Nil(t, iterm.since, "does a full update")
	require.Equal(t, iterm.pullRequests[0], prCache.Get("saturn-bot--unittest", "unittest"), "stores the pull request in the cache")
	ts, err := cacher.Get("mock_pr_ts")
	require.NoError(t, err, "reads pr timestamp from cache")
	require.Equal(t, "946684800", string(ts), "updates the pr timestamp for the host")
}

func TestUpdatePullRequestCache_Error(t *testing.T) {
	iterm := &iterMock{
		err: errors.New("some error"),
	}
	hostm := &hostMock{iter: iterm}
	cacher, err := cache.New(filepath.Join(t.TempDir(), "cache.db"))
	require.NoError(t, err, "creates the cache db")
	prCache := NewPullRequestCache(cacher)

	err = UpdatePullRequestCache(clock.NewFakeDefault(), []Host{hostm}, prCache)
	require.Error(t, err, "returns the error")

	_, err = cacher.Get("mock_pr_ts")
	require.ErrorIs(t, err, cache.ErrNotFound, "does not store the timestamp in the cache")
}
