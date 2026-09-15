package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Logger returns a structured request logging middleware built on top of
// the given zerolog.Logger. It logs one line per request (method, path,
// status, latency, client ip, request id) after the handler chain
// completes, so it must run after RequestID (to have an id to log) and
// before Recovery would be a mistake — Recovery must wrap it so a panic
// downstream is still reported with a 500 status in this log line.
func Logger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("request_id", GetRequestID(c)).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Int("body_size", c.Writer.Size()).
			Msg("http_request")
	}
}
