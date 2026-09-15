// Package http will assemble the Gin engine: the base middleware chain
// and route registration. Named "http" per
// plan/architecture/backend.md's internal/delivery/http package;
// callers that also need net/http should import this package under an
// alias (e.g. deliveryhttp) to avoid a naming collision.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the package exists
// per the architecture doc, but the middleware chain, router, and
// health route land in A3 (middleware chain, health route & server
// bootstrap).
package http
