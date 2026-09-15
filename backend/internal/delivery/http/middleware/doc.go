// Package middleware will hold Gin middleware: request_id, logger,
// cors, recovery (the base chain), plus auth and tenant-resolution
// middleware added once there's something to protect.
//
// No middleware exists yet — this is Cycle 1 (project scaffolding). The
// base chain lands in A3 (middleware chain, health route & server
// bootstrap); auth/tenant middleware land in Cycle 2.
package middleware
