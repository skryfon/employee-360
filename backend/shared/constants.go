// Package shared holds cross-cutting constants and small utility types
// used across multiple layers (e.g. API path prefixes, context header
// names). Keep this package dependency-free and small — anything with
// real behavior belongs in pkg/ or a proper layer package instead.
package shared

const (
	// AppName is the service name used in logs, the DB seeder, and
	// anywhere else the platform needs to identify itself.
	AppName = "employee360"

	// APIVersionPrefix is the route prefix for the versioned, headless
	// REST API. Every client-facing business route lives under this
	// prefix (see internal/delivery/http/routes.go).
	APIVersionPrefix = "/api/v1"

	// RequestIDHeader is the HTTP header used to propagate a request's
	// trace id, both inbound (client-supplied) and outbound (echoed
	// back on the response).
	RequestIDHeader = "X-Request-ID"

	// TenantIDHeader is reserved for the multi-tenant resolution
	// middleware landing in Cycle 2. Not wired to anything yet.
	TenantIDHeader = "X-Tenant-ID"
)
