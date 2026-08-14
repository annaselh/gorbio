package base

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	passwordResetTTL     = time.Hour
	emailVerificationTTL = 24 * time.Hour

	// Password reset is an unauthenticated endpoint that sends mail, so it is
	// rate limited per IP on top of the generic login limiter.
	resetRequestLimit  = 5
	resetRequestWindow = time.Hour
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrWeakPassword = errors.New("password does not meet requirements")
)

// RequestPasswordReset issues a reset token and mails it.
//
// It reports success even when the address is unknown: a caller must not be
// able to tell registered addresses from unregistered ones by watching the
// response. The only signal of failure is the rate limiter.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email, ipAddress string) error {
	if ipAddress != "" && !s.resetLimiter.Allow(ipAddress) {
		return ErrTooManyAttempts
	}

	normalized := strings.ToLower(strings.TrimSpace(email))
	var user User
	if err := s.db.WithContext(ctx).Where("email = ?", normalized).First(&user).Error; err != nil {
		// Unknown address: stop here, but answer as though mail was sent.
		return nil
	}
	if user.Status == UserStatusDisabled {
		return nil
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Supersede outstanding tokens so only the newest link works; an
		// attacker who obtained an earlier one loses it the moment the real
		// owner asks again.
		if err := tx.Model(&PasswordResetToken{}).
			Where("user_id = ? AND used_at IS NULL", user.ID).
			Update("used_at", now).Error; err != nil {
			return fmt.Errorf("supersede reset tokens: %w", err)
		}
		return tx.Create(&PasswordResetToken{
			ID: uuid.New(), UserID: user.ID, TokenHash: tokenHash,
			ExpiresAt: now.Add(passwordResetTTL),
		}).Error
	})
	if err != nil {
		return fmt.Errorf("create reset token: %w", err)
	}

	s.recordAudit(ctx, &user.ID, nil, "auth.password_reset_requested", "user", user.ID.String(), ipAddress, "")

	return s.sendPasswordResetEmail(ctx, user, token)
}

// ResetPassword consumes a reset token and replaces the password.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if strings.TrimSpace(token) == "" {
		return ErrInvalidToken
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		// hashPassword enforces the length policy; surface it as a validation
		// failure rather than an internal error.
		return fmt.Errorf("%w: %s", ErrWeakPassword, err)
	}

	hash := hashToken(token)
	now := time.Now().UTC()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reset PasswordResetToken
		if err := tx.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, now).
			First(&reset).Error; err != nil {
			return ErrInvalidToken
		}

		// Mark used before writing the password so a concurrent replay of the
		// same token loses the race rather than resetting twice.
		if err := tx.Model(&reset).Update("used_at", now).Error; err != nil {
			return fmt.Errorf("consume reset token: %w", err)
		}

		if err := tx.Model(&User{}).Where("id = ?", reset.UserID).Updates(map[string]any{
			"password_hash":         passwordHash,
			"password_changed_at":   now,
			"failed_login_attempts": 0,
			"locked_until":          nil,
		}).Error; err != nil {
			return fmt.Errorf("update password: %w", err)
		}

		// A password change must end every existing session: the point of the
		// reset is usually that someone else holds the old credentials.
		if err := tx.Model(&Session{}).
			Where("user_id = ? AND revoked_at IS NULL", reset.UserID).
			Update("revoked_at", now).Error; err != nil {
			return fmt.Errorf("revoke sessions: %w", err)
		}

		s.recordAuditTx(ctx, tx, &reset.UserID, nil, "auth.password_reset_completed", "user", reset.UserID.String())
		return nil
	})
}

// RequestEmailVerification issues a verification token for a signed-in user.
func (s *AuthService) RequestEmailVerification(ctx context.Context, principal Principal) error {
	var user User
	if err := s.db.WithContext(ctx).Where("id = ?", principal.UserID).First(&user).Error; err != nil {
		return fmt.Errorf("load user: %w", err)
	}
	if user.EmailVerifiedAt != nil {
		// Already verified; nothing to send and no error to report.
		return nil
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&EmailVerificationToken{}).
			Where("user_id = ? AND used_at IS NULL", user.ID).
			Update("used_at", now).Error; err != nil {
			return fmt.Errorf("supersede verification tokens: %w", err)
		}
		return tx.Create(&EmailVerificationToken{
			ID: uuid.New(), UserID: user.ID, TokenHash: tokenHash,
			ExpiresAt: now.Add(emailVerificationTTL),
		}).Error
	})
	if err != nil {
		return fmt.Errorf("create verification token: %w", err)
	}

	return s.sendEmailVerification(ctx, user, token)
}

// VerifyEmail consumes a verification token and marks the address confirmed.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrInvalidToken
	}

	hash := hashToken(token)
	now := time.Now().UTC()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var verification EmailVerificationToken
		if err := tx.Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, now).
			First(&verification).Error; err != nil {
			return ErrInvalidToken
		}

		if err := tx.Model(&verification).Update("used_at", now).Error; err != nil {
			return fmt.Errorf("consume verification token: %w", err)
		}

		updates := map[string]any{"email_verified_at": now}
		// A pending account becomes usable once its address is confirmed; an
		// account disabled by an administrator stays disabled.
		if err := tx.Model(&User{}).
			Where("id = ? AND status = ?", verification.UserID, UserStatusPending).
			Updates(map[string]any{"status": UserStatusActive}).Error; err != nil {
			return fmt.Errorf("activate user: %w", err)
		}
		if err := tx.Model(&User{}).Where("id = ?", verification.UserID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("mark email verified: %w", err)
		}

		s.recordAuditTx(ctx, tx, &verification.UserID, nil, "auth.email_verified", "user", verification.UserID.String())
		return nil
	})
}
