package cache

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wndhydrnt/saturn-bot/pkg/db"
	"github.com/wndhydrnt/saturn-bot/pkg/options"
	"gorm.io/gorm"
)

const (
	dbName = "cache.db"
)

func NewCacheDb(opts options.Opts) (*gorm.DB, error) {
	dbPath := filepath.Join(opts.DataDir(), dbName)
	gormDb, err := db.New(false, dbPath, db.Migrate(migrations))
	if err != nil {
		return nil, fmt.Errorf("initialize Cache db: %w", err)
	}

	return gormDb, nil
}

type Cache struct {
	db *gorm.DB
}

func New(opts options.Opts) (*Cache, error) {
	gormDb, err := NewCacheDb(opts)
	if err != nil {
		return nil, err
	}

	return &Cache{db: gormDb}, nil
}

func (c *Cache) Delete(key string) error {
	result := c.db.Delete("key = ?", key)
	if result.Error != nil {
		return fmt.Errorf("delete cache item %s: %w", key, result.Error)
	}

	return nil
}

func (c *Cache) Get(key string) ([]byte, error) {
	item := &Item{}
	result := c.db.Where("key = ?", key).Find(item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return []byte{}, nil
		}

		return nil, fmt.Errorf("get cache item %s: %w", key, result.Error)
	}

	return item.Value, nil
}

func (c *Cache) Set(key string, value []byte) error {
	if value == nil {
		return nil
	}

	item := &Item{
		Key:   key,
		Value: value,
	}
	result := c.db.Save(item)
	if result.Error != nil {
		return fmt.Errorf("set Cache item %s: %w", key, result.Error)
	}

	return nil
}
