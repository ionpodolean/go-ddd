package user_test

import (
	"context"
	"strings"
	"testing"
	"time"

	appUser "go-ddd/internal/application/user"
	domainMail "go-ddd/internal/domain/mail"
	domainUser "go-ddd/internal/domain/user"
)

type mockUserRepository struct {
	users map[string]*domainUser.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*domainUser.User),
	}
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

type mockTokenService struct{}

type captureMailer struct{ messages chan domainMail.Message }

func (m captureMailer) Send(_ context.Context, msg domainMail.Message) error {
	m.messages <- msg
	return nil
}

func TestWelcomeEmailEscapesUserName(t *testing.T) {
	mailer := captureMailer{messages: make(chan domainMail.Message, 1)}
	svc := appUser.NewUserService(newMockUserRepository(), &mockTokenService{}, mailer)
	_, err := svc.Register(context.Background(), appUser.RegisterRequest{Email: "x@example.com", Password: "password123", FirstName: "<img src=x>", LastName: "User"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case msg := <-mailer.messages:
		if strings.Contains(msg.HTMLBody, "<img") || !strings.Contains(msg.HTMLBody, "&lt;img src=x&gt;") {
			t.Fatalf("unsafe HTML body: %s", msg.HTMLBody)
		}
	case <-time.After(time.Second):
		t.Fatal("welcome message not sent")
	}
}

func (m *mockTokenService) GenerateToken(userID int64, email string) (string, error) {
	return "mock-jwt-token", nil
}

func TestUserService_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepository()
	tokenSvc := &mockTokenService{}
	svc := appUser.NewUserService(repo, tokenSvc, nil)
	ctx := context.Background()

	t.Run("register new user successfully", func(t *testing.T) {
		req := appUser.RegisterRequest{
			Email:     "jane@example.com",
			Password:  "password123",
			FirstName: "Jane",
			LastName:  "Doe",
		}

		res, err := svc.Register(ctx, req)
		if err != nil {
			t.Fatalf("expected registration to succeed, got %v", err)
		}
		if res.Token != "mock-jwt-token" {
			t.Errorf("expected token mock-jwt-token, got %s", res.Token)
		}
		if res.User.Email != "jane@example.com" {
			t.Errorf("expected user email jane@example.com, got %s", res.User.Email)
		}
	})

	t.Run("register duplicate user fails", func(t *testing.T) {
		req := appUser.RegisterRequest{
			Email:     "jane@example.com",
			Password:  "password123",
			FirstName: "Jane",
			LastName:  "Doe",
		}

		_, err := svc.Register(ctx, req)
		if err != domainUser.ErrUserAlreadyExists {
			t.Errorf("expected ErrUserAlreadyExists, got %v", err)
		}
	})

	t.Run("login with correct credentials", func(t *testing.T) {
		req := appUser.LoginRequest{
			Email:    "jane@example.com",
			Password: "password123",
		}

		res, err := svc.Login(ctx, req)
		if err != nil {
			t.Fatalf("expected login to succeed, got %v", err)
		}
		if res.Token != "mock-jwt-token" {
			t.Errorf("expected token mock-jwt-token, got %s", res.Token)
		}
	})

	t.Run("login with wrong password fails", func(t *testing.T) {
		req := appUser.LoginRequest{
			Email:    "jane@example.com",
			Password: "wrongpassword",
		}

		_, err := svc.Login(ctx, req)
		if err != domainUser.ErrInvalidCredentials {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}
