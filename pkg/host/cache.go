package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wndhydrnt/saturn-bot/pkg/clock"
	"github.com/wndhydrnt/saturn-bot/pkg/log"
	"github.com/wndhydrnt/saturn-bot/pkg/ptr"
)

type RepositoryFileCache struct {
	Clock clock.Clock
	Dir   string
}

func (rc *RepositoryFileCache) List(hosts []Host, result chan Repository, errChan chan error) {
	listRepos := make(chan []Repository)
	listReposErrs := make(chan error)
	start := rc.Clock.Now()
	for _, h := range hosts {
		since, err := rc.readLastUpdateTimestamp(h)
		if err != nil {
			errChan <- err
			return
		}

		go h.ListRepositories(since, listRepos, listReposErrs)
	}

	err := rc.receiveRepositories(len(hosts), listRepos, listReposErrs)
	if err != nil {
		errChan <- err
	}

	for _, h := range hosts {
		rc.readRepositoriesForHost(h, result, errChan)
		err := rc.writeLastUpdateTimestamp(h, start)
		if err != nil {
			errChan <- err
		}
	}

	errChan <- nil
}

func (rc *RepositoryFileCache) Remove(repo Repository) error {
	d := filepath.Join(rc.Dir, repo.FullName())
	return os.RemoveAll(d)
}

func (rc *RepositoryFileCache) readLastUpdateTimestamp(h Host) (*time.Time, error) {
	b, err := os.ReadFile(filepath.Join(rc.Dir, h.Name(), "timestamp"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, fmt.Errorf("read repository cache timestamp: %w", err)
	}

	ts, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse raw timestamp to int: %w", err)
	}

	return ptr.To(time.Unix(ts, 0)), nil
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

func (rc *RepositoryFileCache) receiveRepositories(expectedFinishes int, results chan []Repository, errChan chan error) error {
	finishes := 0
	for {
		select {
		case repoList := <-results:
			for _, repo := range repoList {
				if repo.IsArchived() {
					log.Log().Debugf("Removing archived repository from file cache %s", repo.FullName())
					_ = rc.Remove(repo)
					continue
				}

				err := rc.writeRepository(repo)
				if err != nil {
					finishes += 1
					return err
				}
			}

		case err := <-errChan:
			finishes += 1
			if err != nil {
				return err
			}
		}

		if expectedFinishes == finishes {
			return nil
		}
	}
}

func (rc *RepositoryFileCache) writeLastUpdateTimestamp(h Host, t time.Time) error {
	dir := filepath.Join(rc.Dir, h.Name())
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("create directory for timestamp: %w", err)
	}

	ts := strconv.FormatInt(t.Unix(), 10)
	err = os.WriteFile(filepath.Join(dir, "timestamp"), []byte(ts), 0600)
	if err != nil {
		return fmt.Errorf("write repository cache timestamp: %w", err)
	}

	return nil
}
