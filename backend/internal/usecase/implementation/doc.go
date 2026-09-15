// Package usecaseimpl holds the usecase adapters: one file per operation
// (e.g. create_user.go defines CreateUserUseCaseImpl), each implementing
// the matching interface declared in usecase/interface, with an injected
// set of repository/service dependencies, a NewXxxUseCase constructor,
// and a compile-time assertion such as:
//
//	var _ usecaseinterface.CreateUserUseCase = (*CreateUserUseCaseImpl)(nil)
//
// The directory is named "implementation" per
// plan/architecture/backend.md; the package is named usecaseimpl to
// mirror the usecaseinterface naming used for its sibling "interface"
// package (which is a Go reserved word and cannot be a package name).
//
// No usecase implementations are defined yet — this is Cycle 1 (project
// scaffolding). They are introduced per-feature starting in Cycle 2.
package usecaseimpl
