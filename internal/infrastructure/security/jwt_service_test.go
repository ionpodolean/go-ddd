package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateTokenRequiresExpirationAndValidIdentity(t *testing.T) {
	const secret = "test-secret"
	svc := NewJWTService(secret, time.Hour)
	valid, err := svc.GenerateToken(42, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ValidateToken(valid); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	for _, tc := range []struct {
		name   string
		method jwt.SigningMethod
		claims JWTClaims
	}{
		{"missing expiry", jwt.SigningMethodHS256, JWTClaims{UserID: 42}},
		{"expired", jwt.SigningMethodHS256, JWTClaims{UserID: 42, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))}}},
		{"missing user", jwt.SigningMethodHS256, JWTClaims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}},
		{"wrong algorithm", jwt.SigningMethodHS384, JWTClaims{UserID: 42, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			signed, err := jwt.NewWithClaims(tc.method, tc.claims).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := svc.ValidateToken(signed); err == nil {
				t.Fatal("invalid token accepted")
			}
		})
	}
}
