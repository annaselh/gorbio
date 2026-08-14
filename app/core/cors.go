package core

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const corsPreflightMaxAge = "600"

// CORS allows credentialed cross-origin requests from an explicit allowlist.
//
// The session lives in a cookie, so every browser call is credentialed, and the
// spec forbids pairing Access-Control-Allow-Credentials with a "*" origin. Each
// allowed origin is therefore echoed back individually, and Vary: Origin stops a
// shared cache from serving one origin's response to another.
//
// An empty allowlist disables CORS entirely, which is the correct posture when
// the SPA is served same-origin behind a reverse proxy.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[normalizeOrigin(origin)] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := normalizeOrigin(c.GetHeader("Origin"))
		if origin == "" {
			c.Next()
			return
		}

		c.Writer.Header().Add("Vary", "Origin")

		if _, ok := allowed[origin]; !ok {
			// Omit the CORS headers so the browser blocks the response. A
			// preflight is rejected outright; a simple request still runs but
			// its result is unreadable to the calling page.
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}

		header := c.Writer.Header()
		header.Set("Access-Control-Allow-Origin", origin)
		header.Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			requested := c.GetHeader("Access-Control-Request-Headers")
			if strings.TrimSpace(requested) == "" {
				requested = "Content-Type"
			}
			header.Set("Access-Control-Allow-Headers", requested)
			header.Set("Access-Control-Max-Age", corsPreflightMaxAge)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func normalizeOrigin(origin string) string {
	return strings.TrimRight(strings.TrimSpace(origin), "/")
}
