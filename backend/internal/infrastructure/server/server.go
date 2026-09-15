// Package server owns HTTP server lifecycle: starting the listener and
// shutting it down gracefully on signal.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/your-org/your-project/backend/config"
)

// Server wraps a standard library *http.Server configured from
// config.ServerConfig, exposing a minimal Start/Shutdown lifecycle so
// callers (cmd/api/main.go) don't need to know about net/http directly.
type Server struct {
	httpServer *http.Server
}

// New constructs a Server bound to cfg.Address(), serving handler, with
// read/write timeouts taken from cfg.
func New(cfg config.ServerConfig, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Address(),
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

// Start begins serving requests and blocks until the server stops. It
// returns nil on a graceful shutdown (triggered by Shutdown) and a
// non-nil error for any other failure (e.g. the listener address is
// already in use). Intended to be run in its own goroutine, with its
// return value reported back over a channel.
func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Addr returns the address the server is configured to listen on.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// Shutdown gracefully stops the server: it stops accepting new
// connections and waits for in-flight requests to complete, up to the
// deadline carried by ctx (callers should derive ctx from
// config.ServerConfig.ShutdownTimeout).
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
