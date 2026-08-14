package base

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const principalContextKey = "base.principal"

func (s *AuthService) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(SessionCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		principal, err := s.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.Set(principalContextKey, principal)
		c.Next()
	}
}

func RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := PrincipalFromContext(c)
		if !ok || !principal.HasPermission(code) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}
		c.Next()
	}
}

func PrincipalFromContext(c *gin.Context) (Principal, bool) {
	value, ok := c.Get(principalContextKey)
	if !ok {
		return Principal{}, false
	}
	principal, ok := value.(Principal)
	return principal, ok
}

// sameSiteMode pairs with Secure: a cross-site SPA needs SameSite=None, which
// browsers only honour on a Secure cookie. Over plain HTTP in development that
// combination is rejected outright, so fall back to Lax there.
func (s *AuthService) sameSiteMode() http.SameSite {
	if s.cookieSecure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func (s *AuthService) setSessionCookie(c *gin.Context, token string, expiresAt time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: SessionCookieName, Value: token, Path: "/", Expires: expiresAt,
		HttpOnly: true, Secure: s.cookieSecure, SameSite: s.sameSiteMode(),
	})
}

func (s *AuthService) clearSessionCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: SessionCookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: s.cookieSecure, SameSite: s.sameSiteMode(),
	})
}
