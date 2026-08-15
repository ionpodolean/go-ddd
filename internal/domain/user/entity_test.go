package user_test

import (
	"testing"

	"go-ddd/internal/domain/user"
)

func TestNewUser(t *testing.T) {
	t.Run("valid user creation", func(t *testing.T) {
		u, err := user.NewUser("john@example.com", "secret123", "John", "Doe")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.Email != "john@example.com" {
			t.Errorf("expected email john@example.com, got %s", u.Email)
		}
		if u.FirstName != "John" || u.LastName != "Doe" {
			t.Errorf("expected name John Doe, got %s %s", u.FirstName, u.LastName)
		}
		if !u.CheckPassword("secret123") {
			t.Errorf("expected password verification to succeed")
		}
		if u.CheckPassword("wrongpassword") {
			t.Errorf("expected password verification to fail for wrong password")
		}
	})

	t.Run("invalid email format", func(t *testing.T) {
		_, err := user.NewUser("not-an-email", "secret123", "John", "Doe")
		if err != user.ErrInvalidEmail {
			t.Errorf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("short password", func(t *testing.T) {
		_, err := user.NewUser("john@example.com", "123", "John", "Doe")
		if err != user.ErrInvalidPassword {
			t.Errorf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("empty names", func(t *testing.T) {
		_, err := user.NewUser("john@example.com", "secret123", "", "Doe")
		if err != user.ErrEmptyFields {
			t.Errorf("expected ErrEmptyFields, got %v", err)
		}
	})
}
