package cache

import (
	"errors"
	"fmt"

	"github.com/wndhydrnt/saturn-bot/pkg/db"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("item not found")
)

func NewCacheDb(path string) (*gorm.DB, error) {
	gormDb, err := db.New(false, path, db.Migrate(migrations))
	if err != nil {
		return nil, fmt.Errorf("initialize Cache db: %w", err)
	}

	return gormDb, nil
}

type Cache struct {
	db *gorm.DB
}

func New(dbPath string) (*Cache, error) {
	gormDb, err := NewCacheDb(dbPath)
	if err != nil {
		return nil, err
	}

	return &Cache{db: gormDb}, nil
}

func (c *Cache) Delete(key string) error {
	result := c.db.Delete(&Item{Key: key})
	if result.Error != nil {
		return fmt.Errorf("delete cache item %s: %w", key, result.Error)
	}

	return nil
}

func (c *Cache) Get(key string) ([]byte, error) {
	item := &Item{}
	result := c.db.Where("key = ?", key).First(item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
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
