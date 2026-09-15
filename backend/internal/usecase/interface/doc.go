// Package usecaseinterface holds the usecase ports: one file per feature
// (auth, user, rbac, department, position, holiday, category, tenant,
// audit), one interface per operation, each with a single
// Execute(ctx, ...) (result, error) method.
//
// The directory is named "interface" per plan/architecture/backend.md,
// but "interface" is a Go reserved word, so the package itself is named
// usecaseinterface. Import it with an explicit or implied alias, e.g.:
//
//	import usecaseinterface "github.com/your-org/your-project/backend/internal/usecase/interface"
//
// The delivery layer must depend only on these ports, never on
// usecase/implementation directly — concrete implementations are wired
// in via the DI container (internal/infrastructure/container).
//
// No usecase ports are defined yet — this is Cycle 1 (project
// scaffolding). They are introduced per-feature starting in Cycle 2.
package usecaseinterface
