package base

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	TenantSlug string `json:"tenant_slug"`
}

func (s *AuthService) loginHandler(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login request"})
		return
	}
	result, err := s.Login(c.Request.Context(), LoginInput{
		Email: request.Email, Password: request.Password, TenantSlug: request.TenantSlug,
		IPAddress: c.ClientIP(), UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		case errors.Is(err, ErrTooManyAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, try again later"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		}
		return
	}
	s.setSessionCookie(c, result.Token, result.ExpiresAt)
	c.JSON(http.StatusOK, gin.H{"user_id": result.Principal.UserID, "tenant_id": result.Principal.TenantID, "expires_at": result.ExpiresAt})
}

func (s *AuthService) logoutHandler(c *gin.Context) {
	principal, _ := PrincipalFromContext(c)
	if err := s.Logout(c.Request.Context(), principal.SessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
		return
	}
	s.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

func (s *AuthService) changePasswordHandler(c *gin.Context) {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var request changePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current and new password are required"})
		return
	}

	switch err := s.ChangePassword(c.Request.Context(), principal, request.CurrentPassword, request.NewPassword); {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
	case errors.Is(err, ErrSamePassword):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, ErrWeakPassword):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		slog.Error("password change failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not change password"})
	}
}

// forgotPasswordHandler always answers 202, whether or not the address exists.
// Distinguishing the two would turn this endpoint into an account oracle.
func (s *AuthService) forgotPasswordHandler(c *gin.Context) {
	var request forgotPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a valid email address is required"})
		return
	}

	err := s.RequestPasswordReset(c.Request.Context(), request.Email, c.ClientIP())
	if errors.Is(err, ErrTooManyAttempts) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many reset requests, try again later"})
		return
	}
	if err != nil {
		// Mail transport failures are logged upstream; still answer neutrally
		// so a delivery outage does not leak which addresses are registered.
		slog.Error("password reset request failed", "error", err)
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "If that address has an account, a reset link is on its way.",
	})
}

func (s *AuthService) resetPasswordHandler(c *gin.Context) {
	var request resetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and password are required"})
		return
	}

	switch err := s.ResetPassword(c.Request.Context(), request.Token, request.Password); {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, ErrInvalidToken):
		c.JSON(http.StatusBadRequest, gin.H{"error": "this reset link is invalid or has expired"})
	case errors.Is(err, ErrWeakPassword):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		slog.Error("password reset failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not reset password"})
	}
}

func (s *AuthService) verifyEmailHandler(c *gin.Context) {
	var request verifyEmailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	switch err := s.VerifyEmail(c.Request.Context(), request.Token); {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, ErrInvalidToken):
		c.JSON(http.StatusBadRequest, gin.H{"error": "this verification link is invalid or has expired"})
	default:
		slog.Error("email verification failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not verify email"})
	}
}

func (s *AuthService) resendVerificationHandler(c *gin.Context) {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	if err := s.RequestEmailVerification(c.Request.Context(), principal); err != nil {
		slog.Error("email verification request failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not send verification email"})
		return
	}
	c.Status(http.StatusAccepted)
}

func (s *AuthService) meHandler(c *gin.Context) {
	principal, ok := PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	profile, err := s.Profile(c.Request.Context(), principal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":      profile.UserID,
		"email":        profile.Email,
		"display_name": profile.DisplayName,
		"tenant_id":    profile.TenantID,
		"tenant_slug":  profile.TenantSlug,
		"tenant_name":  profile.TenantName,
		"permissions":  profile.Permissions,
	})
}
