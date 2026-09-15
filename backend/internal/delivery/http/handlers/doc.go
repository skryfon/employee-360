// Package handlers will hold thin Gin request handlers: bind/validate
// input, call a usecase, write a response via
// internal/delivery/http/response. Handlers must never touch GORM/the
// database directly.
//
// No handlers exist yet — this is Cycle 1 (project scaffolding). The
// health handler lands in A3 (middleware chain, health route & server
// bootstrap); feature handlers land in later cycles.
package handlers
