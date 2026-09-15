# Review: EMPLOYEE36-3 — A2 — Backend: config, DB connectivity & migration runner

> Branch: ebin/feat/EPIC-A/EMPLOYEE36-3 | Last reviewed: 2026-09-15 12:26 | Iteration: 1 | Verdict: 🔴

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
- [ ] AC-1: `docker-compose up -d` brings up PostgreSQL successfully
- [ ] AC-2: Application connects to PostgreSQL using configured values
- [x] AC-3: Database connection failure produces exit code 1 and a stderr log line
      containing "database unreachable" and the underlying error, within a bounded
      startup timeout (no retry loop)
- [ ] AC-4: `cmd/migrate` connects and executes cleanly against PostgreSQL
- [ ] AC-5: Migration runner handles zero migrations without error

## Note on review base

`origin/main` on GitHub (`17bfe53`) has not yet caught up with PR #3
(EMPLOYEE36-2), which is already merged into this branch's own history via commit
`42798d2` and already has its own review (`plan/reviews/review-EMPLOYEE36-2.md`,
verdict Merge). Diffing straight against `origin/main` would re-surface that
already-reviewed work, so this review instead diffs `42798d2..HEAD` — i.e. only the
commit(s) added for EMPLOYEE36-3 on top of the merged EMPLOYEE36-2 base.

## Latest commit reviewed
`29fba54` — feat(backend): configure viper, postgres gorm connectivity, and golang-migrate runner (EMPLOYEE36-3)

## Findings

### 🔴 Critical

- [ ] `backend/cmd/api/main.go`, `backend/internal/infrastructure/database/postgres_test.go` — **AC-2 unmet: no automated test exercises a successful connection to a running PostgreSQL.** Every new test (`backend/cmd/api/main_test.go:TestAPI_DatabaseUnreachableFailsFast`, `backend/cmd/migrate/main_test.go:TestMigrate_DatabaseUnreachableFailsFast`, `backend/internal/infrastructure/database/postgres_test.go:TestConnect_UnreachableDatabaseFailsFast` and `TestMustConnect_ExitsNonZeroWithStderr`) only exercises the *failure* path (unreachable host). Nothing asserts that `Connect`/`EnsureDatabaseExists` actually succeeds against a real PostgreSQL with configured values. Add an integration test (gated behind a `TEST_DATABASE_*`/build-tag or run against the docker-compose Postgres) that connects successfully and pings.
- [ ] `backend/cmd/migrate/main.go` — **AC-4 and AC-5 unmet: no test runs `migrate up`/`status` against a real PostgreSQL.** `cmd/migrate/main_test.go` only covers `TestMigrate_DatabaseUnreachableFailsFast` (failure path), `TestFindMigrationsDir`, and `TestRunCreate_GeneratesMigrationPair` (pure filesystem, no DB). Neither "connects and executes cleanly" nor "handles zero migrations without error" (`runUp`'s `isNoChangeOrEmpty` branch) is exercised against a live database anywhere.
- [ ] `docker-compose.yml`, `backend/docker-compose.yml` — **AC-1 unmet: nothing verifies `docker-compose up -d` actually brings Postgres up successfully.** No CI step or test runs `docker compose up -d --wait` (or equivalent `pg_isready` check) against either compose file.
- [ ] `.github/workflows/backend-ci.yml:24-32` (unchanged by this branch) combined with `backend/cmd/api/main.go` / `backend/cmd/migrate/main.go` — **This branch will break CI.** The workflow's "smoke-test dev and migrate entrypoints" step still carries its Cycle-1 comment "cmd/api and cmd/migrate are currently no-op stubs... A2 will similarly need a Postgres service added here once cmd/migrate does real work" — but this ticket (A2) makes `cmd/api` and `cmd/migrate` perform real DB connections, and the workflow was **not** updated to add a `services: postgres:` block. On `ubuntu-latest` there is no reachable Postgres at `localhost:5432`, so `database.EnsureDatabaseExists`/`Connect` will fail and `database.FailFast` calls `os.Exit(1)` — `timeout 10s go run ./cmd/api` and `go run ./cmd/migrate up` will both exit non-zero, failing the CI job on every push to this branch. Add the missing Postgres service (with matching `DATABASE_*` env vars) to `backend-ci.yml` before merging.

### 🟡 Major

- [ ] `backend/internal/infrastructure/database/postgres.go:~1205` (`EnsureDatabaseExists`) — Silently auto-creates the configured database (`CREATE DATABASE "<name>"` against the `postgres` maintenance DB) when it doesn't exist, instead of failing fast. This isn't requested anywhere in the ticket/AC (which only specifies fail-fast, "no retry loop; no partial startup") and is a behavior change that can mask a misconfigured `DATABASE_NAME` — a typo silently spins up a fresh empty database instead of erroring loudly, which is exactly the kind of "partial startup" the AC says to avoid. Recommend dropping the auto-create for this cycle (migrations/seeding own database provisioning in later cycles) or gating it behind an explicit opt-in config flag.
- [ ] `docker-compose.yml` and `backend/docker-compose.yml` — Byte-identical duplicate files. They will silently drift out of sync the first time either is edited alone. Pick one canonical location and remove the other (or symlink).
- [ ] `backend/cmd/api/main.go:143-151`, `backend/cmd/migrate/main.go:269-281` — Every invocation opens two full GORM connections back-to-back: `EnsureDatabaseExists` connects, pings, and closes one connection just to check existence, then the caller immediately opens a second one via `database.Connect`. Not incorrect, but doubles connection/ping latency on the happy path (and `cmd/migrate` opens a third via `golang-migrate`'s own postgres driver). Consider having `EnsureDatabaseExists` return the already-open `*gorm.DB` for reuse instead of closing and reopening.

### 🟢 Minor

- [ ] `backend/config/config.go:171` — Hardcoded JWT default secret (`"employee360-super-secret-key-change-in-production"`) as a Viper default. Fine for local dev (overridable via `JWT_SECRET`), but nothing yet warns/fails if it's still the default in a non-development `app.environment` — worth a follow-up ticket rather than blocking this one.

## Verdict
- **Score:** N/A — hard gate triggered (see below)
- **Flag:** 🔴 Block
- **Notes:** Four of five acceptance criteria (AC-1, AC-2, AC-4, AC-5) have no automated test evidence for their actual success-path behavior; every new test in this diff exercises the *unreachable-database* failure path only. Per the review process, any unmet AC is an automatic block regardless of score. Compounding this, the one place that could have exercised the happy path — the CI workflow's smoke-test step — will now actively fail on this branch because `.github/workflows/backend-ci.yml` was not updated to run a Postgres service, despite its own inline comment flagging that exact requirement for this ticket. Recommend: add a docker-compose/Postgres service to CI, add at least one integration test per AC-2/AC-4/AC-5 that runs against a real (ephemeral) Postgres, and reconsider the silent auto-create-database behavior in `EnsureDatabaseExists` before merging.

## Re-review Log
_(empty — first review)_
