package config

import (
	"os"
	"strconv"
	"strings"
)

type MailDriver string

const (
	DriverSMTP MailDriver = "smtp"
	DriverAPI  MailDriver = "api"
	DriverLog  MailDriver = "log"
)

type MailConfig struct {
	Driver      MailDriver
	Host        string
	Port        int
	Username    string
	Password    string
	APIToken    string
	FromAddress string
	FromName    string
}

func GetMailConfig() MailConfig {
	driverStr := strings.ToLower(os.Getenv("MAIL_DRIVER"))
	var driver MailDriver
	switch driverStr {
	case "smtp":
		driver = DriverSMTP
	case "api", "mailtrap_api":
		driver = DriverAPI
	case "log":
		driver = DriverLog
	default:
		driver = DriverSMTP
	}

	host := os.Getenv("MAIL_HOST")
	if host == "" {
		host = "sandbox.smtp.mailtrap.io"
	}

	portStr := os.Getenv("MAIL_PORT")
	port := 2525
	if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
		port = p
	}

	fromAddr := os.Getenv("MAIL_FROM_ADDRESS")
	if fromAddr == "" {
		fromAddr = "noreply@example.com"
	}

	fromName := os.Getenv("MAIL_FROM_NAME")
	if fromName == "" {
		fromName = "Go-DDD Application"
	}

	return MailConfig{
		Driver:      driver,
		Host:        host,
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		APIToken:    os.Getenv("MAIL_API_TOKEN"),
		FromAddress: fromAddr,
		FromName:    fromName,
	}
}
