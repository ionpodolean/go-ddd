package config

import (
	"os"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
}

func init() {
	// Try loading .env file if available, ignore error if missing (e.g. in docker environment)
	_ = godotenv.Load()
}

func GetDatabaseConfig() DBConfig {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "user"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}

	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "data"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}

	return DBConfig{
		Host:     host,
		User:     user,
		Password: password,
		Name:     name,
		Port:     port,
	}
}
