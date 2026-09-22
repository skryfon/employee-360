// Package ctx provides typed context accessors for request-scoped values.
package ctx

import "context"

type contextKey string

const (
	tenantIDKey contextKey = "employee360:tenant_id"
	userIDKey   contextKey = "employee360:user_id"
	rolesKey    contextKey = "employee360:roles"
)

// WithTenantID returns a new context carrying the given tenant ID.
func WithTenantID(parent context.Context, tenantID string) context.Context {
	return context.WithValue(parent, tenantIDKey, tenantID)
}

// TenantIDFromContext extracts the tenant ID from the context.
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
