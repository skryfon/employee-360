package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/your-org/your-project/backend/internal/delivery/http/response"
)

// Recovery returns a panic-recovery middleware. It must be the outermost
// middleware in effect around handler execution so a panic anywhere in a
// later middleware or handler is caught, logged with a stack trace, and
// turned into a standardized 500 error envelope instead of crashing the
// process or leaking a raw stack trace to the client.
func Recovery(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Str("request_id", GetRequestID(c)).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Interface("panic", r).
					Bytes("stack", debug.Stack()).
					Msg("panic recovered")

				if !c.Writer.Written() {
					response.Internal(c, "an unexpected error occurred")
				}
				c.Abort()
			}
		}()

		c.Next()
	}
}
