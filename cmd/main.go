package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go-ddd/docs"
	appUser "go-ddd/internal/application/user"
	infraPersistence "go-ddd/internal/infrastructure/persistence"
	infraSecurity "go-ddd/internal/infrastructure/security"
	presentHTTP "go-ddd/internal/presentation/http"
	"go-ddd/pkg/db"
)

// @title           Go DDD API
// @version         1.0
// @description     Clean Domain-Driven Design REST API with User Registration, Login, and JWT Authentication.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer " followed by a space and JWT token.
func main() {
	dbSql := db.InitDB()
	defer dbSql.Close()

	db.RunMigrations(dbSql)

	userRepo := infraPersistence.NewMySQLUserRepository(dbSql)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-jwt-key"
	}
	jwtService := infraSecurity.NewJWTService(jwtSecret, 24*time.Hour)

	userService := appUser.NewUserService(userRepo, jwtService)
	userHandler := presentHTTP.NewUserHandler(userService)

	router := presentHTTP.NewRouter(userHandler, jwtService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a background goroutine
	go func() {
		log.Printf("Server starting on port :%s ...", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	// Listen for OS interrupt signals (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Received shutdown signal. Gracefully shutting down HTTP server...")

	// Give active connections up to 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown complete.")
}
