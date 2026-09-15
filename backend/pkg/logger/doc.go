// Package logger will provide a small zerolog wrapper so every part of
// the application constructs its structured logger the same way.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the package exists
// per plan/architecture/backend.md, but the actual wrapper lands in A3
// (middleware chain, health route & server bootstrap), alongside the
// request logger middleware that needs it.
package logger
