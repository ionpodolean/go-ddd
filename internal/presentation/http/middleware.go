package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	infraSecurity "go-ddd/internal/infrastructure/security"
	"go-ddd/pkg/telemetry"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	EmailKey  contextKey = "email"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := telemetry.Logger(r.Context())
				logger.Error().
					Interface("panic", err).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("Panic recovered")
				respondError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(jwtService *infraSecurity.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				respondError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			claims, err := jwtService.ValidateToken(parts[1])
			if err != nil {
				log.Warn().Err(err).Str("path", r.URL.Path).Msg("Invalid or expired token")
				respondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
