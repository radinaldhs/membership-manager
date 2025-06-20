package app

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

const (
	sqliteMigrationFilesDir = "./data/database/sqlite/migrations"
)

func RunSQLiteDatabaseMigration(db *sql.DB) error {
	goose.SetDialect(string(goose.DialectSQLite3))
	if err := goose.Up(db, sqliteMigrationFilesDir); err != nil {
		return fmt.Errorf("sqlite database migration error: %v", err)
	}

	return nil
}
