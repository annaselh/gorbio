package base

import (
	"context"
	"strings"
	"testing"

	"github.com/annaselh/gorbio/core"
)

// recordingMailer captures what would have been sent.
type recordingMailer struct {
	sent []core.Message
}

func (m *recordingMailer) Send(_ context.Context, message core.Message) error {
	m.sent = append(m.sent, message)
	return nil
}

func mailerFixture(baseURL string) (*AuthService, *recordingMailer) {
	recorder := &recordingMailer{}
	return &AuthService{baseURL: baseURL, mailer: recorder}, recorder
}

func TestBuildLinkJoinsBaseAndRoute(t *testing.T) {
	service, _ := mailerFixture("https://erp.example.com")

	got := service.buildLink("/reset-password", "abc123")
	want := "https://erp.example.com/reset-password?token=abc123"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// A trailing slash in APP_BASE_URL must not produce a double slash in the link.
func TestBuildLinkNormalisesTrailingSlash(t *testing.T) {
	service, _ := mailerFixture("https://erp.example.com/")

	if got := service.buildLink("/reset-password", "abc"); strings.Contains(got, "com//") {
		t.Fatalf("link contains a doubled slash: %q", got)
	}
}

// Session tokens are base64url, which can contain characters that change
// meaning in a query string if not escaped.
func TestBuildLinkEscapesToken(t *testing.T) {
	service, _ := mailerFixture("https://erp.example.com")

	got := service.buildLink("/reset-password", "a+b/c=d&e")
	if strings.Contains(got, "&e") {
		t.Fatalf("token must be query-escaped, got %q", got)
	}
	if strings.Contains(got, "a+b/c=d") {
		t.Fatalf("token must be query-escaped, got %q", got)
	}
}

func TestPasswordResetEmailCarriesLinkAndBothBodies(t *testing.T) {
	service, recorder := mailerFixture("https://erp.example.com")
	user := User{Email: "user@example.com", DisplayName: "Ada Lovelace"}

	if err := service.sendPasswordResetEmail(context.Background(), user, "tok"); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if len(recorder.sent) != 1 {
		t.Fatalf("expected one message, got %d", len(recorder.sent))
	}
	message := recorder.sent[0]

	if message.To[0] != user.Email {
		t.Fatalf("wrong recipient: %v", message.To)
	}
	if message.Text == "" || message.HTML == "" {
		t.Fatal("both a text and an HTML body should be offered")
	}
	link := "https://erp.example.com/reset-password?token=tok"
	if !strings.Contains(message.Text, link) || !strings.Contains(message.HTML, link) {
		t.Fatal("both bodies must contain the reset link")
	}
	if !strings.Contains(message.Text, "Ada Lovelace") {
		t.Fatal("the greeting should use the display name")
	}
	if err := message.Validate(); err != nil {
		t.Fatalf("generated message should be valid: %v", err)
	}
}

// A display name is user-controlled input rendered into an HTML body.
func TestPasswordResetEmailEscapesDisplayName(t *testing.T) {
	service, recorder := mailerFixture("https://erp.example.com")
	user := User{Email: "user@example.com", DisplayName: `<script>alert(1)</script>`}

	if err := service.sendPasswordResetEmail(context.Background(), user, "tok"); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if strings.Contains(recorder.sent[0].HTML, "<script>") {
		t.Fatalf("display name must be HTML-escaped:\n%s", recorder.sent[0].HTML)
	}
}

func TestVerificationEmailCarriesVerifyRoute(t *testing.T) {
	service, recorder := mailerFixture("https://erp.example.com")
	user := User{Email: "user@example.com", DisplayName: "Ada"}

	if err := service.sendEmailVerification(context.Background(), user, "tok"); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	message := recorder.sent[0]
	if !strings.Contains(message.Text, "/verify-email?token=tok") {
		t.Fatalf("verification email should link to the verify route:\n%s", message.Text)
	}
	if strings.Contains(message.Text, "/reset-password") {
		t.Fatal("verification email must not carry a password reset link")
	}
}

func TestDisplayNameFallsBackToEmail(t *testing.T) {
	if got := displayNameOrEmail(User{Email: "user@example.com", DisplayName: "  "}); got != "user@example.com" {
		t.Fatalf("a blank display name should fall back to the address, got %q", got)
	}
	if got := displayNameOrEmail(User{Email: "user@example.com", DisplayName: "Ada"}); got != "Ada" {
		t.Fatalf("a present display name should win, got %q", got)
	}
}

// NewAuthService must never leave the mailer nil: the recovery flows call it
// unconditionally and would panic.
func TestNewAuthServiceDefaultsMailer(t *testing.T) {
	service := NewAuthService(nil, core.DefaultSettings(), nil)
	if service.mailer == nil {
		t.Fatal("mailer should default rather than stay nil")
	}
}
