// Package http assembles the Gin engine: the base middleware chain and
// route registration. Named "http" per plan/architecture/backend.md's
// internal/delivery/http package; callers that also need net/http
// should import this package under an alias (e.g. deliveryhttp) to
// avoid a naming collision.
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/your-org/your-project/backend/internal/delivery/http/handlers"
	"github.com/your-org/your-project/backend/internal/delivery/http/middleware"
)

// Handlers groups every handler the router needs to wire up routes.
// Constructed by the DI container (internal/infrastructure/container)
// and passed in here — routes.go never constructs a handler itself.
type Handlers struct {
	Health *handlers.HealthHandler
}

// NewRouter builds and returns a fully configured Gin engine: the base
// middleware chain (request_id -> logger -> cors -> recovery, in that
// order) followed by route registration. Auth/tenant-resolution
// middleware are not part of this chain yet — they land in Cycle 2 once
// there are protected routes to guard.
func NewRouter(log zerolog.Logger, h Handlers) *gin.Engine {
	engine := gin.New()

	// Base middleware chain, in the required order.
	engine.Use(
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.CORS(),
		middleware.Recovery(log),
	)

	registerRoutes(engine, h)

	return engine
}

// registerRoutes attaches every route group to the engine.
func registerRoutes(engine *gin.Engine, h Handlers) {
	// Deliberately outside /api/v1: an infra probe, not a versioned API endpoint.
	engine.GET("/health", h.Health.Health)
}
