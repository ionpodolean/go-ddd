package main

import (
	"strings"
	"testing"
)

func TestSecurityConfig(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("AUTH_ADMIN_EMAIL", " admin@example.com ")
	t.Setenv("AUTH_ADMIN_PASSWORD", "a-strong-password")
	t.Setenv("AUTH_ADMIN_FIRST_NAME", "")
	t.Setenv("AUTH_ADMIN_LAST_NAME", "")

	secret, admin, err := securityConfig()
	if err != nil || len(secret) != 32 || admin.Email != "admin@example.com" {
		t.Fatalf("valid configuration: admin=%+v, err=%v", admin, err)
	}

	for _, tc := range []struct{ name, key, value string }{
		{"missing secret", "JWT_SECRET", ""},
		{"short secret", "JWT_SECRET", "short"},
		{"missing email", "AUTH_ADMIN_EMAIL", " "},
		{"short password", "AUTH_ADMIN_PASSWORD", "short"},
		{"old default password", "AUTH_ADMIN_PASSWORD", "change-me-now"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, _, err := securityConfig(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
