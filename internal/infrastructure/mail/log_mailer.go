package mail

import (
	"context"
	"log"

	domainMail "go-ddd/internal/domain/mail"
)

type LogMailer struct{}

func NewLogMailer() *LogMailer {
	return &LogMailer{}
}

func (m *LogMailer) Send(ctx context.Context, msg domainMail.Message) error {
	log.Printf("[MAIL LOG DRIVER] From: %s <%s> | To: %v | Subject: %q\nText Body:\n%s\nHTML Body:\n%s",
		msg.From.Name, msg.From.Email, msg.To, msg.Subject, msg.TextBody, msg.HTMLBody)
	return nil
}
