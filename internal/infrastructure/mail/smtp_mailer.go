package mail

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"go-ddd/config"
	domainMail "go-ddd/internal/domain/mail"
)

type SMTPMailer struct {
	cfg config.MailConfig
}

func NewSMTPMailer(cfg config.MailConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(ctx context.Context, msg domainMail.Message) error {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	// Determine sender and recipient strings
	from := msg.From.Email
	if from == "" {
		from = m.cfg.FromAddress
	}
	fromName := msg.From.Name
	if fromName == "" {
		fromName = m.cfg.FromName
	}
	if strings.ContainsAny(from+fromName+msg.Subject, "\r\n") {
		return fmt.Errorf("smtp mailer: invalid header value")
	}

	var recipients []string
	var toHeaderParts []string
	for _, recipient := range msg.To {
		if strings.ContainsAny(recipient.Name+recipient.Email, "\r\n") {
			return fmt.Errorf("smtp mailer: invalid header value")
		}
		recipients = append(recipients, recipient.Email)
		toHeaderParts = append(toHeaderParts, (&mail.Address{Name: recipient.Name, Address: recipient.Email}).String())
	}

	if len(recipients) == 0 {
		return fmt.Errorf("smtp mailer: no recipients specified")
	}

	// Prepare MIME email body
	var bodyBuilder strings.Builder
	bodyBuilder.WriteString(fmt.Sprintf("From: %s\r\n", (&mail.Address{Name: fromName, Address: from}).String()))
	bodyBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(toHeaderParts, ", ")))
	bodyBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	bodyBuilder.WriteString("MIME-Version: 1.0\r\n")

	if msg.HTMLBody != "" {
		bodyBuilder.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		bodyBuilder.WriteString(msg.HTMLBody)
	} else {
		bodyBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		bodyBuilder.WriteString(msg.TextBody)
	}

	var auth smtp.Auth
	if m.cfg.Username != "" || m.cfg.Password != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	return smtp.SendMail(addr, auth, from, recipients, []byte(bodyBuilder.String()))
}
