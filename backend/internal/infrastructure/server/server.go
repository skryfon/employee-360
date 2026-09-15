// Package server will own HTTP server lifecycle: starting the listener
// and shutting it down gracefully on signal.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the package exists
// per plan/architecture/backend.md, but server bootstrap itself lands
// in A3 (middleware chain, health route & server bootstrap).
package server
