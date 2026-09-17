# Review: EMPLOYEE36-3 — A2 — Backend: config, DB connectivity & migration runner

> Branch: ebin/feat/EPIC-A/EMPLOYEE36-3 | Last reviewed: 2026-09-15 13:05 | Iteration: 2 | Verdict: 🟡

## Ticket
**Identifier:** EMPLOYEE36-3
**State:** Backlog (state_group: backlog)
**Link:** N/A (not returned by Plane MCP)
**Cycle:** `plan/cycles/cycle-01-project-setup.md` (Cycle 1 — Project Setup & Scaffolding)

### Description
Set up Viper configuration loading, local PostgreSQL infrastructure, GORM database
connectivity, and the `golang-migrate` runner.

- Viper config loading in `config/config.go`, `config.yaml.example`, env vars + `.env`
- PostgreSQL service in `docker-compose.yml`, GORM connection in
  `internal/infrastructure/database/postgres.go`, reads config, fails fast (exit 1,
  stderr containing `"database unreachable"` + underlying error, no retry loop, no
  partial startup)
- `golang-migrate` wired through `cmd/migrate/main.go`, connects and executes cleanly,
  handles zero migrations without error, `migrations/` stays empty
- Depends on A1 (EMPLOYEE36-2)

### Acceptance Criteria
- [x] AC-1: `docker-compose up -d` brings up PostgreSQL successfully — implementation:
      `backend/docker-compose.yml` (now the single canonical file; root `docker-compose.yml`
      is a symlink to it). Test evidence: `.github/workflows/backend-ci.yml`'s "verify
      docker-compose configurations" step validates the compose file via `docker compose
      config` (both direct and through the symlink), and the CI `services.postgres` block
      uses the same image/credentials, proven reachable end-to-end by the live-DB tests
      below. Note: no step literally runs `docker compose up -d` against this file — see
      🟢 Minor finding for closing that last gap.
- [x] AC-2: Application connects to PostgreSQL using configured values — implementation:
      `backend/cmd/api/main.go`, `backend/internal/infrastructure/database/postgres.go:Connect`.
      Test evidence (new): `backend/cmd/api/main_test.go:TestAPI_LiveDatabaseSuccess`,
      `backend/internal/infrastructure/database/postgres_test.go:TestConnect_SuccessLiveDatabase`
      — both fail the build in CI if PostgreSQL isn't reachable (`sha 535e903`).
- [x] AC-3: Database connection failure produces exit code 1 and a stderr log line
      containing "database unreachable" and the underlying error, within a bounded
      startup timeout (no retry loop) — unchanged since iteration 1, still passing.
- [x] AC-4: `cmd/migrate` connects and executes cleanly against PostgreSQL —
      implementation: `backend/cmd/migrate/main.go`. Test evidence (new):
      `backend/cmd/migrate/main_test.go:TestMigrate_LiveDatabaseZeroMigrationsClean`,
      `TestMigrate_LiveDatabaseApplyAndRevert` (`sha 535e903`).
- [x] AC-5: Migration runner handles zero migrations without error — test evidence (new):
      `backend/cmd/migrate/main_test.go:TestMigrate_LiveDatabaseZeroMigrationsClean`
      asserts `migrate up`/`status` succeed and report "no migrations applied yet"
      against a live, empty database (`sha 535e903`).

## Note on review base

`origin/main` on GitHub (`17bfe53`) has not yet caught up with PR #3
(EMPLOYEE36-2), which is already merged into this branch's own history via commit
`42798d2` and already has its own review (`plan/reviews/review-EMPLOYEE36-2.md`,
verdict Merge). Diffing straight against `origin/main` would re-surface that
already-reviewed work, so this review instead diffs `42798d2..HEAD` — i.e. only the
commit(s) added for EMPLOYEE36-3 on top of the merged EMPLOYEE36-2 base.

## Latest commit reviewed
`3afb2c5` — feat(config): enforce custom JWT secret validation in production and staging (EMPLOYEE36-3)

## Findings

### 🔴 Critical

- [x] `backend/cmd/api/main.go`, `backend/internal/infrastructure/database/postgres_test.go` — AC-2 unmet: no automated test exercised a successful connection to a running PostgreSQL — (resolved in `535e903`: `backend/cmd/api/main_test.go:TestAPI_LiveDatabaseSuccess` and `backend/internal/infrastructure/database/postgres_test.go:TestConnect_SuccessLiveDatabase` connect against a live Postgres and fail the build (`t.Fatalf`) rather than skip when running in CI.)
- [x] `backend/cmd/migrate/main.go` — AC-4 and AC-5 unmet: no test ran `migrate up`/`status` against a real PostgreSQL — (resolved in `535e903`: `backend/cmd/migrate/main_test.go:TestMigrate_LiveDatabaseZeroMigrationsClean` and `TestMigrate_LiveDatabaseApplyAndRevert` exercise both the zero-migrations case and a real apply/status/revert cycle against a live database. Verified locally: `go test ./...` passes end-to-end against a live Postgres.)
- [x] `docker-compose.yml`, `backend/docker-compose.yml` — AC-1 unmet: nothing verified `docker-compose up -d` actually brings Postgres up successfully — (substantially resolved in `535e903`/`5911325`: the two compose files were deduplicated into one canonical `backend/docker-compose.yml` with the root file now a symlink to it, `.github/workflows/backend-ci.yml` added a "verify docker-compose configurations" step that runs `docker compose config` against both paths, and the CI `services.postgres` block uses the identical image/credentials, proven reachable by the live-DB tests above. Residual gap: no step literally runs `docker compose up -d` — tracked as a 🟢 Minor below rather than reopening this as Critical, since the functional behavior it guards is now proven end-to-end.)
- [x] `.github/workflows/backend-ci.yml:24-32` combined with `backend/cmd/api/main.go` / `backend/cmd/migrate/main.go` — This branch would have broken CI (no Postgres service, but `cmd/api`/`cmd/migrate` now do real DB work) — (resolved in `535e903`: `backend-ci.yml` now declares a `services.postgres` container (postgres:16-alpine, matching `DATABASE_*` env), and the smoke-test step now includes `go run ./cmd/migrate status` in addition to `up`.)

### 🟡 Major

- [x] `backend/internal/infrastructure/database/postgres.go:~1205` (`EnsureDatabaseExists`) — Silently auto-created the configured database instead of failing fast, deviating from the AC's fail-fast/no-partial-startup intent — (resolved in `535e903`: `EnsureDatabaseExists` and `isDatabaseNotExistError` were removed entirely; `cmd/api`/`cmd/migrate`/`MustConnect` now go straight to `database.Connect`, so an unreachable *or* missing database both fail fast with no side effects.)
- [x] `docker-compose.yml` and `backend/docker-compose.yml` — Byte-identical duplicate files that would drift out of sync — (resolved in `535e903`: root `docker-compose.yml` is now a symlink to `backend/docker-compose.yml`, a single source of truth.)
- [x] `backend/cmd/api/main.go:143-151`, `backend/cmd/migrate/main.go:269-281` — Redundant back-to-back GORM connections on every invocation — (resolved in `535e903` as a side effect of removing `EnsureDatabaseExists`: `cmd/api` now opens exactly one connection; `cmd/migrate` is down to two (its own verification connect + `golang-migrate`'s internal driver connect), which is reasonable.)
- [ ] `.gitignore` (new), `plan/reviews/review-EMPLOYEE36-2.md`, `plan/reviews/review-EMPLOYEE36-3.md` — **New: out-of-scope change that broke the team-mate-review persistence model.** Commit `5911325` ("chore: ignore review folder and contents in .gitignore") adds `plan/reviews/` and `reviews/` to `.gitignore` and `git rm`'s both the already-merged `review-EMPLOYEE36-2.md` and this very ticket's own `review-EMPLOYEE36-3.md` from tracking. This is unrelated to EMPLOYEE36-3's scope (config/DB/migrate work) — it's cycle-scope creep into repo tooling policy. Concretely confirmed: `git check-ignore -v plan/reviews/review-EMPLOYEE36-3.md` now matches `.gitignore:3:reviews/`, so this review file (and any future re-review) will **not** be committed or visible to teammates pulling this branch or reviewing the eventual PR, defeating the skill's "persists findings... so re-reviews across sessions are low-effort" design. Recommend reverting this `.gitignore`/deletion out of this ticket's diff and, if the team genuinely wants review artifacts untracked, making that a deliberate, separate decision (not bundled silently into a backend-connectivity ticket).

### 🟢 Minor

- [x] `backend/config/config.go:171` — Hardcoded JWT default secret used as-is with no non-development safeguard — (resolved in `3afb2c5`: `Config.Validate()` now rejects `Load()` in `production`/`staging` environments if `JWT.Secret` is still `DefaultJWTSecret` or under 32 chars, with matching test coverage in `backend/config/config_test.go:TestConfig_JWTSecretValidationInProduction`.)
- [ ] `.github/workflows/backend-ci.yml` "verify docker-compose configurations" step — Only runs `docker compose config` (syntax/interpolation check), never an actual `docker compose -f docker-compose.yml up -d` + `pg_isready`/`docker compose ps --status running` assertion. Low priority given the live-DB tests already prove a Postgres started from the same image/credentials works end-to-end, but closing this would make AC-1's evidence literal rather than inferred.
- [ ] `backend/cmd/migrate/main_test.go:TestMigrate_LiveDatabaseZeroMigrationsClean` and `TestMigrate_LiveDatabaseApplyAndRevert` — Both tests run sequentially against the *same* live target database and its `schema_migrations` table (the first assumes a clean/zero state, the second applies then reverts a real migration). This implicitly depends on Go's default in-file test execution order; running with `-shuffle=on`, filtering via `-run` to only the second test, or future reordering could produce flaky, order-dependent failures. Consider a dedicated schema/database per test or explicitly resetting migration state at the start of each.

## Verdict
- **Score:** 93/100 (100 − 1×5 − 2×1) — informational only; see Flag rationale
- **Flag:** 🟡 Reviewer call
- **Notes:** All 5 acceptance criteria are now met with real implementation + test evidence, including genuine live-database integration tests for the connect/migrate happy paths that were entirely missing in iteration 1, and CI now provisions an actual Postgres service so those tests run for real on every push. The three iteration-1 Major findings (silent DB auto-create, duplicate compose files, redundant connections) are all cleanly resolved. However, this iteration introduces one new open 🟡 Major that is not cosmetic: an unrelated `.gitignore` change (bundled into a "chore" commit inside this ticket's branch) now excludes `plan/reviews/` from git, and has already deleted this exact review file and EMPLOYEE36-2's from tracking — a real, demonstrated regression to the team's review-persistence workflow, and scope creep beyond what EMPLOYEE36-3 asked for. Per the rubric, an open Major caps the verdict at Reviewer call regardless of score. Recommend: pull the `.gitignore`/review-file changes out of this branch (or explicitly re-add and commit the review files) before merge; the two 🟢 Minor items are non-blocking polish.

## Re-review Log

### Iteration 2 — 2026-09-15 — sha `3afb2c5`
- **Resolved:** AC-1 (docker-compose deduped + CI config validation + live-DB proof), AC-2 (live connect test), AC-4 (live migrate-up test), AC-5 (live zero-migrations test), CI-breakage Critical (Postgres service added to `backend-ci.yml`), auto-create-database Major, duplicate docker-compose Major, redundant-connections Major, hardcoded-JWT-secret Minor.
- **Still open:** none carried forward.
- **New issues:** `.gitignore`/`plan/reviews/` scope-creep Major (this review file was itself deleted from git tracking by this branch); two Minor polish items (docker-compose `up -d` not literally exercised in CI; live-DB migrate tests share state across a hidden execution-order assumption).
- **Score:** N/A → 93 (hard gate lifted)
- **Verdict:** 🔴 Block → 🟡 Reviewer call
