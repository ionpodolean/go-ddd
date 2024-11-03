package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB) {
	fmt.Println("Running migrations...")

	migrationsDir := "./migrations"

	if err := goose.SetDialect("mysql"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		fmt.Printf("Failed to run migrations: %v", err)
	}

	fmt.Println("Migrations completed")
}
