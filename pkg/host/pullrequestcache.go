package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type PrCacheRawFactory func() any

// Cacher defines functions expected by a cache.
// [github.com/wndhydrnt/saturn-bot/pkg/cache.Cache] implements this interface.
type Cacher interface {
	Delete(key string) error
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
}

// A PullRequestCache caches pull request data.
type PullRequestCache interface {
	// Delete deletes the data in the cache, identified by branchName and repo.
	Delete(branchName, repo string)
	// Get reads the data from the cache, identified by branchName and repo.
	// It returns nil if the cache doesn't contain data.
	Get(branchName, repoName string) *PullRequest
	// Set writes pr to the cache, identified by branchName and repoName.
	Set(branchName, repoName string, pr *PullRequest)
	// LastUpdatedAtFor returns the last time at which the cache was updated for host.
	LastUpdatedAtFor(host Host) *time.Time
	// SetLastUpdatedAtFor sets the last time at which the cache was updated for host.
	SetLastUpdatedAtFor(host Host, updatedAt time.Time)
	// SetRawFactory adds a factory function that returns the data struct for hostType to unmarshal
	// a pull request struct.
	SetRawFactory(hostType Type, fac PrCacheRawFactory)
}

type typePeek struct {
	Type Type
}

type pullRequestCache struct {
	cache        Cacher
	rawFactories map[Type]PrCacheRawFactory
}

func NewPullRequestCache(c Cacher) PullRequestCache {
	rawFactories := map[Type]PrCacheRawFactory{
		GitHubType: func() any { return &github.PullRequest{} },
		GitLabType: func() any { return &gitlab.MergeRequest{} },
	}

	return &pullRequestCache{
		cache:        c,
		rawFactories: rawFactories,
	}
}

// Delete implements [PullRequestCache].
func (c *pullRequestCache) Delete(branchName, repo string) {
	_ = c.cache.Delete(createKey(branchName, repo))
}

// Get implements [PullRequestCache].
func (c *pullRequestCache) Get(branchName, repoName string) *PullRequest {
	key := createKey(branchName, repoName)
	data, err := c.cache.Get(key)
	if err != nil {
		return nil
	}

	pt := &typePeek{}
	err = json.Unmarshal(data, pt)
	if err != nil {
		return nil
	}

	rawStruct := c.rawFactories[pt.Type]()
	pr := &PullRequest{
		Raw: rawStruct,
	}
	err = json.Unmarshal(data, pr)
	if err != nil {
		return nil
	}

	return pr
}

// Set implements [PullRequestCache].
func (c *pullRequestCache) Set(branchName, repoName string, pr *PullRequest) {
	if pr == nil {
		return
	}

	data, err := json.Marshal(pr)
	if err != nil {
		return
	}

	_ = c.cache.Set(createKey(branchName, repoName), data)
}

// LastUpdatedAtFor implements [PullRequestCache].
func (c *pullRequestCache) LastUpdatedAtFor(host Host) *time.Time {
	var since *time.Time
	tsKey := fmt.Sprintf("%s_pr_ts", host.Name())
	tsRaw, err := c.cache.Get(tsKey)
	if err == nil && len(tsRaw) > 0 {
		ts, err := strconv.ParseInt(string(tsRaw), 10, 64)
		if err == nil {
			since = ptr.To(time.Unix(ts, 0))
		}
	}

	return since
}

// SetLastUpdatedAtFor implements [PullRequestCache].
func (c *pullRequestCache) SetLastUpdatedAtFor(host Host, updatedAt time.Time) {
	tsKey := fmt.Sprintf("%s_pr_ts", host.Name())
	tsValue := strconv.FormatInt(updatedAt.Unix(), 10)
	_ = c.cache.Set(tsKey, []byte(tsValue))
}

// SetRawFactory implements [PullRequestCache].
func (c *pullRequestCache) SetRawFactory(hostType Type, fac PrCacheRawFactory) {
	c.rawFactories[hostType] = fac
}

// UpdatePullRequestCache iterates over all hosts and updates prCache.
// It performs a full update if no previous update ran for a host
// and performs a partial update otherwise.
func UpdatePullRequestCache(clock clock.Clock, hosts []Host, prCache PullRequestCache) error {
	if prCache == nil {
		return nil
	}

	var iterErr error
	for _, h := range hosts {
		since := prCache.LastUpdatedAtFor(h)
		if since == nil {
			log.Log().Infof("Performing full update of PR cache for host %s", h.Name())
		} else {
			log.Log().Infof("Performing partial update of PR cache for host %s with changes since %s", h.Name(), since.Format(time.RFC3339))
		}

		iter := h.PullRequestIterator()
		start := clock.Now().UTC()
		updateCounter := 0
		for pr := range iter.ListPullRequests(since) {
			existingPr := prCache.Get(pr.BranchName, pr.RepositoryName)
			// Only update the cache with the latest version of the pull request.
			if existingPr != nil && existingPr.CreatedAt.After(pr.CreatedAt) {
				continue
			}

			prCache.Set(pr.BranchName, pr.RepositoryName, pr)
			updateCounter++
		}

		if iter.Error() == nil {
			log.Log().Infof("Updated PR cache with %d new items from host %s", updateCounter, h.Name())
			prCache.SetLastUpdatedAtFor(h, start)
		} else {
			iterErr = errors.Join(iterErr, iter.Error())
		}
	}

	return iterErr
}

func createKey(branchName string, repoName string) string {
	return repoName + "_" + branchName
}
