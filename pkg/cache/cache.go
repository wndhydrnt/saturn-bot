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

type Cache interface {
	GetLastExecutionAt() int64
	GetCachedTasks() []CachedTask
	Read() error
	SetLastExecutionAt(at int64)
	UpdateCachedTasks(tasks []task.Task)
	Write() error
}

type jsonFile struct {
	LastExecutionAt int64
	Tasks           []CachedTask

	file string
}

func NewJsonFile(file string) Cache {
	return &jsonFile{file: file}
}

func (c *jsonFile) GetLastExecutionAt() int64 {
	return c.LastExecutionAt
}

func (c *jsonFile) Read() error {
	b, err := os.ReadFile(c.file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("read cache file %s: %w", c.file, err)
	}

	err = json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("unmarshal cache file %s: %w", c.file, err)
	}

	return nil
}

func (c *jsonFile) SetLastExecutionAt(at int64) {
	c.LastExecutionAt = at
}

func (c *jsonFile) GetCachedTasks() []CachedTask {
	return c.Tasks
}

func (c *jsonFile) UpdateCachedTasks(tasks []task.Task) {
	var cachedTasks []CachedTask
	for _, t := range tasks {
		cachedTasks = append(cachedTasks, CachedTask{Checksum: t.Checksum(), Name: t.SourceTask().Name})
	}

	c.Tasks = cachedTasks
}

func (c *jsonFile) Write() error {
	b, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal cache file %s: %w", c.file, err)
	}

	err = os.WriteFile(c.file, b, 0600)
	if err != nil {
		return fmt.Errorf("write cache file %s: %w", c.file, err)
	}

	return nil
}

type CachedTask struct {
	Checksum string
	Name     string
}
