# Review: EMPLOYEE36-26 — Backend: Department CRUD API (tenant-scoped)

> Branch: aby/feat/EPIC-E/EMPLOYEE36-26 | Last reviewed: 2026-10-02 | Iteration: 1 | Verdict: 🟡

## Ticket
**Identifier:** EMPLOYEE36-26
**State:** started (Plane state group)
**Link:** n/a (not returned by Plane)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
Department CRUD through all layers (entity, repo port + GORM adapter, one usecase per operation, DI container, handler + swagger, routes with Auth → Tenant → RequireRole(admin, super_admin)). tenant_id from context only.

### Acceptance Criteria
- [x] AC-1: Unique name per tenant → 409 — impl `usecase/implementation/department/create.go:68`, `persistence/department_repository.go:Create` (23505 → NameTaken), migration 000016 unique index; tests `handlers/department_handler_test.go:129`, `department/create_test.go`, `persistence/department_repository_test.go:56` (integration tag)
- [x] AC-2: Delete blocked while referenced → 409 — impl `department/delete.go:50`, `department_repository.go:IsReferenced`; tests `department_handler_test.go:329`, `department/delete_test.go`, `department_repository_test.go:142` (integration tag)
- [x] AC-3: Cross-tenant → 404 — impl tenant-scoped repo queries; test `routes_test.go` `TestDepartmentRoutes_CrossTenant404` (admin and super_admin)
- [x] AC-4: 401/403/200 per role on all routes — `routes.go:112-126`; `TestDepartmentRoutes_AuthAndRoles`
- [x] AC-5: Tests pass — `go build` and unit tests for usecase/delivery/container pass locally; the `integration`-tagged repo tests were NOT run (no DB here)
- [x] AC-6: Swagger regenerated (`backend/docs/*`) and cycle-02 doc note updated

## Latest commit reviewed
`47de6ba` — Merge branch 'TESTING' into aby/feat/EPIC-E/EMPLOYEE36-26
(ticket work is `54d601f`; `b1c5f2e` only touches `.gitignore`)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [ ] `usecase/implementation/department/delete.go:42-58` — IsReferenced then Delete with no lock on the department row — a concurrent invite/user create can attach the department between check and soft-delete (soft delete never trips an FK), leaving users pointing at a deleted department — fix: `SELECT … FOR UPDATE` on the department row inside the transaction, and have invite/user creation take a shared lock (or re-check) on it.
- [ ] `persistence/department_repository.go` `IsReferenced` (user_invitations query) — counts every non-deleted invitation, including revoked, accepted and expired ones; contradicts the "active" comment and permanently blocks deleting a department once it has ever been used in an invite — fix: add `accepted_at IS NULL AND revoked_at IS NULL AND expires_at > now()` and a test for revoked/expired invitations.
- [ ] `migrations/000016_…up.sql:3-6` — unique index creation fails if any tenant already has case-insensitive duplicate department names; the precondition is only a comment — fix: add a dedupe step (e.g. suffix duplicates) before `CREATE UNIQUE INDEX`.
- [ ] `.gitignore` (commit `b1c5f2e` "EMP36-26") — adds `.agents`, but `AGENTS.md`/`CLAUDE.md` document `.agents/` as a checked-in skills/hooks/rules system; out of ticket scope and the commit message is non-descriptive — fix: drop or move to its own justified commit.

### 🟢 Minor
- [ ] `department/create.go:45`, `update.go` — `len(name) > 100` counts bytes, not characters; multi-byte names get rejected early — use `utf8.RuneCountInString`.
- [ ] `handlers/department_handler.go` `List` — `strconv.Atoi` errors ignored, so `page=abc` silently becomes defaults — return 400 or document.
- [ ] `persistence/department_repository.go:isUniqueViolation` — substring match on "duplicate key" / "23505" is broader than needed; `pgconn.PgError` (+ constraint name) suffices.

## Verdict
- **Score:** 77/100
- **Flag:** 🟡 Reviewer call
- **Notes:** Clean layering, tenant scoping on every repo method, admin-only RequireRole on all routes, audit entries on all mutations, and good AC test coverage (handler, usecase, routes incl. cross-tenant super_admin). No critical issues. Main concerns: a delete/reference race, IsReferenced treating dead invitations as references, a migration that can fail on existing duplicates, and an unrelated `.gitignore` change. Repo tests are `integration`-tagged and were not run in this review.

## Re-review Log
_(empty — first review)_
