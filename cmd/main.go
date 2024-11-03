package main

import (
	"go-ddd/internal/application"
	"go-ddd/internal/infrastructure"
	"go-ddd/internal/presentation"
	"go-ddd/pkg/db"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	dbSql := db.InitDB()

	defer dbSql.Close()

	db.RunMigrations(dbSql)

	userRepository := infrastructure.NewMySQLUserRepository(dbSql)
	userService := application.NewUserService(userRepository)
	userHandler := presentation.NewUserHandler(userService)

	router := gin.Default()

	router.POST("/register", func(c *gin.Context) {
		userHandler.RegisterUser(c)
	})
	router.POST("/login", func(c *gin.Context) {
		userHandler.AuthenticateUser(c)
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
