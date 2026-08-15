package mail

import "context"

type Address struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Message struct {
	From     Address   `json:"from"`
	To       []Address `json:"to"`
	Subject  string    `json:"subject"`
	TextBody string    `json:"text_body"`
	HTMLBody string    `json:"html_body"`
}

type Mailer interface {
	Send(ctx context.Context, msg Message) error
}
