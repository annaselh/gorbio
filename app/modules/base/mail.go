package base

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/annaselh/gorbio/core"
)

// buildLink joins the public web origin with a client route and token. The
// token goes in the query string rather than the path so it never lands in a
// path-based access log pattern by accident.
func (s *AuthService) buildLink(routePath, token string) string {
	base := strings.TrimRight(s.baseURL, "/")
	return fmt.Sprintf("%s%s?token=%s", base, routePath, url.QueryEscape(token))
}

func (s *AuthService) sendPasswordResetEmail(ctx context.Context, user User, token string) error {
	link := s.buildLink("/reset-password", token)
	name := displayNameOrEmail(user)

	text := fmt.Sprintf(`Hi %s,

We received a request to reset the password for your Orbio account.

Open this link to choose a new password:
%s

The link expires in 1 hour and can be used once.

If you did not request this, you can ignore this email - your password
stays unchanged.`, name, link)

	htmlBody := fmt.Sprintf(`<p>Hi %s,</p>
<p>We received a request to reset the password for your Orbio account.</p>
<p><a href="%s">Choose a new password</a></p>
<p>The link expires in 1 hour and can be used once.</p>
<p>If you did not request this, you can ignore this email &mdash; your password stays unchanged.</p>`,
		html.EscapeString(name), html.EscapeString(link))

	return s.mailer.Send(ctx, core.Message{
		To:      []string{user.Email},
		Subject: "Reset your Orbio password",
		Text:    text,
		HTML:    htmlBody,
	})
}

func (s *AuthService) sendEmailVerification(ctx context.Context, user User, token string) error {
	link := s.buildLink("/verify-email", token)
	name := displayNameOrEmail(user)

	text := fmt.Sprintf(`Hi %s,

Confirm this address to finish setting up your Orbio account:
%s

The link expires in 24 hours and can be used once.`, name, link)

	htmlBody := fmt.Sprintf(`<p>Hi %s,</p>
<p>Confirm this address to finish setting up your Orbio account.</p>
<p><a href="%s">Verify email address</a></p>
<p>The link expires in 24 hours and can be used once.</p>`,
		html.EscapeString(name), html.EscapeString(link))

	return s.mailer.Send(ctx, core.Message{
		To:      []string{user.Email},
		Subject: "Verify your Orbio email address",
		Text:    text,
		HTML:    htmlBody,
	})
}

func displayNameOrEmail(user User) string {
	if name := strings.TrimSpace(user.DisplayName); name != "" {
		return name
	}
	return user.Email
}
