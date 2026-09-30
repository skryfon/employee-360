# Review: EMPLOYEE36-9 — B1 — Migrations: auth, tenant-domain & invitation schema

> Branch: ebin/feat/EPIC-B/EMPLOYEE36-9 | Last reviewed: 2026-09-29 15:30 | Iteration: 2 | Verdict: 🟢

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
- [x] AC-3: `make migrate` applies all 12 (now 14) cleanly; `make migrate-down` reverses them cleanly — **resolved in `511f6c0`**: `backend/integration/migrate_test.go`'s new `TestMigrate_LiveDatabaseRealMigrationsApplyAndRevert` applies the real `backend/migrations/` directory end-to-end then reverts the full count, and `.github/workflows/backend-ci.yml`'s `smoke-test` job now runs `down` against the real directory too. Independently re-run against a live Postgres in this session — `go test -tags=integration ./integration/... -run TestMigrate -v` — all three migrate tests pass, and `go run ./cmd/migrate status` confirms "no migrations applied yet" afterward (clean revert).
- [x] AC-4: Full `create-migration` skill invariant checklist satisfied for each hand-written pair (000001–000011, plus new 000013) — **resolved in `511f6c0`**: `backend/migrations/000013_add_tenant_id_to_user_roles.up.sql` adds `tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE` + `idx_user_roles_tenant_id` index to `user_roles`, satisfying the skill's Tenancy rule. Applied and reverted cleanly in the live-DB test run above.

## Latest commit reviewed
`511f6c08163b607881243d5c90d044266d1b17ea` — fix(backend): address EMPLOYEE36-9 review findings

## Findings

### 🔴 Critical

- [x] `backend/migrations/000007_create_user_roles.up.sql` — `user_roles` has no `tenant_id` column, contradicting both `CLAUDE.md` Invariant 1 ("Every database table holding tenant data must carry a `tenant_id` column") and the `create-migration` skill's Tenancy rule ("Every table *except* `tenants` itself must include a `tenant_id` column ... plus an index on it"), which AC-4 explicitly requires be satisfied for every hand-written pair. The migration's own comment says cross-tenant consistency is "enforced at the usecase layer, not via a redundant column here" — but no usecase exists yet in this migrations-only ticket, so today nothing at the DB level stops a row pairing a tenant-A `user_id` with a tenant-B `role_id`. **Fix:** add `tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE` + `CREATE INDEX idx_user_roles_tenant_id ON user_roles (tenant_id)` to the up migration, and the matching drops to the down migration. **(resolved in `511f6c0`** — via new `000013_add_tenant_id_to_user_roles` pair, since `000007` is already committed and can't be edited directly; verified applied/reverted cleanly against a live DB.)

- [x] AC-3's "make migrate-down reverses them cleanly" half has no automated test evidence. `backend/integration/migrate_test.go:75-129` (`TestMigrate_LiveDatabaseApplyAndRevert`) only applies/reverts a single synthetic smoke-test migration written to a `t.TempDir()`, never the real `backend/migrations/` directory. `.github/workflows/backend-ci.yml:163-167` (the `smoke-test` job) runs `go run ./cmd/migrate up` and `status` against the real migrations but never `down`. So the 12 real down-scripts, run in sequence, have never actually been executed anywhere in CI or in a test. **Failure scenario:** a mismatched drop order or a typo'd index/table name in one of the 12 down files (e.g. `000012`'s function/type/index drop ordering) would only surface the first time someone runs `make migrate-down` for real — e.g. mid-incident. **Fix:** add a step (CI job or integration test) that runs `go run ./cmd/migrate down 12` (or equivalent) against the real `backend/migrations/` directory right after `up`, and assert success. **(resolved in `511f6c0`** — new `TestMigrate_LiveDatabaseRealMigrationsApplyAndRevert` test plus a new CI smoke-test step, both dynamically counting migrations rather than hardcoding; independently re-run in this session against a live Postgres and confirmed passing, with the DB left clean afterward.)

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [x] `backend/migrations/000011_create_user_invitations.up.sql` — `role_id` and `invited_by` FKs have no explicit `ON DELETE` behavior (defaults to `NO ACTION`), while `department_id`/`position_id` on the same table use `ON DELETE SET NULL`. Plausibly intentional (protects referential integrity so a role/user with live invitations can't be deleted out from under them), but worth a one-line comment confirming it's deliberate so a future migration author doesn't "fix" the inconsistency. **(resolved in `511f6c0`** — new `000014_document_user_invitations_fk_delete_behavior` pair adds `COMMENT ON CONSTRAINT` for both FKs, comment-only, no schema change; constraint names verified correct since the up migration ran successfully in the live-DB test.)

## Verdict
- **Score:** 100/100
- **Flag:** 🟢 Merge
- **Notes:** All findings from iteration 1 are resolved, correctly worked around the "never edit a committed migration" convention by adding new `000013`/`000014` pairs instead of touching `000007`/`000011`, and both prior hard-gate ACs (AC-3, AC-4) now have direct test evidence. Independently re-ran `go build`, `go vet`, `gofmt -l`, and the full `integration` test suite against a live Postgres in this session (not just trusting the fix report) — everything passes, including the new real-migrations round-trip test, and the database is left clean after the down revert. No new issues introduced by the fix commit. Clear to merge.

## Re-review Log

### Iteration 2 — 2026-09-29 — sha `511f6c0`
- **Resolved:** `000007_create_user_roles` missing `tenant_id` (via new `000013`); AC-3 "migrate-down reverses cleanly" untested (via new integration test + CI step); `000011_create_user_invitations` undocumented `ON DELETE` inconsistency (via new `000014` constraint comments).
- **Still open:** none.
- **New issues:** none.
- **Score:** N/A → 100 (+100)
- **Verdict:** 🟢 Merge
