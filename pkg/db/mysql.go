package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"go-ddd/config"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB() *sql.DB {
	DBConfig := config.GetDatabaseConfig()
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", DBConfig.User, DBConfig.Password, DBConfig.Host, DBConfig.Port, DBConfig.Name)

	log.Printf("Connecting to database at %s:%s...", DBConfig.Host, DBConfig.Port)

	var DB *sql.DB
	var err error

	maxRetries := 15
	for i := 1; i <= maxRetries; i++ {
		DB, err = sql.Open("mysql", connectionString)
		if err == nil {
			err = DB.Ping()
			if err == nil {
				log.Println("Database connection established successfully")
				return DB
			}
		}

		log.Printf("Waiting for database to be ready (attempt %d/%d): %v", i, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	panic(fmt.Errorf("failed to connect to database at %s:%s after %d attempts: %w", DBConfig.Host, DBConfig.Port, maxRetries, err))
}
