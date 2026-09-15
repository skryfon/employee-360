// Package database will provide the PostgreSQL/GORM connection pool
// used across the application, plus the seeder (in the seeder/
// subpackage) that bootstraps platform data in later cycles.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the package exists
// per plan/architecture/backend.md, but the actual connection logic
// (Connect/Ping/Close) lands in A2 (config, DB connectivity & migration
// runner).
package database
