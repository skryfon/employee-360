# Review: EMPLOYEE36-9 — B1 — Migrations: auth, tenant-domain & invitation schema

> Branch: ebin/feat/EPIC-B/EMPLOYEE36-9 | Last reviewed: 2026-09-29 14:00 | Iteration: 1 | Verdict: 🔴

## Ticket
**Identifier:** EMPLOYEE36-9
**State:** In Progress
**Link:** (not available via Plane MCP response — no workspace slug returned)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md (Migrations sub-feature)

### Description
All 12 migration pairs for Cycle 2, in dependency order, via the `create-migration` skill. Foundational — every other ticket in EPIC-B and EPIC-C depends on this landing first. `000002` and `000008`–`000012` are new; `000003`–`000007` keep their original names but shift by one position. `000012_create_river_schema` is River's own generated schema (via River's CLI), not hand-written.

### Acceptance Criteria
- [x] AC-1: `created_at`/`updated_at` on every entity table — verified directly against all 11 hand-written migration files (`000001`–`000011`); every table has both columns with `DEFAULT NOW()`.
- [x] AC-2: An index on every FK column — verified directly; every FK column across `000001`–`000011` has a matching `CREATE INDEX` (e.g. `users.department_id`/`position_id`, `user_invitations`' five FK columns).
- [ ] AC-3: `make migrate` applies all 12 cleanly; `make migrate-down` reverses them cleanly — **unmet, no test evidence for the "down" half** (see 🔴-2 below).
- [ ] AC-4: Full `create-migration` skill invariant checklist satisfied for each hand-written pair (000001–000011) — **unmet**: `000007_create_user_roles` fails the skill's Tenancy rule (see 🔴-1 below).

## Latest commit reviewed
`76fddc927ef331df052619902e4f15bd0a73cc4c` — feat(backend): add auth, tenant-domain & invitation schema migrations

## Findings

### 🔴 Critical

- [ ] `backend/migrations/000007_create_user_roles.up.sql` — `user_roles` has no `tenant_id` column, contradicting both `CLAUDE.md` Invariant 1 ("Every database table holding tenant data must carry a `tenant_id` column") and the `create-migration` skill's Tenancy rule ("Every table *except* `tenants` itself must include a `tenant_id` column ... plus an index on it"), which AC-4 explicitly requires be satisfied for every hand-written pair. The migration's own comment says cross-tenant consistency is "enforced at the usecase layer, not via a redundant column here" — but no usecase exists yet in this migrations-only ticket, so today nothing at the DB level stops a row pairing a tenant-A `user_id` with a tenant-B `role_id`. **Fix:** add `tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE` + `CREATE INDEX idx_user_roles_tenant_id ON user_roles (tenant_id)` to the up migration, and the matching drops to the down migration.

- [ ] AC-3's "make migrate-down reverses them cleanly" half has no automated test evidence. `backend/integration/migrate_test.go:75-129` (`TestMigrate_LiveDatabaseApplyAndRevert`) only applies/reverts a single synthetic smoke-test migration written to a `t.TempDir()`, never the real `backend/migrations/` directory. `.github/workflows/backend-ci.yml:163-167` (the `smoke-test` job) runs `go run ./cmd/migrate up` and `status` against the real migrations but never `down`. So the 12 real down-scripts, run in sequence, have never actually been executed anywhere in CI or in a test. **Failure scenario:** a mismatched drop order or a typo'd index/table name in one of the 12 down files (e.g. `000012`'s function/type/index drop ordering) would only surface the first time someone runs `make migrate-down` for real — e.g. mid-incident. **Fix:** add a step (CI job or integration test) that runs `go run ./cmd/migrate down 12` (or equivalent) against the real `backend/migrations/` directory right after `up`, and assert success.

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [ ] `backend/migrations/000011_create_user_invitations.up.sql` — `role_id` and `invited_by` FKs have no explicit `ON DELETE` behavior (defaults to `NO ACTION`), while `department_id`/`position_id` on the same table use `ON DELETE SET NULL`. Plausibly intentional (protects referential integrity so a role/user with live invitations can't be deleted out from under them), but worth a one-line comment confirming it's deliberate so a future migration author doesn't "fix" the inconsistency.

## Verdict
- **Score:** N/A — hard gate hit (unmet AC), score not computed per review rules
- **Flag:** 🔴 Block
- **Notes:** Structurally this is a clean, well-commented set of 12 migrations — timestamps, FK indexes, naming, and down-migration reversal logic are all correct except for one gap. The blocker is narrow: `000007_create_user_roles` needs a `tenant_id` column to satisfy the project's own non-negotiable multi-tenancy invariant (currently the only table besides `tenants` without one), and the "migrate-down reverses cleanly" half of AC-3 needs an actual automated run against the real migration set rather than a synthetic smoke test. Both are small, contained fixes — add the column/index pair (+ down-migration drops) and add one CI/test step that runs `down` against `backend/migrations/`. Once those land, re-run with `--re-review`.

## Re-review Log
_(empty — first review)_
