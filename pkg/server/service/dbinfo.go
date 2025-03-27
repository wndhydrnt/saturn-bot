package service

import (
	"fmt"

	"gorm.io/gorm"
)

// DbInfo provides information about the sqlite database.
type DbInfo struct {
	db *gorm.DB
}

// NewDbInfo returns a new [DbInfo].
func NewDbInfo(db *gorm.DB) *DbInfo {
	return &DbInfo{
		db: db,
	}
}

// Size returns the size of the sqlite database file in bytes.
func (di *DbInfo) Size() (float64, error) {
	var size float64
	tx := di.db.Raw("select page_size * page_count from pragma_page_count(), pragma_page_size()").
		Scan(&size)
	if tx.Error != nil {
		return 0, fmt.Errorf("read size of database: %w", tx.Error)
	}

	return size, nil
}
