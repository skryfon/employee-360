// Package container is the dependency-injection wiring point: it will
// construct repositories, domain services, usecases, and handlers, and
// connect them together. cmd/api/main.go will depend only on this
// package (plus config/logger/database bootstrapping), never on
// infrastructure internals directly.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the package exists
// per plan/architecture/backend.md, but actual wiring starts once
// there's something to wire (A2/A3 and later cycles).
package container
