// Package http assembles the Gin engine, base middleware, and route registration.
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/delivery/http/middleware"
	"github.com/skryfon/employee360/backend/internal/infrastructure/container"
	"github.com/skryfon/employee360/backend/shared"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter builds and returns a fully configured Gin engine with middleware and routes.
func SetupRouter(cfg *config.Config, log zerolog.Logger, ctr *container.AppContainer) *gin.Engine {
	engine := gin.New()

	// Base middleware chain: request_id -> logger -> cors -> recovery.
	engine.Use(
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.CORS(cfg.CORS.AllowedOrigins),
		middleware.Recovery(log),
	)

	registerRoutes(engine, ctr)

	return engine
}

// registerRoutes attaches all route groups to the engine.
func registerRoutes(engine *gin.Engine, c *container.Container) {
	// Unversioned operational health checks for load balancers / Kubernetes / monitoring.
	if c.Health != nil && c.Health.Handler != nil {
		engine.GET("/health", c.Health.Handler.Health)
		engine.GET("/healthz", c.Health.Handler.Health)
	}

	v1 := engine.Group(shared.APIVersionPrefix)
	{
		// Versioned health check for client SDKs / smoke tests.
		if c.Health != nil && c.Health.Handler != nil {
			v1.GET("/health", c.Health.Handler.Health)
		}

		if c.Auth != nil && c.Auth.Handler != nil {
			authGroup := v1.Group("/auth")
			{
				authGroup.POST("/login", c.Auth.Handler.Login)
				authGroup.POST("/refresh", c.Auth.Handler.Refresh)
				authGroup.POST("/forgot-password", c.Auth.Handler.ForgotPassword)
				authGroup.POST("/reset-password", c.Auth.Handler.ResetPassword)

				authProtected := authGroup.Group("", middleware.Auth(c.Auth.TokenService), middleware.Tenant())
				{
					authProtected.POST("/logout", c.Auth.Handler.Logout)
				}
			}
		}
	}

	// Swagger UI: interactive API docs, always available (no environment gating).
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
