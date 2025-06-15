package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

// PullRequestFactory is a function that returns a new instance of the struct that represents a pull request of a host.
// This function is used to properly unmarshal cached data.
type PullRequestFactory func() any

// Cacher defines functions expected by a cache.
// [github.com/wndhydrnt/saturn-bot/pkg/cache.Cache] implements this interface.
type Cacher interface {
	Delete(key string) error
	DeleteAllByTag(tagName string) error
	Get(key string) ([]byte, error)
	GetAllByTag(tag string) ([][]byte, error)
	Set(key string, value []byte) error
	SetWithTags(key string, value []byte, tags ...string) error
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
	// SetPullRequestFactory adds a factory function that returns the data struct for hostType to unmarshal
	// a pull request struct.
	SetPullRequestFactory(hostType Type, fac PullRequestFactory)
}

type typePeek struct {
	Type Type
}

type pullRequestCache struct {
	cache     Cacher
	factories map[Type]PullRequestFactory
}

func NewPullRequestCache(c Cacher, factories map[Type]PullRequestFactory) PullRequestCache {
	return &pullRequestCache{
		cache:     c,
		factories: factories,
	}
}

func NewPullRequestCacheFromHosts(c Cacher, hosts []Host) PullRequestCache {
	factories := make(map[Type]PullRequestFactory, len(hosts))
	for _, host := range hosts {
		factories[host.Type()] = host.PullRequestFactory()
	}

	return NewPullRequestCache(c, factories)
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

	fac, ok := c.factories[pt.Type]
	if !ok {
		return nil
	}

	pr := &PullRequest{
		Raw: fac(),
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

// SetPullRequestFactory implements [PullRequestCache].
func (c *pullRequestCache) SetPullRequestFactory(hostType Type, fac PullRequestFactory) {
	c.factories[hostType] = fac
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
