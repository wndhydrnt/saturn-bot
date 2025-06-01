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

// Cacher defines functions expected by a cache.
// [github.com/wndhydrnt/saturn-bot/pkg/cache.Cache] implements this interface.
type Cacher interface {
	Delete(key string) error
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
}

type PullRequestCache struct {
	cache Cacher
	clock clock.Clock
}

func NewPullRequestCache(c Cacher, clock clock.Clock) *PullRequestCache {
	return &PullRequestCache{cache: c, clock: clock}
}

func (c *PullRequestCache) Delete(branchName, repo string) {
	_ = c.cache.Delete(createKey(branchName, repo))
}

func (c *PullRequestCache) Get(branchName, repoName string) *PullRequest {
	key := createKey(branchName, repoName)
	data, err := c.cache.Get(key)
	if err != nil {
		return nil
	}

	pr := &PullRequest{}
	err = json.Unmarshal(data, pr)
	if err != nil {
		return nil
	}

	return pr
}

func (c *PullRequestCache) Set(branchName, repoName string, pr *PullRequest) {
	if pr == nil {
		return
	}

	data, err := json.Marshal(pr)
	if err != nil {
		return
	}

	_ = c.cache.Set(createKey(branchName, repoName), data)
}

func (c *PullRequestCache) Update(hosts []Host) error {
	var iterErr error
	for _, h := range hosts {
		var since *time.Time
		tsKey := fmt.Sprintf("%s_pr_ts", h.Name())
		tsRaw, err := c.cache.Get(tsKey)
		if err == nil && len(tsRaw) > 0 {
			ts, err := strconv.ParseInt(string(tsRaw), 10, 64)
			if err == nil {
				since = ptr.To(time.Unix(ts, 0))
			}
		}

		if since == nil {
			log.Log().Infof("Performing full update of PR cache for host %s", h.Name())
		} else {
			log.Log().Infof("Performing partial update of PR cache for host %s with changes since %s", h.Name(), since.Format(time.RFC3339))
		}

		iter := h.PullRequestIterator()
		start := c.clock.Now().UTC().Unix()
		updateCounter := 0
		for pr := range iter.ListPullRequests(since) {
			existingPr := c.Get(pr.BranchName, pr.RepositoryName)
			// Only update the cache with the latest version of the pull request.
			if existingPr != nil && existingPr.CreatedAt.After(pr.CreatedAt) {
				continue
			}

			c.Set(pr.BranchName, pr.RepositoryName, pr)
			updateCounter++
		}

		if iter.ListPullRequestsError() == nil {
			log.Log().Infof("Updated PR cache with %d new items from host %s", updateCounter, h.Name())
			tsValue := strconv.FormatInt(start, 10)
			_ = c.cache.Set(tsKey, []byte(tsValue))
		} else {
			iterErr = errors.Join(iterErr, err)
		}
	}

	return iterErr
}

func createKey(branchName string, repoName string) string {
	return repoName + "_" + branchName
}
