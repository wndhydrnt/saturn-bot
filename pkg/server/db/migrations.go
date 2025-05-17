package db

import (
	"embed"
	"io/fs"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrations returns all migrations bundled in a filesystem object.
func Migrations() fs.FS {
	return migrations
}
