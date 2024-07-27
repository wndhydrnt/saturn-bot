package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func New(migrate bool, path string) (*gorm.DB, error) {
	dir := filepath.Dir(path)
	if dir != "" {
		_, err := os.Stat(dir)
		if errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return nil, fmt.Errorf("create directory for database: %w", err)
			}
		}
	}

	db, err := gorm.Open(gormlite.Open(path))
	if err != nil {
		return nil, err
	}

	if migrate {
		err := db.AutoMigrate(&Run{}, &Task{}, &TaskResult{})
		if err != nil {
			return nil, fmt.Errorf("db migration failed: %w", err)
		}
	}

	return db, nil
}
