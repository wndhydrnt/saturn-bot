package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/wndhydrnt/saturn-bot/pkg/host"
)

const (
	locksDir = "locks"
	lockName = "repo.lock"
)

var (
	lockTryWait = 1 * time.Second
)

type locker struct {
	fl *flock.Flock
}

func (l *locker) lock(dataDir string, repo host.Repository) error {
	if dataDir == "" {
		return nil
	}

	dir := filepath.Join(dataDir, locksDir, repo.FullName())
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("create lock directory: %w", err)
	}

	l.fl = flock.New(filepath.Join(dir, lockName))
	for {
		locked, err := l.fl.TryLock()
		if err != nil {
			return err
		}

		if locked {
			return nil
		}

		time.Sleep(lockTryWait)
	}
}

func (l *locker) unlock() error {
	if l.fl == nil {
		return nil
	}

	return l.fl.Unlock()
}
