package http

import (
	"encoding/json"
	"net/http"

	infraSecurity "go-ddd/internal/infrastructure/security"
	"go-ddd/internal/presentation/authguard"
	"go-ddd/pkg/telemetry"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(userHandler *UserHandler, jwtService *infraSecurity.JWTService, aclHandlers ...*ACLHandler) http.Handler {
	return NewRouterWithLimiter(userHandler, jwtService, authguard.New(), aclHandlers...)
}

func NewRouterWithLimiter(userHandler *UserHandler, jwtService *infraSecurity.JWTService, limiter *authguard.Limiter, aclHandlers ...*ACLHandler) http.Handler {
	mux := http.NewServeMux()

	// Swagger UI
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Documentation routes
	docsHandler := NewDocsHandler("docs/assets/manifest.json", "docs/assets")
	mux.HandleFunc("GET /docs", docsHandler.ServeDocs)
	mux.HandleFunc("GET /docs/assets/{filename}", docsHandler.ServeAssets)

	// Prometheus metrics endpoint
	mux.Handle("GET /metrics", telemetry.PrometheusHandler())

	// Public routes
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.Handle("POST /api/v1/auth/register", limitAuth(limiter, http.HandlerFunc(userHandler.Register)))
	mux.Handle("POST /api/v1/auth/login", limitAuth(limiter, http.HandlerFunc(userHandler.Login)))

	// Protected routes (require JWT authentication)
	authMiddleware := AuthMiddleware(jwtService)
	mux.Handle("GET /api/v1/users/me", authMiddleware(http.HandlerFunc(userHandler.GetProfile)))

	if len(aclHandlers) > 0 && aclHandlers[0] != nil {
		aclHandler := aclHandlers[0]
		aclAdmin := func(handler http.HandlerFunc) http.Handler {
			return authMiddleware(aclHandler.RequireAuthAdmin(handler))
		}
		domainAdmin := func(handler http.HandlerFunc) http.Handler {
			return authMiddleware(aclHandler.RequireAnyAdmin(handler))
		}
		mux.Handle("GET /api/v1/acl/domains", domainAdmin(aclHandler.ListDomains))
		mux.Handle("POST /api/v1/acl/domains", aclAdmin(aclHandler.CreateDomain))
		mux.Handle("GET /api/v1/acl/domains/{id}", domainAdmin(aclHandler.GetDomain))
		mux.Handle("PUT /api/v1/acl/domains/{id}", domainAdmin(aclHandler.UpdateDomain))
		mux.Handle("DELETE /api/v1/acl/domains/{id}", aclAdmin(aclHandler.DeleteDomain))
		mux.Handle("GET /api/v1/acl/roles", domainAdmin(aclHandler.ListRoles))
		mux.Handle("POST /api/v1/acl/roles", domainAdmin(aclHandler.CreateRole))
		mux.Handle("GET /api/v1/acl/roles/{id}", domainAdmin(aclHandler.GetRole))
		mux.Handle("PUT /api/v1/acl/roles/{id}", domainAdmin(aclHandler.UpdateRole))
		mux.Handle("DELETE /api/v1/acl/roles/{id}", domainAdmin(aclHandler.DeleteRole))
		mux.Handle("GET /api/v1/acl/permissions", domainAdmin(aclHandler.ListPermissions))
		mux.Handle("POST /api/v1/acl/permissions", aclAdmin(aclHandler.CreatePermission))
		mux.Handle("GET /api/v1/acl/permissions/{id}", domainAdmin(aclHandler.GetPermission))
		mux.Handle("PUT /api/v1/acl/permissions/{id}", aclAdmin(aclHandler.UpdatePermission))
		mux.Handle("DELETE /api/v1/acl/permissions/{id}", aclAdmin(aclHandler.DeletePermission))
	}

	// Apply global middleware chain (Observability -> Recovery -> Router)
	var handler http.Handler = mux
	handler = limitBody(handler)
	handler = RecoveryMiddleware(handler)
	handler = telemetry.ObservabilityMiddleware(handler)

	return handler
}
