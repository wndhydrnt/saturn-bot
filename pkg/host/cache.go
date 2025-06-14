package host

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

const (
	metadataFileName = "metadata.json"
)

type fileCacheMetadata struct {
	NextFullUpdateAfter int64
	LastUpdateTs        int64
}

// RepositoryFileCache reads all repositories from hosts and stores them in a file cache.
// It helps reduce API requests to the host.
type RepositoryFileCache struct {
	Clock clock.Clock
	Dir   string
	Ttl   time.Duration

	mu sync.Mutex
}

// List implements [RepositoryLister].
func (rc *RepositoryFileCache) List(hosts []Host, result chan Repository, errChan chan error) {
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

func (rc *RepositoryFileCache) remove(repo Repository) error {
	d := filepath.Join(rc.Dir, repo.FullName())
	_, err := os.Stat(d)
	if err == nil {
		log.Log().Debugf("Removing archived repository from file cache %s", repo.FullName())
		return os.RemoveAll(d)
	}

	return nil
}

func (rc *RepositoryFileCache) readLastUpdateTimestamp(h Host) (*time.Time, error) {
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

func (rc *RepositoryFileCache) writeLastUpdateTimestamp(h Host, t time.Time) error {
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

func (rc *RepositoryFileCache) readMetadata(h Host) (fileCacheMetadata, error) {
	var meta fileCacheMetadata
	b, err := os.ReadFile(filepath.Join(rc.Dir, h.Name(), metadataFileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return meta, nil
		}

		return meta, fmt.Errorf("read repository cache metadata: %w", err)
	}

	err = json.NewDecoder(bytes.NewReader(b)).Decode(&meta)
	if err != nil {
		return meta, fmt.Errorf("decode metadata from JSON: %w", err)
	}

	return meta, nil
}

func (rc *RepositoryFileCache) writeMetadata(h Host, meta fileCacheMetadata) error {
	dir := filepath.Join(rc.Dir, h.Name())
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("create directory for metadata: %w", err)
	}

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(meta)
	if err != nil {
		return fmt.Errorf("encode cache metadata to JSON: %w", err)
	}

	err = os.WriteFile(filepath.Join(dir, metadataFileName), buf.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("write cache metadata to JSON: %w", err)
	}

	return nil
}

func (rc *RepositoryFileCache) readRepositoriesForHost(h Host, result chan Repository, errChan chan error) {
	err := filepath.WalkDir(filepath.Join(rc.Dir, h.Name()), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() != "repository.json" {
			return nil
		}

		repo, err := rc.readRepository(h, path)
		if err != nil {
			return err
		}

		result <- repo
		return nil
	})

	if err != nil {
		errChan <- err
	}
}

func (rc *RepositoryFileCache) readRepository(h Host, path string) (Repository, error) {
	cacheFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read repository cache file of %s: %w", h.Name(), err)
	}

	defer cacheFile.Close()
	dec := json.NewDecoder(cacheFile)
	repo, err := h.CreateFromJson(dec)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (rc *RepositoryFileCache) writeRepository(repo Repository) error {
	d := filepath.Join(rc.Dir, repo.FullName())
	err := os.MkdirAll(d, 0755)
	if err != nil {
		return fmt.Errorf("create cache directory of repository: %w", err)
	}

	f, err := os.Create(filepath.Join(d, "repository.json"))
	if err != nil {
		return fmt.Errorf("open repository cache file: %w", err)
	}

	defer f.Close()
	err = json.NewEncoder(f).Encode(repo.Raw())
	if err != nil {
		return fmt.Errorf("encode repository to JSON for caching: %w", err)
	}

	return nil
}

func (rc *RepositoryFileCache) updateCache(hosts []Host) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	wg := &sync.WaitGroup{}
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

func (rc *RepositoryFileCache) updateCacheForHost(host Host) error {
	since, err := rc.readLastUpdateTimestamp(host)
	if err != nil {
		return err
	}

	if since == nil {
		log.Log().Infof("Full update of repository cache for host %s", host.Name())
		_ = os.RemoveAll(rc.Dir)
	} else {
		log.Log().Infof("Partial update of repository cache for host %s since %s", host.Name(), since)
	}

	start := rc.Clock.Now()
	repoIterator := host.RepositoryIterator()
	updateCounter := 0
	for repo := range repoIterator.ListRepositories(since) {
		if repo.IsArchived() {
			_ = rc.remove(repo)
			continue
		}

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
