// Package http assembles the Gin engine, base middleware, and route registration.
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	// Blank import registers the generated Swagger spec with gin-swagger.
	// Regenerate via `make swagger` after changing @-annotations.
	_ "github.com/skryfon/employee360/backend/docs"
	"github.com/skryfon/employee360/backend/internal/delivery/http/handlers"
	"github.com/skryfon/employee360/backend/internal/delivery/http/middleware"
	"github.com/skryfon/employee360/backend/shared"
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

	// Swagger UI: interactive API docs, always available (no environment gating).
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
