package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appUser "go-ddd/internal/application/user"
	domainUser "go-ddd/internal/domain/user"
	infraSecurity "go-ddd/internal/infrastructure/security"
	presentHTTP "go-ddd/internal/presentation/http"
)

type mockUserRepository struct {
	users map[string]*domainUser.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{users: make(map[string]*domainUser.User)}
}

func (m *mockUserRepository) Create(ctx context.Context, u *domainUser.User) error {
	u.ID = int64(len(m.users) + 1)
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, domainUser.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*domainUser.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domainUser.ErrUserNotFound
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := m.users[email]
	return ok, nil
}

func TestHTTPHandlersAndMiddleware(t *testing.T) {
	repo := newMockUserRepository()
	jwtSvc := infraSecurity.NewJWTService("test-secret-key", 1*time.Hour)
	userSvc := appUser.NewUserService(repo, jwtSvc, nil)
	userHandler := presentHTTP.NewUserHandler(userSvc)
	router := presentHTTP.NewRouter(userHandler, jwtSvc)

	var authToken string

	t.Run("POST /api/v1/auth/register success", func(t *testing.T) {
		body, _ := json.Marshal(appUser.RegisterRequest{
			Email:     "user@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d", rec.Code)
		}

		var res appUser.AuthResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if res.Token == "" {
			t.Errorf("expected non-empty JWT token")
		}
		authToken = res.Token
	})

	t.Run("POST /api/v1/auth/login success", func(t *testing.T) {
		body, _ := json.Marshal(appUser.LoginRequest{
			Email:    "user@example.com",
			Password: "password123",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}
	})

	t.Run("GET /api/v1/users/me without token fails 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 Unauthorized, got %d", rec.Code)
		}
	})

	t.Run("GET /api/v1/users/me with valid Bearer token succeeds 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var userRes appUser.UserDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &userRes); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if userRes.Email != "user@example.com" {
			t.Errorf("expected email user@example.com, got %s", userRes.Email)
		}
	})

	t.Run("GET /health success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}
	})
}
