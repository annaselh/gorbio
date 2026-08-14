// Package mailer delivers transactional email over SMTP.
package mailer

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/annaselh/gorbio/core"
)

// Config describes the SMTP endpoint. Implicit decides between implicit TLS
// (the whole connection is wrapped, conventionally port 465) and STARTTLS (the
// session upgrades in-band, conventionally port 587).
type Config struct {
	Host        string
	Port        string
	Username    string
	Password    string
	From        string
	FromName    string
	ImplicitTLS bool
	Timeout     time.Duration
}

type SMTPMailer struct {
	config Config
}

func New(config Config) *SMTPMailer {
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	return &SMTPMailer{config: config}
}

func (m *SMTPMailer) Send(ctx context.Context, message core.Message) error {
	if err := message.Validate(); err != nil {
		return err
	}

	payload, err := m.build(message)
	if err != nil {
		return err
	}

	address := net.JoinHostPort(m.config.Host, m.config.Port)
	dialer := &net.Dialer{Timeout: m.config.Timeout}

	var conn net.Conn
	if m.config.ImplicitTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", address, &tls.Config{ServerName: m.config.Host})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return fmt.Errorf("connect to smtp server: %w", err)
	}
	defer conn.Close()

	// A context cancellation cannot interrupt net/smtp mid-command, so bound
	// the whole exchange with a deadline instead.
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(m.config.Timeout))
	}

	client, err := smtp.NewClient(conn, m.config.Host)
	if err != nil {
		return fmt.Errorf("start smtp session: %w", err)
	}
	defer client.Close()

	if !m.config.ImplicitTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: m.config.Host}); err != nil {
				return fmt.Errorf("upgrade to TLS: %w", err)
			}
		}
	}

	if m.config.Username != "" {
		// PLAIN sends the password in the clear, so only offer it once the
		// connection is encrypted.
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp authentication: %w", err)
			}
		}
	}

	if err := client.Mail(m.config.From); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	for _, recipient := range message.To {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("smtp RCPT TO %q: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err := writer.Write(payload); err != nil {
		return fmt.Errorf("write message body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finalise message: %w", err)
	}

	return client.Quit()
}

// build renders an RFC 5322 message. Bodies are base64 encoded so any UTF-8
// content is safe against the 998-octet line limit without quoted-printable's
// escaping rules.
func (m *SMTPMailer) build(message core.Message) ([]byte, error) {
	var buf strings.Builder

	from := m.config.From
	if m.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", m.config.FromName), m.config.From)
	}

	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(message.To, ", "))
	fmt.Fprintf(&buf, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", message.Subject))
	fmt.Fprintf(&buf, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")

	hasText := strings.TrimSpace(message.Text) != ""
	hasHTML := strings.TrimSpace(message.HTML) != ""

	switch {
	case hasText && hasHTML:
		boundary, err := randomBoundary()
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
		// Least-capable part first: a client that cannot render HTML shows the
		// text alternative it encountered earlier.
		writePart(&buf, boundary, "text/plain; charset=utf-8", message.Text)
		writePart(&buf, boundary, "text/html; charset=utf-8", message.HTML)
		fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	case hasHTML:
		writeSingleBody(&buf, "text/html; charset=utf-8", message.HTML)
	default:
		writeSingleBody(&buf, "text/plain; charset=utf-8", message.Text)
	}

	return []byte(buf.String()), nil
}

func writeSingleBody(buf *strings.Builder, contentType, body string) {
	fmt.Fprintf(buf, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(buf, "Content-Transfer-Encoding: base64\r\n\r\n")
	buf.WriteString(wrapBase64(body))
}

func writePart(buf *strings.Builder, boundary, contentType, body string) {
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	fmt.Fprintf(buf, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(buf, "Content-Transfer-Encoding: base64\r\n\r\n")
	buf.WriteString(wrapBase64(body))
}

// wrapBase64 folds at 76 characters, the limit RFC 2045 sets for base64 lines.
func wrapBase64(body string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(body))

	var out strings.Builder
	for len(encoded) > 76 {
		out.WriteString(encoded[:76])
		out.WriteString("\r\n")
		encoded = encoded[76:]
	}
	out.WriteString(encoded)
	out.WriteString("\r\n")
	return out.String()
}

func randomBoundary() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate mime boundary: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
