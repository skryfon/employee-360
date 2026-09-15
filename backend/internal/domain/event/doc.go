// Package event defines domain events raised by usecases (e.g. for audit
// logging or future async processing).
//
// This package is part of the domain layer: it must never import Gin,
// GORM, database drivers, or any other external framework.
//
// No domain events are defined yet — this is Cycle 1 (project
// scaffolding). They are introduced alongside the usecases that raise
// them starting in Cycle 2.
package event
