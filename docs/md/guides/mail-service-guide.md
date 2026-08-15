# Mail Service Integration Guide

<!-- metadata
title: Mail Service Integration Guide
category: Guides
status: active
last_updated: 2026-08-15
-->

This guide describes the Mail Service architecture in **go-ddd**, supporting **SMTP**, **Mailtrap API**, and local **Log** drivers.

## Related Documentation
- [Architecture Reference](/docs?page=architecture)
- [User Management Module](/docs?page=user-module)
- [External Client Integration Guide](/docs?page=external-client-integration)

---

## Architecture Overview

<!-- covers: internal/domain/mail/**, internal/infrastructure/mail/** -->

The Mail Service implements the DDD Strategy pattern. The domain defines a generic `Mailer` interface, allowing drivers to be swapped dynamically via environment variables without changing business logic.

```
                  +-------------------------+
                  |  domain/mail.Mailer     |
                  |  (Send method interface)|
                  +-------------------------+
                               ^
                               |
        +----------------------+----------------------+
        |                      |                      |
+---------------+      +---------------+      +---------------+
|  SMTPMailer   |      | MailtrapAPI   |      |   LogMailer   |
|  (net/smtp)   |      |  (HTTP API)   |      | (stdout dev)  |
+---------------+      +---------------+      +---------------+
```

---

## Driver Configuration (`.env`)

Choose the driver using `MAIL_DRIVER`:

### 1. SMTP Driver (Mailtrap Sandbox or Production SMTP)
```env
MAIL_DRIVER=smtp
MAIL_HOST=sandbox.smtp.mailtrap.io
MAIL_PORT=2525
MAIL_USERNAME=your_smtp_username
MAIL_PASSWORD=your_smtp_password
MAIL_FROM_ADDRESS=noreply@example.com
MAIL_FROM_NAME="Go-DDD App"
```

### 2. Mailtrap API Driver (HTTP Sending API)
```env
MAIL_DRIVER=api
MAIL_API_TOKEN=your_mailtrap_api_token
MAIL_FROM_ADDRESS=noreply@example.com
MAIL_FROM_NAME="Go-DDD App"
```

### 3. Log Driver (Development & Testing)
Logs emails to stdout without sending external network requests:
```env
MAIL_DRIVER=log
```

---

## Code Example: Sending Emails

```go
package example

import (
    "context"
    domainMail "go-ddd/internal/domain/mail"
)

func SendWelcome(ctx context.Context, mailer domainMail.Mailer, userEmail, userName string) error {
    msg := domainMail.Message{
        From:    domainMail.Address{Name: "Go-DDD Team", Email: "noreply@example.com"},
        To:      []domainMail.Address{{Name: userName, Email: userEmail}},
        Subject: "Welcome to Go-DDD",
        TextBody: "Welcome to our application!",
        HTMLBody: "<h1>Welcome to our application!</h1>",
    }
    return mailer.Send(ctx, msg)
}
```
