// Package migrations embeds the versioned Goose SQL migrations in both binaries.
package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var files embed.FS

func Run(db *sql.DB, command string) error {
	goose.SetBaseFS(files)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Run(command, db, ".")
}
