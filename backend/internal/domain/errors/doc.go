// Package errors holds domain-specific sentinel errors (e.g.
// ErrNotFound, ErrTenantMismatch, ErrDuplicateEntry) shared across the
// usecase and infrastructure layers.
//
// This package is part of the domain layer: it must never import Gin,
// GORM, database drivers, or any other external framework.
//
// No sentinel errors are defined yet — this is Cycle 1 (project
// scaffolding). They are introduced alongside the usecases that need
// them starting in Cycle 2.
package errors
