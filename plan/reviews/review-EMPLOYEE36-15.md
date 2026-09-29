# Review: EMPLOYEE36-15 — B7 — Bootstrap seeding: system tenant, roles & super admin

> Branch: ebin/feat/EPIC-B/EMPLOYEE36-15 | Last reviewed: 2026-09-29 21:45 | Iteration: 1 | Verdict: 🟡

## Ticket
**Identifier:** EMPLOYEE36-15
**State:** unstarted (Plane state group)
**Link:** n/a (not returned by Plane)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md (Seeding sub-feature)

### Description
`backend/internal/infrastructure/database/seeder/`, run via `cmd/bootstrap`. Seed one system tenant, default roles (`super_admin`, `admin`, `employee`), and the platform Super Admin (password from env config, never hardcoded); `seeder_test.go` verifies idempotency.

### Acceptance Criteria
- [x] AC-1: `make bootstrap` (or `cmd/bootstrap`) succeeds on a clean DB — impl `backend/cmd/bootstrap/main.go:26-58`, `seeder/seeder.go:109-171`; test `seeder_test.go:68-83` (integration tag; first run creates tenant/roles/admin/domain)
- [x] AC-2: Second run produces no duplicate tenant, roles, or admin — impl `seeder.go:122-166` (advisory lock + `ON CONFLICT DO NOTHING` on `uq_roles_tenant_id_name`, `uq_users_tenant_id_email`, `uq_user_roles_user_id_role_id`); test `seeder_test.go:85-98`
- [x] AC-3: Super Admin password read from config/env, never hardcoded — impl `config/config.go:213-220,297-299` (no default), `seeder.go:33-37,70-72`; tests `seeder_validation_test.go:12-21`, `seeder_test.go:108-116`, bcrypt-hash check `seeder_test.go:100-105`

## Latest commit reviewed
`5426d25` — feat(backend): implement idempotent database seeder and bootstrap command (EMPLOYEE36-15)

Scope note: the branch is stacked (EMPLOYEE36-9..14 already reviewed); only commit `5426d25` was reviewed for this ticket.

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] `.env.example:53` — ships a concrete, valid-looking password (`BOOTSTRAP_SUPER_ADMIN_PASSWORD=employee360@2026`) — anyone copying the example verbatim gets a platform super admin with a publicly known password; the comment says "no defaults" but the example effectively is one. Fix: use a placeholder (`<CHANGE_ME>`) like `SMTP_PASSWORD=<RESEND_API_KEY>`, and consider rejecting known-example values in `Options.Validate`.
  - **Fixed (uncommitted):** `.env.example` now uses `<CHANGE_ME>`; `Options.Validate` rejects it with `ErrSuperAdminPasswordPlaceholder`; test `TestOptionsValidate_RejectsPasswordPlaceholder` in `seeder_validation_test.go`.

### 🟢 Minor
- [x] `Makefile:61` — target is `bootstrap-admin`; `make bootstrap` installs deps, while ticket AC-1 and `shared-context.md` say `make bootstrap` seeds. Align docs/ticket or Makefile naming.
  - **Fixed (uncommitted):** Makefile not renamed; docs now say `make bootstrap-admin` (`shared-context.md`, `README.md`, `.claude/agents/backend-agent.md`, `plan/architecture/diagrams/backend-architecture.html`). Ticket AC-1 text still says `make bootstrap` (Plane ticket not edited).
- [x] `seeder.go:173-175` — system tenant looked up by non-unique `name`; a tenant later created with the same name (e.g. "System") would be adopted as the system tenant. Consider a dedicated marker or unique constraint.
  - **Mitigated (uncommitted), no migration:** `ensureTenant` now prefers the same-named tenant that owns the super admin's email domain in `tenant_domains`, then the oldest by name, else creates. Test `TestRun_PrefersDuplicateNamedTenantOwningDomain` (integration). `tenants.name` is still not unique, so duplicates can still be created elsewhere; a unique index/marker would need a migration.
- [ ] (won't fix, no history rewrite) Commit includes unrelated generated output (`backend/docs/*`, `packages/api-client/src/generated/**`) from the EMPLOYEE36-14 auth handlers — noise in this ticket's diff; ideally a separate commit.

## Verdict
- **Score:** 92/100
- **Flag:** 🟡 Reviewer call
- **Notes:** All three ACs have both implementation and test evidence. Seeder is transactional, serialized by a Postgres advisory lock, idempotent, never overwrites an existing admin's password, registers the email domain so login resolves to the system tenant, and fails loudly if the domain belongs to another tenant. Integration tests are build-tagged (`integration`) and run in CI via `make test-backend-integration`; I did not run them locally. Score is ≥85 but the open Major (default password in `.env.example`) makes this a reviewer call.

## Re-review Log
### Fix pass (after iteration 1) — uncommitted, not yet re-reviewed
- Addressed the Major and both seeder/docs Minors above; generated-files Minor deliberately left (no history rewrite).
- Verification: `go build ./...`, `go vet ./...` (also `-tags integration`), `go test -count=1 ./...` pass; seeder integration tests pass against a live PostgreSQL (5 repeated runs).
- Pending: commit and re-review.
