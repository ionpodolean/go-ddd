package config

import (
	"log"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func GetDatabaseConfig() DBConfig {
	return DBConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Port:     os.Getenv("DB_PORT"),
	}
}
