package mailer

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/annaselh/gorbio/core"
)

func testMailer() *SMTPMailer {
	return New(Config{
		Host: "smtp.example.com", Port: "587",
		From: "no-reply@example.com", FromName: "Orbio",
	})
}

func decodeBase64Block(t *testing.T, block string) string {
	t.Helper()
	cleaned := strings.ReplaceAll(strings.TrimSpace(block), "\r\n", "")
	decoded, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		t.Fatalf("body should be valid base64: %v", err)
	}
	return string(decoded)
}

func TestBuildSetsRequiredHeaders(t *testing.T) {
	raw, err := testMailer().build(core.Message{
		To: []string{"user@example.com"}, Subject: "Reset your password", Text: "hello",
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	message := string(raw)
	for _, header := range []string{"From: ", "To: user@example.com", "Subject: ", "Date: ", "MIME-Version: 1.0"} {
		if !strings.Contains(message, header) {
			t.Fatalf("message is missing %q:\n%s", header, message)
		}
	}
}

func TestBuildUsesMultipartWhenBothBodiesPresent(t *testing.T) {
	raw, err := testMailer().build(core.Message{
		To: []string{"user@example.com"}, Subject: "Hi",
		Text: "plain version", HTML: "<p>rich version</p>",
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	message := string(raw)
	if !strings.Contains(message, "multipart/alternative") {
		t.Fatal("two bodies should produce multipart/alternative")
	}
	if !strings.Contains(message, "text/plain") || !strings.Contains(message, "text/html") {
		t.Fatal("both alternatives should be present")
	}

	// The plain part must come first so a text-only client shows it.
	if strings.Index(message, "text/plain") > strings.Index(message, "text/html") {
		t.Fatal("text/plain must precede text/html in multipart/alternative")
	}
}

func TestBuildSingleBodyIsDecodable(t *testing.T) {
	body := "Open this link to choose a new password"
	raw, err := testMailer().build(core.Message{
		To: []string{"user@example.com"}, Subject: "Hi", Text: body,
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	message := string(raw)
	_, encoded, found := strings.Cut(message, "\r\n\r\n")
	if !found {
		t.Fatal("message should separate headers from body with a blank line")
	}
	if decoded := decodeBase64Block(t, encoded); decoded != body {
		t.Fatalf("decoded body mismatch: got %q, want %q", decoded, body)
	}
}

func TestBuildEncodesNonASCIISubject(t *testing.T) {
	raw, err := testMailer().build(core.Message{
		To: []string{"user@example.com"}, Subject: "Setel ulang kata sandi — Orbio", Text: "hi",
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	message := string(raw)
	subjectLine := ""
	for _, line := range strings.Split(message, "\r\n") {
		if strings.HasPrefix(line, "Subject: ") {
			subjectLine = line
			break
		}
	}
	if subjectLine == "" {
		t.Fatal("no Subject header found")
	}
	// A raw em dash on the wire is not valid in a header; it must be encoded.
	if strings.Contains(subjectLine, "—") {
		t.Fatalf("non-ASCII subject must be RFC 2047 encoded, got %q", subjectLine)
	}
	if !strings.Contains(subjectLine, "=?utf-8?") {
		t.Fatalf("expected an encoded-word subject, got %q", subjectLine)
	}
}

// Base64 lines longer than 76 characters violate RFC 2045 and some servers
// reject or mangle them.
func TestWrapBase64FoldsLongLines(t *testing.T) {
	wrapped := wrapBase64(strings.Repeat("abcdefghij", 40))

	for _, line := range strings.Split(strings.TrimSpace(wrapped), "\r\n") {
		if len(line) > 76 {
			t.Fatalf("line exceeds the 76 character limit: %d chars", len(line))
		}
	}
}

func TestWrapBase64RoundTrips(t *testing.T) {
	original := strings.Repeat("Orbio reset link. ", 30)
	decoded, err := base64.StdEncoding.DecodeString(
		strings.ReplaceAll(strings.TrimSpace(wrapBase64(original)), "\r\n", ""),
	)
	if err != nil {
		t.Fatalf("wrapped output should decode: %v", err)
	}
	if string(decoded) != original {
		t.Fatal("wrapping must not alter the payload")
	}
}

func TestMessageValidateRejectsIncompleteMessages(t *testing.T) {
	cases := map[string]core.Message{
		"no recipient":    {Subject: "Hi", Text: "body"},
		"blank recipient": {To: []string{"  "}, Subject: "Hi", Text: "body"},
		"no subject":      {To: []string{"user@example.com"}, Text: "body"},
		"no body":         {To: []string{"user@example.com"}, Subject: "Hi"},
	}

	for name, message := range cases {
		if err := message.Validate(); err == nil {
			t.Fatalf("%s: expected validation to fail", name)
		}
	}
}

func TestMessageValidateAcceptsHTMLOnly(t *testing.T) {
	message := core.Message{To: []string{"user@example.com"}, Subject: "Hi", HTML: "<p>body</p>"}
	if err := message.Validate(); err != nil {
		t.Fatalf("an HTML-only message is valid, got %v", err)
	}
}

func TestRandomBoundaryIsUnique(t *testing.T) {
	first, err := randomBoundary()
	if err != nil {
		t.Fatalf("boundary generation failed: %v", err)
	}
	second, err := randomBoundary()
	if err != nil {
		t.Fatalf("boundary generation failed: %v", err)
	}
	if first == second {
		t.Fatal("boundaries must not repeat between messages")
	}
}
