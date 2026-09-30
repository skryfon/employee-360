# Review: EMPLOYEE36-7 — EPIC-B: Cycle 2 — Auth Core, Tenant Resolution & Email Service (Overview)

> Branch: feature/EPIC-B | Last reviewed: 2026-09-30 | Iteration: 2 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-7 (epic; children EMPLOYEE36-9..15, 24 = B1..B8)
**State:** started
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
Real auth (login/refresh/logout/forgot/reset), tenant-domain resolution, async River-backed email outbox, bootstrap seeding. No frontend, no invitations (EPIC-C).

### Acceptance Criteria (epic Definition of Done)
- [x] DoD-1: `make migrate` applies 12 migrations, `migrate-down` reverses — `integration/migrate_test.go:253` `TestMigrate_LiveDatabaseRealMigrationsApplyAndRevert` (passes)
- [x] DoD-2: bootstrap seeds tenant/roles/super admin, idempotent — `database/seeder` tests (pass)
- [x] DoD-3: admin login returns access+refresh — `routes.go:50`; usecase + handler tests
- [x] DoD-4: reset end to end with worker — `integration/worker_test.go:50`, `outbox_test.go`; usecase tests
- [x] DoD-5: employee login same endpoint/usecase (single mechanism by design)
- [x] DoD-6: refresh + logout against tracked refresh tokens — `routes.go:50-58`; usecase tests

Child tickets B1–B8 each carry a prior review with 🟢 (B2/EMPLOYEE36-10 shows 🟡 at its last iteration).

## Latest commit reviewed
`9973d2d` — test(backend): add live-DB auth flow integration tests (EMPLOYEE36-7)

## Verification run
- `go build ./...`, `go vet ./...` clean
- `go test ./...` and `go test -tags=integration -count=1 ./...` (live Postgres) all pass
- Usecase/domain packages import no River/GORM/sql; `EmailService.Send` called only from `job/email_worker.go`; no Resend SDK; tenant comes only from JWT claims (`middleware/tenant.go`); only `/logout` sits behind Auth+Tenant middleware; no `clients/` changes (scope respected)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] Prior EMPLOYEE36-10 (B2) review last recorded 🟡 — confirm its open findings were closed (`plan/reviews/review-EMPLOYEE36-10.md`) before merging the epic. **(resolved in 9973d2d** — re-checked against current code: every code finding in review-EMPLOYEE36-10.md is already resolved and ticked; the sole remaining unchecked item is a ticket-scope process note with no code defect, now marked acknowledged there.)
- [x] No live-DB, HTTP-level test exercising login → refresh → logout (or forgot → worker → reset) through the real router; `/auth/*` is only covered in `internal/delivery/http/routes_test.go` and mocked usecase tests. DoD-3/5/6 are therefore proven per layer, not end to end. Add one `integration/auth_flow_test.go`. **(resolved in 9973d2d** — `backend/integration/auth_flow_test.go` added: login/refresh-rotation/logout-revocation and forgot-password -> river_job -> worker -> reset-password (all refresh tokens revoked) through the real router/container on live Postgres.)

### 🟢 Minor
- [ ] `ea2be49` and `9973d2d` are not yet pushed (`feature/EPIC-B` is 2 ahead of origin).

## Verdict
- **Score:** 99/100
- **Flag:** 🟢 Merge
- **Notes:** Both majors closed. New `auth_flow_test.go` proves login/refresh/logout and forgot->worker->reset end to end on live Postgres; build, vet and integration suite pass. Only remaining item is pushing the branch.

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `9973d2d`
- **Resolved:** B2 (EMPLOYEE36-10) open-findings check; missing end-to-end auth-flow integration test
- **Still open:** unpushed commits (minor)
- **New issues:** none
- **Score:** 89 → 99 (+10)
- **Verdict:** 🟢 Merge
