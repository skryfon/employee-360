// Package persistence contains the GORM repository implementations
// (adapters) for the ports declared in internal/domain/repository. Every
// query and mutation on tenant-owned data must scope by the tenant_id
// pulled from context via internal/ctx — never a client-supplied value.
//
// No repository adapters are defined yet — this is Cycle 1 (project
// scaffolding). They are introduced alongside their entities starting
// in Cycle 2.
package persistence
