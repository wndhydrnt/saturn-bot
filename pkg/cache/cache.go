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

func newCacheDb(path string) (*gorm.DB, error) {
	gormDb, err := db.New(false, path, db.Migrate(migrations))
	if err != nil {
		return nil, fmt.Errorf("initialize Cache db: %w", err)
	}

	return gormDb, nil
}

type Cache struct {
	db *gorm.DB
}

// New returns a new [Cache] that stores its data in dbPath.
func New(dbPath string) (*Cache, error) {
	gormDb, err := newCacheDb(dbPath)
	if err != nil {
		return nil, err
	}

	return &Cache{db: gormDb}, nil
}

// Delete deletes the item identified by key in the cache.
func (c *Cache) Delete(key string) error {
	result := c.db.Delete(&item{Key: key})
	if result.Error != nil {
		return fmt.Errorf("delete cache item %s: %w", key, result.Error)
	}

	return nil
}

// Get returns the item identified by key from the cache.
// It returns [ErrNotFound] if the item doesn't exist in the cache.
func (c *Cache) Get(key string) ([]byte, error) {
	it := &item{}
	result := c.db.Where("key = ?", key).First(it)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get cache item %s: %w", key, result.Error)
	}

	return it.Value, nil
}

// Set writes the item identified by key with value to the cache.
func (c *Cache) Set(key string, value []byte) error {
	if value == nil {
		return nil
	}

	it := &item{
		Key:   key,
		Value: value,
	}
	result := c.db.Save(it)
	if result.Error != nil {
		return fmt.Errorf("set Cache item %s: %w", key, result.Error)
	}

	return nil
}
