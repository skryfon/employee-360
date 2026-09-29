package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skryfon/employee360/backend/internal/ctx"
	"github.com/skryfon/employee360/backend/internal/delivery/http/response"
	domainerrors "github.com/skryfon/employee360/backend/internal/domain/errors"
)

// Tenant enforces strict multi-tenant isolation by resolving tenant_id
// exclusively from validated JWT claims (already placed in context by Auth middleware).
//
// Invariant 1: Handlers and repositories must never trust client-supplied tenant IDs.
// Any client-supplied tenant identifier (such as X-Tenant-ID headers, query parameters,
// or JSON payload fields) is strictly ignored. The context tenant_id is guaranteed to
// originate solely from the validated access token claims.
//
// If tenant_id is missing or is not a valid non-nil UUID, Tenant aborts with 401 Unauthorized.
func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDStr := ""

		// Check Gin context (populated by Auth middleware)
		if val, ok := c.Get(ContextKeyTenantID); ok {
			if s, ok := val.(string); ok && s != "" {
				tenantIDStr = s
			}
		}

		// Check Go request context if not found in Gin context
		if tenantIDStr == "" {
			if s, ok := ctx.TenantIDFromContext(c.Request.Context()); ok && s != "" {
				tenantIDStr = s
			}
		}

		// Or extract from claims if available
		if tenantIDStr == "" {
			if claims, ok := GetClaims(c); ok && claims != nil && claims.TenantID != uuid.Nil {
				tenantIDStr = claims.TenantID.String()
			}
		}

		if tenantIDStr == "" {
			response.Unauthorized(c, "tenant context required")
			c.Abort()
			return
		}

		tenantUUID, err := uuid.Parse(tenantIDStr)
		if err != nil || tenantUUID == uuid.Nil {
			response.Unauthorized(c, "invalid tenant in token claims")
			c.Abort()
			return
		}

		// Ensure tenant_id is set in request context exclusively from verified JWT claims
		reqCtx := ctx.WithTenantID(c.Request.Context(), tenantUUID.String())
		c.Request = c.Request.WithContext(reqCtx)
		c.Set(ContextKeyTenantID, tenantUUID.String())

		c.Next()
	}
}

// GetTenantID extracts and parses the tenant UUID from the request context or Gin context.
// Returns an error if tenant_id is missing or not a valid UUID.
func GetTenantID(c *gin.Context) (uuid.UUID, error) {
	if val, ok := c.Get(ContextKeyTenantID); ok {
		if strVal, ok := val.(string); ok && strVal != "" {
			return uuid.Parse(strVal)
		}
		if uid, ok := val.(uuid.UUID); ok && uid != uuid.Nil {
			return uid, nil
		}
	}

	if strVal, ok := ctx.TenantIDFromContext(c.Request.Context()); ok && strVal != "" {
		return uuid.Parse(strVal)
	}

	return uuid.Nil, domainerrors.ErrUnauthorized
}
