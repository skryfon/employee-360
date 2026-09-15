// Package ctx defines the context keys and typed accessors used to pass
// request-scoped values (tenant id, user id, roles) down through the
// usecase and infrastructure layers without those layers depending on
// Gin or *gin.Context directly.
//
// Nothing populates these yet — the auth and tenant-resolution
// middleware that call WithTenantID/WithUserID/WithRoles land in
// Cycle 2. This package only defines the contract in advance so that
// layer code introduced later has a stable, already-reviewed place to
// read/write these values, per plan/architecture/backend.md.
package ctx

import "context"

type contextKey string

const (
	tenantIDKey contextKey = "employee360:tenant_id"
	userIDKey   contextKey = "employee360:user_id"
	rolesKey    contextKey = "employee360:roles"
)

// WithTenantID returns a new context carrying the given tenant id.
func WithTenantID(parent context.Context, tenantID string) context.Context {
	return context.WithValue(parent, tenantIDKey, tenantID)
}

// TenantIDFromContext extracts the tenant id injected by the (future)
// tenant-resolution middleware. Persistence-layer code must use this —
// and never a client-supplied tenant id — to scope every query and
// mutation on a tenant-owned table.
func TenantIDFromContext(c context.Context) (string, bool) {
	tenantID, ok := c.Value(tenantIDKey).(string)
	return tenantID, ok
}

// WithUserID returns a new context carrying the given authenticated
// user id.
func WithUserID(parent context.Context, userID string) context.Context {
	return context.WithValue(parent, userIDKey, userID)
}

// UserIDFromContext extracts the authenticated user id injected by the
// (future) auth middleware.
func UserIDFromContext(c context.Context) (string, bool) {
	userID, ok := c.Value(userIDKey).(string)
	return userID, ok
}

// WithRoles returns a new context carrying the given role names.
func WithRoles(parent context.Context, roles []string) context.Context {
	return context.WithValue(parent, rolesKey, roles)
}

// RolesFromContext extracts the authenticated user's roles injected by
// the (future) auth middleware.
func RolesFromContext(c context.Context) ([]string, bool) {
	roles, ok := c.Value(rolesKey).([]string)
	return roles, ok
}
