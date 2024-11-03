package db

import (
	"database/sql"
	"fmt"
	"log"

	"go-ddd/config"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB() *sql.DB {
	DBConfig := config.GetDatabaseConfig()
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", DBConfig.User, DBConfig.Password, DBConfig.Host, DBConfig.Port, DBConfig.Name)

	fmt.Println(connectionString)

	DB, err := sql.Open("mysql", connectionString)

	if err != nil {
		panic(fmt.Errorf("sql.Open: %w", err))
	}

	if err = DB.Ping(); err != nil {
		panic(fmt.Errorf("DB.Ping: %w", err))
	}

	log.Println("Database connection established")
	return DB
}
