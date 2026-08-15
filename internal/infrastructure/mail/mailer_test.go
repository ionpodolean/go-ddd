package mail

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-ddd/config"
	domainMail "go-ddd/internal/domain/mail"
)

func TestLogMailer_Send(t *testing.T) {
	mailer := NewLogMailer()
	msg := domainMail.Message{
		From:     domainMail.Address{Name: "App", Email: "noreply@example.com"},
		To:       []domainMail.Address{{Name: "User", Email: "user@example.com"}},
		Subject:  "Test Email",
		TextBody: "Hello Text",
		HTMLBody: "<p>Hello HTML</p>",
	}

	err := mailer.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("expected no error from LogMailer, got %v", err)
	}
}

func TestMailtrapAPIMailer_Send(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Api-Token") != "test-api-token" {
			t.Errorf("expected Api-Token header test-api-token, got %s", r.Header.Get("Api-Token"))
		}

		bodyBytes, _ := io.ReadAll(r.Body)
		bodyStr := string(bodyBytes)

		if !strings.Contains(bodyStr, "Welcome to Go-DDD") {
			t.Errorf("expected body to contain subject, got %s", bodyStr)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer mockServer.Close()

	cfg := config.MailConfig{
		Driver:      config.DriverAPI,
		APIToken:    "test-api-token",
		FromAddress: "noreply@example.com",
		FromName:    "Go-DDD",
	}

	apiMailer := NewMailtrapAPIMailer(cfg)
	apiMailer.SetAPIURL(mockServer.URL)

	msg := domainMail.Message{
		To:       []domainMail.Address{{Name: "Jane", Email: "jane@example.com"}},
		Subject:  "Welcome to Go-DDD",
		TextBody: "Welcome Jane!",
		HTMLBody: "<h1>Welcome Jane!</h1>",
	}

	err := apiMailer.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("expected no error from MailtrapAPIMailer, got %v", err)
	}
}

func TestNewMailer_Factory(t *testing.T) {
	tests := []struct {
		driver config.MailDriver
	}{
		{driver: config.DriverLog},
		{driver: config.DriverSMTP},
		{driver: config.DriverAPI},
	}

	for _, tt := range tests {
		cfg := config.MailConfig{Driver: tt.driver}
		mailer, err := NewMailer(cfg)
		if err != nil {
			t.Errorf("expected no error for driver %s, got %v", tt.driver, err)
		}
		if mailer == nil {
			t.Errorf("expected mailer instance for driver %s", tt.driver)
		}
	}
}
