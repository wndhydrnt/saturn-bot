package db

import (
	"fmt"

	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func New(migrate bool) (*gorm.DB, error) {
	db, err := gorm.Open(gormlite.Open("gorm.db"))
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
