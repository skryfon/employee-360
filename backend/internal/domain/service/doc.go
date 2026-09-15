// Package service defines domain service interfaces (ports) implemented by
// the infrastructure layer, such as TokenService (JWT issue/verify) and
// HashService (password hashing/comparison).
//
// This package is part of the domain layer: it must never import Gin,
// GORM, database drivers, or any other external framework.
//
// No domain service ports are defined yet — this is Cycle 1 (project
// scaffolding). They are introduced alongside auth work starting in
// Cycle 2.
package service
