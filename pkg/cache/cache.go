package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/wndhydrnt/saturn-bot/pkg/task"
)

const (
	DefaultJsonFileName = "cache.json"
)

// CachedTask represents a single cache entry.
type CachedTask struct {
	Checksum        string
	LastExecutionAt int64
}

// CachedData is the root structure of the cache.
type CachedData map[string]CachedTask

// Reader defines the Read() method.
// Types implementing this interface should return all cached data or an error.
type Reader interface {
	Read() (CachedData, error)
}

// Writer defines the Write() method.
// Types implementing this interface should update the whole cache.
type Writer interface {
	Write(CachedData) error
}

// Cache stores the latest timestamp of the execution of each task.
// Based on methods exposed by Cache, other code can decide to
// query all repositories or to only query repositories that have changed
// since the previous execution.
type Cache struct {
	reader Reader
	writer Writer
}

// NewCache returns a new [Cache] for the given [Reader] and [Writer].
func NewCache(reader Reader, writer Writer) *Cache {
	return &Cache{reader: reader, writer: writer}
}

// GetOldestExecutionAt returns the oldest timestamp of a task in tasks.
// The timestamp is in microseconds.
func (c *Cache) GetOldestExecutionAt(tasks ...*task.Task) (int64, error) {
	cachedData, err := c.reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read cached data: %w", err)
	}

	var oldest int64
	for _, t := range tasks {
		cachedTask, known := cachedData[t.Name]
		if !known {
			// Task not in cache.
			// Return 0 to indicate that a full run is necessary.
			return 0, nil
		}

		if cachedTask.Checksum != t.Checksum() {
			// Task has changed on disk.
			// Return 0 to indicate that a full run is necessary.
			return 0, nil
		}

		if oldest == 0 {
			oldest = cachedTask.LastExecutionAt
		} else {
			if oldest > cachedTask.LastExecutionAt {
				oldest = cachedTask.LastExecutionAt
			}
		}
	}

	return oldest, nil
}

// Update updates all timestamps of the given tasks.
func (c *Cache) Update(at int64, tasks ...*task.Task) error {
	cachedData, err := c.reader.Read()
	if err != nil {
		return fmt.Errorf("read cached data for update: %w", err)
	}

	for _, t := range tasks {
		cachedData[t.Name] = CachedTask{
			Checksum:        t.Checksum(),
			LastExecutionAt: at,
		}
	}

	err = c.writer.Write(cachedData)
	if err != nil {
		return fmt.Errorf("update cached data: %w", err)
	}

	return nil
}

// File is a [Reader] and [Writer] backed by a file.
type File struct {
	Path string
}

// NewFileCache returns a new [Cache] that is backed by a file at path.
// The serialization format of the file is JSON.
func NewFileCache(path string) *Cache {
	storage := &File{Path: path}
	return NewCache(storage, storage)
}

// Read implements [Reader].
func (jf *File) Read() (CachedData, error) {
	b, err := os.ReadFile(jf.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CachedData{}, nil
		}

		return nil, fmt.Errorf("read json cache file %s: %w", jf.Path, err)
	}

	var cachedData CachedData
	err = json.Unmarshal(b, &cachedData)
	if err != nil {
		// If un-marshalling fails, assume that the file is broken or the format changed.
		// Return empty data.
		// File gets overwritten when Write() gets called.
		return CachedData{}, nil
	}

	return cachedData, nil
}

// Write implements [Writer].
func (jf *File) Write(data CachedData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal json cache file %s: %w", jf.Path, err)
	}

	err = os.WriteFile(jf.Path, b, 0600)
	if err != nil {
		return fmt.Errorf("write json cache file %s: %w", jf.Path, err)
	}

	return nil
}
