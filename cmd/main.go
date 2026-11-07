package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"go-ddd/config"
	_ "go-ddd/docs"
	appACL "go-ddd/internal/application/acl"
	appUser "go-ddd/internal/application/user"
	infraMail "go-ddd/internal/infrastructure/mail"
	infraPersistence "go-ddd/internal/infrastructure/persistence"
	infraSecurity "go-ddd/internal/infrastructure/security"
	"go-ddd/internal/presentation/authguard"
	presentGRPC "go-ddd/internal/presentation/grpc"
	presentHTTP "go-ddd/internal/presentation/http"
	"go-ddd/pkg/db"
	"go-ddd/pkg/telemetry"
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
	ctx := context.Background()
	jwtSecret, admin, err := securityConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid security configuration")
	}

	// Initialize telemetry (Zerolog + OpenTelemetry Traces & Metrics)
	shutdownTelemetry, err := telemetry.InitTelemetry(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize telemetry")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Error shutting down telemetry")
		}
	}()

	dbSql := db.InitDB()
	defer dbSql.Close()

	if err := db.RunMigrations(dbSql); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	userRepo := infraPersistence.NewMySQLUserRepository(dbSql)
	aclRepo := infraPersistence.NewMySQLACLRepository(dbSql)
	aclService := appACL.NewService(aclRepo)
	if err := aclService.SeedAuth(ctx, admin); err != nil {
		log.Fatal().Err(err).Msg("Failed to seed Auth ACL domain")
	}

	jwtService := infraSecurity.NewJWTService(jwtSecret, 24*time.Hour)

	mailCfg := config.GetMailConfig()
	mailer, err := infraMail.NewMailer(mailCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize mailer")
	}
	log.Info().Str("driver", string(mailCfg.Driver)).Msg("Mail service initialized")

	userService := appUser.NewUserService(userRepo, jwtService, mailer)
	userHandler := presentHTTP.NewUserHandler(userService)
	aclHandler := presentHTTP.NewACLHandler(aclService)

	limiter := authguard.New()
	router := presentHTTP.NewRouterWithLimiter(userHandler, jwtService, limiter, aclHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	bindAddress := envOrDefault("BIND_ADDRESS", "127.0.0.1")
	srv := &http.Server{
		Addr:              net.JoinHostPort(bindAddress, port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	grpcServer := presentGRPC.NewServerWithLimiter(userService, jwtService, limiter)
	grpcListener, err := net.Listen("tcp", net.JoinHostPort(bindAddress, grpcPort))
	if err != nil {
		log.Fatal().Err(err).Str("port", grpcPort).Msg("Failed to listen for gRPC server")
	}

	// Start server in a background goroutine
	go func() {
		log.Info().Str("port", port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Server ListenAndServe error")
		}
	}()
	go func() {
		log.Info().Str("port", grpcPort).Msg("gRPC server starting")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Error().Err(err).Msg("gRPC server error")
		}
	}()

	// Listen for OS interrupt signals (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Received shutdown signal. Gracefully shutting down servers...")

	// Give active connections up to 5 seconds to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()
	select {
	case <-grpcStopped:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}

	log.Info().Msg("Server shutdown complete")
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func securityConfig() (string, appACL.AdminUserRequest, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return "", appACL.AdminUserRequest{}, errors.New("JWT_SECRET must contain at least 32 bytes")
	}
	admin := appACL.AdminUserRequest{
		Email:     strings.TrimSpace(os.Getenv("AUTH_ADMIN_EMAIL")),
		Password:  os.Getenv("AUTH_ADMIN_PASSWORD"),
		FirstName: envOrDefault("AUTH_ADMIN_FIRST_NAME", "Auth"),
		LastName:  envOrDefault("AUTH_ADMIN_LAST_NAME", "Admin"),
	}
	if admin.Email == "" || len(admin.Password) < 12 || admin.Password == "change-me-now" {
		return "", appACL.AdminUserRequest{}, errors.New("AUTH_ADMIN_EMAIL and an AUTH_ADMIN_PASSWORD of at least 12 bytes are required")
	}
	return secret, admin, nil
}
