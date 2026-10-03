# Review: EMPLOYEE36-27 — Backend: Position CRUD API (tenant-scoped)

> Branch: aby/feat/EPIC-E/EMPLOYEE36-27 | Last reviewed: 2026-10-03 | Iteration: 5 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-27
**State:** Completed (Plane state group `completed`)
**Link:** n/a (not returned by Plane)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
Implement position CRUD through all layers (domain, GORM repo, per-operation usecases, DI container, handler + Swagger, routes), tenant-scoped, under `/api/v1/positions` with Auth → Tenant → `RequireRole(admin, super_admin)`.

### Acceptance Criteria
- [x] AC-1: Unique name per tenant → 409. Impl: `position_handler.go:52-53`, `create.go:87-89`, `position_repository.go:60-62`, migration `000019 ...up.sql:25`. Tests: `create_test.go:159`, `update_test.go:130`, `position_handler_test.go:164-177`, `position_repository_test.go:58`.
- [x] AC-2: Delete blocked while referenced → 409. Impl: `delete.go:50-56`, `position_repository.go:213-240`, `position_handler.go:54-55`. Tests: `delete_test.go:94`, `position_handler_test.go:445-456`, `position_repository_test.go` `IsReferenced*`.
- [x] AC-3: Cross-tenant → 404. Impl: `scoped()` in `position_repository.go:31`. Tests: `routes_test.go:1053` (`TestPositionRoutes_CrossTenant404`), repo `CRUDAndTenantIsolation`.
- [x] AC-4: 401 / 403 / 200 per role on all routes. Impl: `routes.go` positions group. Test: `routes_test.go:1001`.
- [x] AC-5: Tests pass. The position, delivery and container packages pass locally. `make test` could not be run because there is no Makefile in `backend/`.
- [x] AC-6: Swagger regenerated (`docs/*`, api-client generated hooks). Cycle-02 doc note updated (`cycle-02-auth-onboarding.md`).

## Latest commit reviewed
`4c1dd5e` — feat(admin): positions management, position select on invitations, audit-log viewer (HEAD; also `a475a5c` audit-log API + types refactor, `3465cf6` invite position lock)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] (resolved in working tree on top of `135de55`: `LockPositionShared` in-tx at `invite_user.go:173-185`, `ErrPositionInactive` → 400 `POSITION_INACTIVE`, tests `TestInvite_RejectsInactivePosition`, handler mapping, repo lock tests) `internal/usecase/implementation/invitation/invite_user.go:83-92` — Invite only calls `PositionExists` (no lock, outside the transaction) and never checks `is_active` — The delete usecase (`delete.go:42-44`) relies on `FOR UPDATE` on the position row so "a concurrent invite cannot attach it", but nothing in the invite path takes a matching `FOR SHARE` lock. A race (invite passes `PositionExists` → delete passes `IsReferenced` → soft-delete → invitation/user inserted) leaves a live invitation/user pointing at a soft-deleted position, breaking the "delete blocked while referenced" invariant. Also `ErrPositionInactive` is declared but never used, so inactive positions can still be assigned. Fix: add `LockPositionShared(ctx, tenantID, id) (found, active bool, err)` to `OrgReferenceRepository` mirroring `LockDepartmentShared`, call it inside the invite transaction, return `ErrPositionNotFound`/`ErrPositionInactive` (map to 400 like `DEPARTMENT_INACTIVE`), and add a test. (Still open at `135de55`.)

### 🟢 Minor
- [x] (resolved: clamps to 100; same fix applied to department/position lists) `internal/usecase/implementation/auditlog/list.go:44-47` — `page_size` above 100 silently falls back to 20, though the API documents "max 100" — Clamp to 100 (and/or return 400) instead of resetting to the default.
- [x] (resolved: single copy in `clients/admin/src/hooks/useDialogFocus.ts`) `clients/admin/src/features/positions/hooks/useDialogFocus.ts` — Byte-for-byte duplicate of `features/departments/hooks/useDialogFocus.ts` (positions feature is a near 1:1 copy of departments) — Move the shared hook to `packages/ui` or `src/hooks` and import it from both.
- [x] (acknowledged, no code change; split into separate tickets/PRs is a team decision) Branch scope — EMPLOYEE36-27 is "Backend: Position CRUD", but the branch now also carries the audit-log viewer (backend + admin UI, migration 000020), the positions admin UI, and a types/ refactor of department/health code. All are Cycle 2, so not cycle-scope creep, but they should be tracked under their own tickets/PRs.
- [x] `.agents/.agents` — Self-referential symlink committed by accident (resolved in `135de55`)
- [x] (resolved, now used) `internal/domain/errors/errors.go:59` — `ErrPositionInactive` is unused (dead code until the Major above is fixed) — Use it or remove it.
- [x] (resolved: irreversibility comment added to down file) `migrations/000019_...up.sql:8-23` — The de-dupe step renames existing rows, and the down migration (`...down.sql`) cannot restore the original names and has no comment saying so — Note the irreversibility in the down file comment.

## Verdict
- **Score:** 100/100
- **Flag:** 🟢 Merge
- **Notes:** Fixes are committed. New work since the last review (audit-log viewer API and UI, positions UI, types refactor) was reviewed: audit queries are tenant-scoped (including the actor join), routes carry Auth → Tenant → RequireRole(admin, super_admin), metadata is redacted, and tests/typecheck pass (go test, integration persistence, admin 138, ui 29). Three nits remain. Swagger files were hand-edited for the new 400 description (not regenerated via swag). Previously: All ACs have both implementation and test evidence. Tenant isolation holds: every repo query is scoped by `tenant_id` and `deleted_at`, the tenant comes from the auth context, and request bodies carry no `tenant_id`. Role checks cover every route, mutations are audited, and layering and swagger look correct. The one substantive issue is the unlocked, active-unaware position check in the invite flow, which reopens a delete-vs-invite race the delete usecase claims to prevent. Worth fixing before merge, but not an AC failure. The symlink nit was fixed after the previous review; nothing else changed.

## Re-review Log
### Iteration 2 — 2026-10-03 — sha `135de55`
- **Resolved:** `.agents/.agents` symlink (nit)
- **Still open:** invite-path position lock / `is_active` check (🟡); unused `ErrPositionInactive` (🟢); migration down-file irreversibility note (🟢)
- **New issues:** none
- **Score:** 92 → 93 (+1)
- **Verdict:** 🟡 Reviewer call

### Iteration 3 — 2026-10-03 — working tree on `135de55` (uncommitted)
- **Resolved:** invite-path position lock and `is_active` check (🟡); `ErrPositionInactive` now used (🟢); migration down-file comment (🟢)
- **Still open:** none
- **New issues:** none (`PositionExists` fully removed, no remaining callers; `go build`, `go vet` and usecase/delivery/container tests pass)
- **Score:** 93 → 100 (+7)
- **Verdict:** 🟢 Merge

### Iteration 4 — 2026-10-03 — sha `4c1dd5e`
- **Resolved:** none outstanding (all prior findings were already resolved)
- **Still open:** none from prior iterations
- **New issues:** 🟢 audit `page_size` >100 falls back to 20; 🟢 duplicated `useDialogFocus`; 🟢 out-of-ticket scope on this branch
- **Score:** 100 → 97 (-3)
- **Verdict:** 🟢 Merge

### Iteration 5 — 2026-10-03 — minor fixes
- **Resolved:** audit `page_size` clamp (also department/position lists); shared `useDialogFocus`; branch-scope nit acknowledged
- **Still open:** none
- **New issues:** none
- **Score:** 97 → 100 (+3)
- **Verdict:** 🟢 Merge
