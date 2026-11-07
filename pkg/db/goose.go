package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("configure migrations: %w", err)
	}
	if err := goose.Up(db, "./migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
