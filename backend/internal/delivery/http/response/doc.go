// Package response will provide the standardized JSON envelope and
// error helpers every handler must use instead of calling c.JSON
// directly.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the package exists
// per plan/architecture/backend.md, but the envelope helpers land in A3
// (middleware chain, health route & server bootstrap), alongside the
// first handler that needs them.
package response
