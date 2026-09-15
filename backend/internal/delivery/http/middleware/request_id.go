// Package middleware holds Gin middleware: request_id, logger, cors,
// recovery (the base chain). Auth and tenant-resolution middleware are
// added in Cycle 2 once there's something to protect.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/your-project/backend/shared"
)

// ContextKeyRequestID is the Gin context key the request id is stored
// under, for handlers/other middleware running later in the chain.
const ContextKeyRequestID = "request_id"

// RequestID assigns a unique id to every request: it reuses an inbound
// X-Request-ID header when the client supplied one (useful for tracing
// across services), otherwise it generates a new UUID. The id is stored
// on the Gin context and echoed back on the response header so clients
// and downstream services can correlate logs.
//
// Must run first in the middleware chain so every later middleware
// (logger, recovery) and every handler can rely on it being present.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(shared.RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(ContextKeyRequestID, requestID)
		c.Writer.Header().Set(shared.RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID returns the request id stored on the Gin context by
// RequestID, or "" if it hasn't run (or ran in a different chain).
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ContextKeyRequestID); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
