package host

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/cache"
	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

const (
	cacheKeyRepositoryTpl         = "repo_%s"
	cacheKeyRepositoryMetadataTpl = "repo_meta_%s"
	cacheTagRepositoryTpl         = "repo_host_%s"
)

type fileCacheMetadata struct {
	NextFullUpdateAfter int64
	LastUpdateTs        int64
}

// RepositoryCache reads all repositories from hosts and stores them in a file cache.
// It helps reduce API requests to the host.
type RepositoryCache struct {
	Cacher Cacher
	Clock  clock.Clock
	Ttl    time.Duration

	mu sync.Mutex
}

// NewRepositoryCache returns a new [RepositoryCache].
func NewRepositoryCache(c Cacher, clock clock.Clock, dir string, ttl time.Duration) *RepositoryCache {
	// TODO(v1): Remove argument dir.
	fi, err := os.Stat(dir)
	if err == nil && fi.IsDir() {
		removeErr := os.RemoveAll(dir)
		if removeErr == nil {
			log.Log().Infof("Removed old cache direcotry %s", dir)
		} else {
			log.Log().Warnf("Failed to remove old cache directory %s: %v", dir, removeErr)
		}
	}

	return &RepositoryCache{
		Cacher: c,
		Clock:  clock,
		Ttl:    ttl,
	}
}

// List implements [RepositoryLister].
func (rc *RepositoryCache) List(hosts []Host, result chan Repository, errChan chan error) {
	err := rc.updateCache(hosts)
	if err != nil {
		errChan <- err
		return
	}

	for _, h := range hosts {
		rc.readRepositoriesForHost(h, result, errChan)
	}

	errChan <- nil
}

func (rc *RepositoryCache) readLastUpdateTimestamp(h Host) (*time.Time, error) {
	meta, err := rc.readMetadata(h)
	if err != nil {
		return nil, err
	}

	if meta.NextFullUpdateAfter < rc.Clock.Now().Unix() {
		return nil, nil
	}

	if meta.LastUpdateTs == 0 {
		// No update ever.
		// Return nil to trigger a full update.
		return nil, nil
	}

	// Return last update timestamp to trigger partial update.
	return ptr.To(time.Unix(meta.LastUpdateTs, 0)), nil
}

func (rc *RepositoryCache) writeLastUpdateTimestamp(h Host, t time.Time) error {
	meta, err := rc.readMetadata(h)
	if err != nil {
		return fmt.Errorf("read metadata before update: %w", err)
	}

	meta.LastUpdateTs = t.Unix()
	if meta.NextFullUpdateAfter < rc.Clock.Now().Unix() {
		meta.NextFullUpdateAfter = rc.Clock.Now().Add(rc.Ttl).Unix()
	}

	return rc.writeMetadata(h, meta)
}

func (rc *RepositoryCache) readMetadata(h Host) (fileCacheMetadata, error) {
	var meta fileCacheMetadata
	key := getRepositoryMetadataCacheKey(h)
	data, err := rc.Cacher.Get(key)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return meta, nil
		}

		return meta, fmt.Errorf("read repository metadata from cache for host %s: %w", h.Name(), err)
	}

	err = json.NewDecoder(bytes.NewReader(data)).Decode(&meta)
	if err != nil {
		return meta, fmt.Errorf("decode metadata from JSON: %w", err)
	}

	return meta, nil
}

func (rc *RepositoryCache) writeMetadata(h Host, meta fileCacheMetadata) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(meta); err != nil {
		return fmt.Errorf("encode cache metadata to JSON for host %s: %w", h.Name(), err)
	}

	key := getRepositoryMetadataCacheKey(h)
	if err := rc.Cacher.Set(key, buf.Bytes()); err != nil {
		return fmt.Errorf("write cache metadata to JSON for host %s: %w", h.Name(), err)
	}

	return nil
}

func (rc *RepositoryCache) readRepositoriesForHost(h Host, result chan Repository, errChan chan error) {
	tag := getRepositoryCacheTag(h)
	values, err := rc.Cacher.GetAllByTag(tag)
	if err != nil {
		errChan <- fmt.Errorf("get repositories by tag %s: %w", tag, err)
		return
	}

	for _, value := range values {
		dec := json.NewDecoder(bytes.NewReader(value))
		repo, err := h.CreateFromJson(dec)
		if err != nil {
			errChan <- err
			return
		}

		result <- repo
	}
}

func (rc *RepositoryCache) writeRepository(repo Repository) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(repo.Raw())
	if err != nil {
		return fmt.Errorf("encode repository to JSON for caching: %w", err)
	}

	key := getRepositoryCacheKey(repo)
	tag := getRepositoryCacheTag(repo.Host())
	if err := rc.Cacher.SetWithTags(key, buf.Bytes(), tag); err != nil {
		return fmt.Errorf("write repository to cache: %w", err)
	}

	return nil
}

func (rc *RepositoryCache) updateCache(hosts []Host) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	var wg sync.WaitGroup
	errChan := make(chan error, len(hosts))
	for _, h := range hosts {
		wg.Add(1)
		go func() {
			errChan <- rc.updateCacheForHost(h)
			wg.Done()
		}()
	}

	wg.Wait()
	close(errChan)
	var err error
	for updateErr := range errChan {
		err = errors.Join(err, updateErr)
	}

	return err
}

func (rc *RepositoryCache) updateCacheForHost(host Host) error {
	since, err := rc.readLastUpdateTimestamp(host)
	if err != nil {
		return err
	}

	if since == nil {
		log.Log().Infof("Full update of repository cache for host %s", host.Name())
		tag := getRepositoryCacheTag(host)
		if err := rc.Cacher.DeleteAllByTag(tag); err != nil {
			return fmt.Errorf("delete all repositories for host %s from cache: %w", host.Name(), err)
		}
	} else {
		log.Log().Infof("Partial update of repository cache for host %s since %s", host.Name(), since)
	}

	start := rc.Clock.Now()
	repoIterator := host.RepositoryIterator()
	updateCounter := 0
	for repo := range repoIterator.ListRepositories(since) {
		if err := rc.writeRepository(repo); err != nil {
			return err
		}

		updateCounter++
	}

	if err := repoIterator.Error(); err != nil {
		return err
	}

	log.Log().Infof("Updated repository cache with %d new items for host %s", updateCounter, host.Name())
	return rc.writeLastUpdateTimestamp(host, start)
}

func getRepositoryCacheKey(repo Repository) string {
	return fmt.Sprintf(cacheKeyRepositoryTpl, repo.FullName())
}

func getRepositoryMetadataCacheKey(host Host) string {
	return fmt.Sprintf(cacheKeyRepositoryMetadataTpl, host.Name())
}

func getRepositoryCacheTag(host HostDetail) string {
	return fmt.Sprintf(cacheTagRepositoryTpl, host.Name())
}
