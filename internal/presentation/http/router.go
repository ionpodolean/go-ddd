package http

import (
	"encoding/json"
	"net/http"

	infraSecurity "go-ddd/internal/infrastructure/security"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(userHandler *UserHandler, jwtService *infraSecurity.JWTService) http.Handler {
	mux := http.NewServeMux()

	// Swagger UI
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Documentation routes
	docsHandler := NewDocsHandler("docs/assets/manifest.json", "docs/assets")
	mux.HandleFunc("GET /docs", docsHandler.ServeDocs)
	mux.HandleFunc("GET /docs/assets/{filename}", docsHandler.ServeAssets)

	// Public routes
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/v1/auth/register", userHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", userHandler.Login)

	// Protected routes (require JWT authentication)
	authMiddleware := AuthMiddleware(jwtService)
	mux.Handle("GET /api/v1/users/me", authMiddleware(http.HandlerFunc(userHandler.GetProfile)))

	// Apply global middleware chain (Logging -> Recovery -> Router)
	var handler http.Handler = mux
	handler = LoggingMiddleware(handler)
	handler = RecoveryMiddleware(handler)

	return handler
}
