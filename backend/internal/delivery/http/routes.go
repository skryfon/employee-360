// Package http assembles the Gin engine, base middleware, and route registration.
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/skryfon/employee360/backend/config"
	"github.com/skryfon/employee360/backend/internal/delivery/http/middleware"
	"github.com/skryfon/employee360/backend/internal/domain/entity"
	"github.com/skryfon/employee360/backend/internal/infrastructure/container"
	"github.com/skryfon/employee360/backend/shared"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter builds and returns a fully configured Gin engine with middleware and routes.
func SetupRouter(cfg *config.Config, log zerolog.Logger, ctr *container.AppContainer) *gin.Engine {
	engine := gin.New()

	// Trust no proxy unless configured: otherwise any client could spoof
	// X-Forwarded-For to dodge the per-IP rate limiter (gin trusts all by default).
	if err := engine.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		log.Fatal().Err(err).Msg("invalid server.trusted_proxies")
	}

	// Base middleware chain: request_id -> logger -> cors -> recovery.
	engine.Use(
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.CORS(cfg.CORS.AllowedOrigins),
		middleware.Recovery(log),
	)

	// Global rate limiting (after recovery so 429s still carry request-id/CORS headers).
	if cfg.RateLimit.Enabled {
		limiter, _ := middleware.RateLimit(middleware.RateLimitConfig{
			RequestsPerSecond: cfg.RateLimit.RequestsPerSecond,
			Burst:             cfg.RateLimit.Burst,
			CleanupInterval:   cfg.RateLimit.CleanupInterval,
			IdleTTL:           cfg.RateLimit.IdleTTL,
			ExemptPaths:       []string{"/health", "/healthz", shared.APIVersionPrefix + "/health"},
		})
		engine.Use(limiter)
	}

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

		if c.Auth != nil && c.Auth.InvitationHandler != nil {
			ih := c.Auth.InvitationHandler

			// Unauthenticated: the invitee proves possession of the emailed token.
			v1.POST("/invitations/accept", ih.Accept)

			// Admin-only management routes: auth -> tenant -> role check.
			invGroup := v1.Group("/users/invitations",
				middleware.Auth(c.Auth.TokenService),
				middleware.Tenant(),
				middleware.RequireRole(entity.RoleAdmin, entity.RoleSuperAdmin),
			)
			{
				invGroup.POST("", ih.Invite)
				invGroup.GET("", ih.List)
				invGroup.POST("/:id/resend", ih.Resend)
				invGroup.DELETE("/:id", ih.Revoke)
			}
		}
	}

	// Swagger UI: interactive API docs, always available (no environment gating).
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
