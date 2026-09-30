// Package ctx provides typed context accessors for request-scoped values.
package ctx

import "context"

type contextKey string

const (
	tenantIDKey  contextKey = "employee360:tenant_id"
	userIDKey    contextKey = "employee360:user_id"
	rolesKey     contextKey = "employee360:roles"
	clientIPKey  contextKey = "employee360:client_ip"
	userAgentKey contextKey = "employee360:user_agent"
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

// WithClientIP returns a new context carrying the given client IP address.
func WithClientIP(parent context.Context, ip string) context.Context {
	return context.WithValue(parent, clientIPKey, ip)
}

// ClientIPFromContext extracts the client IP address from the context.
func ClientIPFromContext(c context.Context) (string, bool) {
	ip, ok := c.Value(clientIPKey).(string)
	return ip, ok
}

// WithUserAgent returns a new context carrying the given user agent string.
func WithUserAgent(parent context.Context, userAgent string) context.Context {
	return context.WithValue(parent, userAgentKey, userAgent)
}

// UserAgentFromContext extracts the user agent string from the context.
func UserAgentFromContext(c context.Context) (string, bool) {
	ua, ok := c.Value(userAgentKey).(string)
	return ua, ok
}
