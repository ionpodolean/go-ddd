package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-ddd/config"
	domainMail "go-ddd/internal/domain/mail"
)

type MailtrapAPIMailer struct {
	cfg        config.MailConfig
	httpClient *http.Client
	apiURL     string
}

type mailtrapAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type mailtrapPayload struct {
	From    mailtrapAddress   `json:"from"`
	To      []mailtrapAddress `json:"to"`
	Subject string            `json:"subject"`
	Text    string            `json:"text,omitempty"`
	HTML    string            `json:"html,omitempty"`
}

func NewMailtrapAPIMailer(cfg config.MailConfig) *MailtrapAPIMailer {
	return &MailtrapAPIMailer{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiURL: "https://send.api.mailtrap.io/api/send",
	}
}

// SetAPIURL allows overriding the Mailtrap API URL for sandbox or mock server testing
func (m *MailtrapAPIMailer) SetAPIURL(url string) {
	m.apiURL = url
}

func (m *MailtrapAPIMailer) Send(ctx context.Context, msg domainMail.Message) error {
	from := mailtrapAddress{
		Email: msg.From.Email,
		Name:  msg.From.Name,
	}
	if from.Email == "" {
		from.Email = m.cfg.FromAddress
	}
	if from.Name == "" {
		from.Name = m.cfg.FromName
	}

	var toList []mailtrapAddress
	for _, recipient := range msg.To {
		toList = append(toList, mailtrapAddress{
			Email: recipient.Email,
			Name:  recipient.Name,
		})
	}

	if len(toList) == 0 {
		return fmt.Errorf("mailtrap api mailer: no recipients specified")
	}

	payload := mailtrapPayload{
		From:    from,
		To:      toList,
		Subject: msg.Subject,
		Text:    msg.TextBody,
		HTML:    msg.HTMLBody,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mailtrap api mailer: failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.apiURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("mailtrap api mailer: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if m.cfg.APIToken != "" {
		req.Header.Set("Api-Token", m.cfg.APIToken)
		req.Header.Set("Authorization", "Bearer "+m.cfg.APIToken)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mailtrap api mailer: http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailtrap api mailer: returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
