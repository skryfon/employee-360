// Package repository defines the repository interfaces (ports) that the
// infrastructure layer implements (e.g. TenantRepository, UserRepository,
// HolidayRepository). Every method that touches tenant-owned data must be
// scoped by tenant_id resolved from context — never a client-supplied value.
//
// This package is part of the domain layer: it must never import Gin,
// GORM, database drivers, or any other external framework.
//
// No repository ports are defined yet — this is Cycle 1 (project
// scaffolding). Ports are introduced alongside their entities starting in
// Cycle 2.
package repository
