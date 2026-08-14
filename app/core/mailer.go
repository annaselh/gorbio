package core

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// Message is a single outbound email. At least one body must be set; when both
// are present the transport sends multipart/alternative.
type Message struct {
	To      []string
	Subject string
	Text    string
	HTML    string
}

func (m Message) Validate() error {
	if len(m.To) == 0 {
		return fmt.Errorf("message requires at least one recipient")
	}
	for _, address := range m.To {
		if strings.TrimSpace(address) == "" {
			return fmt.Errorf("message recipient must not be blank")
		}
	}
	if strings.TrimSpace(m.Subject) == "" {
		return fmt.Errorf("message requires a subject")
	}
	if strings.TrimSpace(m.Text) == "" && strings.TrimSpace(m.HTML) == "" {
		return fmt.Errorf("message requires a text or HTML body")
	}
	return nil
}

// Mailer sends transactional email. Modules depend on this interface rather
// than on an SMTP client so a test can substitute a recorder and a development
// run can substitute the logging transport below.
type Mailer interface {
	Send(ctx context.Context, message Message) error
}

// LogMailer writes messages to the log instead of delivering them. It exists so
// a developer without an SMTP server still gets the reset link - printed to the
// console - rather than a silent dead end. Wiring rejects it in production.
type LogMailer struct{}

func (LogMailer) Send(_ context.Context, message Message) error {
	if err := message.Validate(); err != nil {
		return err
	}
	body := message.Text
	if body == "" {
		body = message.HTML
	}
	slog.Warn("email not delivered: no SMTP transport configured",
		"to", strings.Join(message.To, ", "),
		"subject", message.Subject,
		"body", body,
	)
	return nil
}
