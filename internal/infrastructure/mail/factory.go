package mail

import (
	"fmt"

	"go-ddd/config"
	domainMail "go-ddd/internal/domain/mail"
)

func NewMailer(cfg config.MailConfig) (domainMail.Mailer, error) {
	switch cfg.Driver {
	case config.DriverSMTP:
		return NewSMTPMailer(cfg), nil
	case config.DriverAPI:
		return NewMailtrapAPIMailer(cfg), nil
	case config.DriverLog:
		return NewLogMailer(), nil
	default:
		return nil, fmt.Errorf("unsupported mail driver: %s", cfg.Driver)
	}
}
