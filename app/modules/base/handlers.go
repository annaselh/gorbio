package base

import (
	"errors"
	"net/http"
	"sort"

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

func (s *AuthService) meHandler(c *gin.Context) {
	principal, _ := PrincipalFromContext(c)
	permissions := make([]string, 0, len(principal.Permissions))
	for code := range principal.Permissions {
		permissions = append(permissions, code)
	}
	sort.Strings(permissions)
	c.JSON(http.StatusOK, gin.H{"user_id": principal.UserID, "tenant_id": principal.TenantID, "permissions": permissions})
}
