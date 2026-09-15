// Package http assembles the Gin engine, base middleware, and route registration.
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/your-org/your-project/backend/internal/delivery/http/handlers"
	"github.com/your-org/your-project/backend/internal/delivery/http/middleware"
	"github.com/your-org/your-project/backend/shared"
)

// Handlers groups all HTTP handlers required by the router.
type Handlers struct {
	Health *handlers.HealthHandler
}

// NewRouter builds and returns a fully configured Gin engine with middleware and routes.
func NewRouter(log zerolog.Logger, allowedOrigins []string, h Handlers) *gin.Engine {
	engine := gin.New()

	// Base middleware chain: request_id -> logger -> cors -> recovery.
	engine.Use(
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.CORS(allowedOrigins),
		middleware.Recovery(log),
	)

	registerRoutes(engine, h)

	return engine
}

// registerRoutes attaches all route groups to the engine.
func registerRoutes(engine *gin.Engine, h Handlers) {
	// Unversioned operational health checks for load balancers / Kubernetes / monitoring.
	engine.GET("/health", h.Health.Health)
	engine.GET("/healthz", h.Health.Health)

	v1 := engine.Group(shared.APIVersionPrefix)
	{
		// Versioned health check for client SDKs / smoke tests.
		v1.GET("/health", h.Health.Health)
	}
}
