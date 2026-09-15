// Package ucshared holds usecase-layer helpers shared across feature
// usecases (e.g. a Transactor abstraction for cross-repository
// transactions), per plan/architecture/backend.md.
//
// Nothing lives here yet — this is Cycle 1 (project scaffolding). The
// first shared helper (Transactor) is added once a usecase actually
// needs to coordinate multiple repositories in one transaction.
package ucshared
