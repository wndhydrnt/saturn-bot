package cache

import (
	"bytes"
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
	result := c.db.Where("key = ?", key).Delete(&item{})
	if result.Error != nil {
		return fmt.Errorf("delete cache item %s: %w", key, result.Error)
	}

	return nil
}

// DeleteAllByTag deletes all items in the cache that have been tagged by tagName.
// Deletes an item even if multiple tags refer to it.
func (c *Cache) DeleteAllByTag(tagName string) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		var itemIDs []uint
		// Get all IDs of items first because sqlite doesn't support joins in DELETE statements.
		selectItemIDsResult := tx.
			Table("items").
			Select("items.id").
			Joins("INNER JOIN tags ON tags.item_id = items.id").
			Where("tags.name = ?", tagName).
			Find(&itemIDs)
		if selectItemIDsResult.Error != nil {
			return fmt.Errorf("select items to delete by tag %s: %w", tagName, selectItemIDsResult.Error)
		}

		if len(itemIDs) == 0 {
			return nil
		}

		deleteTagsResult := tx.
			Where("item_id IN ?", itemIDs).
			Delete(&tag{})
		if deleteTagsResult.Error != nil {
			return fmt.Errorf("delete tag entries for %s: %w", tagName, deleteTagsResult.Error)
		}

		deleteItemsResult := tx.
			Where("id IN ?", itemIDs).
			Delete(&item{})
		if deleteItemsResult.Error != nil {
			return fmt.Errorf("delete items for tag %s: %w", tagName, deleteItemsResult.Error)
		}

		return nil
	})
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

// GetAllByTag returns the values of all items tagged with tag.
func (c *Cache) GetAllByTag(tag string) ([][]byte, error) {
	var values [][]byte
	result := c.db.
		Table("items").
		Select("items.value").
		Joins("INNER JOIN tags ON tags.item_id = items.id").
		Where("tags.name = ?", tag).
		Find(&values)
	if result.Error != nil {
		return nil, fmt.Errorf("get all by tag %s: %w", tag, result.Error)
	}

	return values, nil
}

// Set writes the item identified by key with value to the cache.
func (c *Cache) Set(key string, value []byte) error {
	if value == nil {
		return nil
	}

	var it item
	resultFirstOrCreate := c.db.Where("key = ?", key).Attrs(item{Key: key, Value: value}).FirstOrCreate(&it)
	if resultFirstOrCreate.Error != nil {
		return fmt.Errorf("create or get cache item: %w", resultFirstOrCreate.Error)
	}

	if bytes.Equal(it.Value, value) {
		return nil
	}

	it.Value = value
	result := c.db.Save(it)
	if result.Error != nil {
		return fmt.Errorf("set Cache item %s: %w", key, result.Error)
	}

	return nil
}

// SetWithTags adds the item identified by key with value to the cache.
// One or more tags can be passed to tag the item.
func (c *Cache) SetWithTags(key string, value []byte, tags ...string) error {
	if value == nil {
		return nil
	}

	var it item
	result := c.db.Where("key = ?", key).First(&it)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			var tagsNew []tag
			for _, t := range tags {
				tagsNew = append(tagsNew, tag{Name: t})
			}
			itemNew := &item{
				Key:   key,
				Value: value,
				Tags:  tagsNew,
			}
			resultNewSave := c.db.Save(itemNew)
			if resultNewSave.Error != nil {
				return fmt.Errorf("set cache item with tag: %w", resultNewSave.Error)
			}

			return nil
		}

		return fmt.Errorf("get cache item to set with tag: %w", result.Error)
	}

	return c.db.Transaction(func(tx *gorm.DB) error {
		resultDeleteTags := tx.Where("item_id = ?", it.ID).Delete(&tag{})
		if resultDeleteTags.Error != nil {
			return fmt.Errorf("delete tags of cache item: %w", resultDeleteTags.Error)
		}

		var tagsNew []tag
		for _, t := range tags {
			tagsNew = append(tagsNew, tag{Name: t})
		}

		it.Value = value
		it.Tags = tagsNew
		resultSaveCurrent := tx.Save(it)
		if resultSaveCurrent.Error != nil {
			return fmt.Errorf("update existing cache item with tag: %w", resultSaveCurrent.Error)
		}

		return nil
	})
}
