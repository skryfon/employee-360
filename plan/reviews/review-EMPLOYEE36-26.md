# Review: EMPLOYEE36-26 — Backend: Department CRUD API (tenant-scoped)

> Branch: aby/feat/EPIC-E/EMPLOYEE36-26 | Last reviewed: 2026-10-02 | Iteration: 2 | Verdict: 🟢

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
`3e09bf1` — frontend
(iteration 1 reviewed `47de6ba`)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] (resolved in 3e09bf1: delete takes `GetByIDForUpdate`; invite takes `LockDepartmentShared` in-tx; repo lock test added) `usecase/implementation/department/delete.go:42-58` — IsReferenced then Delete with no lock on the department row — a concurrent invite/user create can attach the department between check and soft-delete (soft delete never trips an FK), leaving users pointing at a deleted department — fix: `SELECT … FOR UPDATE` on the department row inside the transaction, and have invite/user creation take a shared lock (or re-check) on it.
- [x] (resolved in 3e09bf1: filters accepted/revoked/expired; `IsReferenced_InvitationStates` test) `persistence/department_repository.go` `IsReferenced` (user_invitations query) — counts every non-deleted invitation, including revoked, accepted and expired ones; contradicts the "active" comment and permanently blocks deleting a department once it has ever been used in an invite — fix: add `accepted_at IS NULL AND revoked_at IS NULL AND expires_at > now()` and a test for revoked/expired invitations.
- [x] (resolved in 3e09bf1: dedupe CTE added; 000016 is not on main so in-place edit is safe) `migrations/000016_…up.sql:3-6` — unique index creation fails if any tenant already has case-insensitive duplicate department names; the precondition is only a comment — fix: add a dedupe step (e.g. suffix duplicates) before `CREATE UNIQUE INDEX`.
- [x] (resolved in 3e09bf1: `.agents` entry removed) `.gitignore` (commit `b1c5f2e` "EMP36-26") — adds `.agents`, but `AGENTS.md`/`CLAUDE.md` document `.agents/` as a checked-in skills/hooks/rules system; out of ticket scope and the commit message is non-descriptive — fix: drop or move to its own justified commit.

### 🟢 Minor
- [x] (resolved in 3e09bf1) `department/create.go:45`, `update.go` — `len(name) > 100` counts bytes, not characters; multi-byte names get rejected early — use `utf8.RuneCountInString`.
- [ ] `migrations/000016_…up.sql` — edited in place after commit; safe because it is not on main, but any dev DB that already applied the old 000016 keeps the old schema and silently skips the dedupe — tell the team to `make migrate-reset` locally.
- [ ] Scope — ticket is titled "Backend: Department CRUD API" but the branch now also adds `is_active` (migration 000018), the `/departments` admin UI, `Switch` in `packages/ui` and an invite-form department select. All documented in cycle-02, so no cross-cycle creep; consider noting it on the ticket.
- [x] (resolved in 3e09bf1: `intQuery` returns 400) `handlers/department_handler.go` `List` — `strconv.Atoi` errors ignored, so `page=abc` silently becomes defaults — return 400 or document.
- [x] (resolved in 3e09bf1: `pgconn.PgError` only) `persistence/department_repository.go:isUniqueViolation` — substring match on "duplicate key" / "23505" is broader than needed; `pgconn.PgError` (+ constraint name) suffices.

## Verdict
- **Score:** 98/100
- **Flag:** 🟢 Merge
- **Notes:** All 7 iteration-1 findings are resolved with tests. The new `is_active` work is tenant-scoped, audited (activate/deactivate), and the inactive-department check runs under the share lock inside the invite transaction. Backend `go test ./...`, frontend typecheck and all vitest suites pass. Two minor notes remain. The `integration`-tagged repo tests (including the new lock/invitation-state/is_active tests) were not run here (no DB), so run `make test-backend-integration` before merge.

## Re-review Log
### Iteration 2 — 2026-10-02 — sha `3e09bf1`
- **Resolved:** all 4 major and 3 minor findings from iteration 1
- **Still open:** none (2 new minor notes below)
- **New issues:** 🟢 000016 edited in place; 🟢 ticket scope wider than its title
- **Score:** 77 → 98 (+21)
- **Verdict:** 🟢 Merge
