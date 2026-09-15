package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS returns a CORS middleware suitable for a headless API consumed by
// multiple, independently-deployed clients (web, mobile, third-party).
// It reflects the request's Origin header (rather than a hardcoded
// wildcard) so credentialed requests (bearer tokens, and in future any
// cookie-based flows) keep working, while still allowing any origin to
// call the API — access control is enforced by auth/tenant middleware,
// not by CORS.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, X-Tenant-ID")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
