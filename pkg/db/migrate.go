package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/wndhydrnt/saturn-bot/pkg/db/golangmigrate/sqlite3"
)

// Migrator is a function that executes migrations.
type Migrator func(db *sql.DB) error

// Migrate creates a new [Migrator] from source.
// The migrations are run against a sqlite3 db.
func Migrate(source fs.FS) Migrator {
	return func(db *sql.DB) error {
		driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("create golang-migrate driver for sqlite3: %w", err)
		}

		d, err := iofs.New(source, "migrations")
		if err != nil {
			return fmt.Errorf("create golang-migrate FS source: %w", err)
		}

		m, err := migrate.NewWithInstance("iofs", d, "sqlite3", driver)
		if err != nil {
			return fmt.Errorf("create instance of golang-migrate: %w", err)
		}

		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("up migration: %w", err)
		}

		return nil
	}
}
