package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	sqlitePragmaJournalMode = "PRAGMA journal_mode = WAL;"
	sqlitePragmaSynchronous = "PRAGMA synchronous = NORMAL;"
	sqlitePragmaCacheSize   = "PRAGMA cache_size = 1000000000;"
	sqlitePragmaForeignKeys = "PRAGMA foreign_keys = true;"
	sqlitePragmaBusyTimeout = "PRAGMA busy_timeout = 5000;"
)

func New(enableLog bool, path string, migrate Migrator) (*gorm.DB, error) {
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

	cfg := &gorm.Config{}
	if enableLog {
		cfg.Logger = logger.Default
	} else {
		cfg.Logger = logger.Discard
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open sql driver sqlite3: %w", err)
	}

	gormDb, err := gorm.Open(gormlite.OpenDB(db), cfg)
	if err != nil {
		return nil, err
	}

	err = configureSqlite(gormDb, sqlitePragmaJournalMode, sqlitePragmaSynchronous, sqlitePragmaCacheSize, sqlitePragmaForeignKeys, sqlitePragmaBusyTimeout)
	if err != nil {
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}

	if migrate != nil {
		err := migrate(db)
		if err != nil {
			return nil, fmt.Errorf("db migration failed: %w", err)
		}
	}

	return gormDb, nil
}

func configureSqlite(db *gorm.DB, stmts ...string) error {
	for _, stmt := range stmts {
		if tx := db.Exec(stmt); tx.Error != nil {
			return fmt.Errorf("execute '%s': %w", stmt, tx.Error)
		}
	}

	return nil
}
