// Package config will load application configuration (env vars, .env,
// config.yaml) via Viper.
//
// This is Cycle 1 (project scaffolding, ticket A1) — the package exists
// per plan/architecture/backend.md's directory tree, but config loading
// itself (the Config struct, defaults, Viper wiring) lands in A2
// (config, DB connectivity & migration runner).
package config
