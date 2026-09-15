// Package logger provides a small zerolog wrapper so every part of the
// application (app-level bootstrap, the request logger middleware, etc.)
// constructs its structured logger the same way.
package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// New constructs a structured, JSON-output zerolog.Logger configured from
// the given level string (e.g. "debug", "info", "warn", "error"). Unknown
// or empty levels fall back to "info" so misconfiguration never disables
// logging entirely.
//
// Every part of the application (app-level logging, the request logger
// middleware, etc.) should construct its logger through this function so
// output format and level parsing stay consistent.
func New(level string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	parsedLevel, err := zerolog.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil || level == "" {
		parsedLevel = zerolog.InfoLevel
	}

	return zerolog.New(os.Stdout).
		Level(parsedLevel).
		With().
		Timestamp().
		Logger()
}

// Nop returns a logger that discards all output. Useful for tests and any
// code path that needs a logger.Logger value but doesn't want output.
func Nop() zerolog.Logger {
	return zerolog.Nop()
}
