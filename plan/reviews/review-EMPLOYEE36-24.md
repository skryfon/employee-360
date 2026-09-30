# Review: EMPLOYEE36-24 — B8 — Async email delivery: domain events, EventPublisher/outbox (River), worker

> Branch: ebin/feat/EPIC-B/EMPLOYEE36-24 | Last reviewed: 2026-09-30 | Iteration: 3 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-24
**State:** started (In progress)
**Link:** n/a (not returned by Plane)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
Makes email delivery asynchronous via event-publish + transactional outbox on River (same Postgres; `river_job` is the outbox). Usecase runs inside `Transactor.WithinTransaction`, calls `EventPublisher.Publish` in the same tx; infra extracts `*sql.Tx` from ctx and calls `InsertTx`. A separate `cmd/worker` drains jobs and calls `EmailService.Send`. Scope: transactor, dispatcher, river_publisher, email_worker, cmd/worker, migration 000012, docker-compose `worker` service, `make worker`.

### Acceptance Criteria
- [x] AC-1: Rolling back an outer usecase transaction leaves neither the business row nor the River job behind — impl `database/transactor.go:28`, `eventing/river_publisher.go:37`; test `eventing/outbox_integration_test.go` `TestOutbox_RollbackLeavesNeitherRowNorJob` (build tag `integration`)
- [x] AC-2: No usecase imports River or `*sql.Tx`/`*gorm.DB` — impl holds; enforced by `internal/archtest/architecture_test.go` `TestUsecaseLayerImportsNoInfrastructure`
- [x] AC-3: `EmailService.Send` only called from `infrastructure/job/email_worker.go` — impl `email_worker.go:29`; enforced by `archtest` `TestEmailServiceSendOnlyCalledFromWorker`
- [x] AC-4: With api + worker running, invite/reset yields an email after the API response — `eventing/worker_integration_test.go` `TestWorker_DeliversAfterCommitOnly` (nothing delivered before commit, exactly one after, via the worker client). Manual Mailpit check still to be noted in the PR
- [x] AC-5: `make migrate` applies through 000012; `make migrate-down` reverses — `integration/migrate_test.go` `TestMigrate_LiveDatabaseRealMigrationsApplyAndRevert` applies all real migrations and reverts them all; now runs on an isolated scratch DB

## Latest commit reviewed
`375f6ffebe05c8d6c64e1e0babb6444af4118236` — feat(backend): implement River background worker and transactional eventing (EMPLOYEE36-24)

> Iterations 2–3 reviewed the **uncommitted working tree** on top of this sha (HEAD had not advanced). Commit the fixes; this sha is then stale.

(Scope reviewed: commit 375f6ff only. The branch is stacked on earlier EMPLOYEE36-9…15 work, which has its own review files.)

## Findings

### 🔴 Critical
- [x] (resolved in working tree on 375f6ff) AC-2 / AC-3 unmet (no test evidence) — no test enforces "no usecase imports river/database/sql/gorm" or "`EmailService.Send` only called from `email_worker.go`" — the invariants are true today but regress silently when B4/C1 usecases land — fix: add a small arch test (e.g. `go list -deps ./internal/usecase/...` / `go/parser` import scan, plus a grep-style scan for `.Send(` on `EmailService`) run in `make test`.
- [x] (resolved in working tree on 375f6ff) AC-4 unmet (no test evidence) — `job/email_worker_test.go` uses a fake; nothing proves a job inserted via `RiverPublisher` is picked up by `NewWorkerClient` and delivered (e.g. to a fake `EmailService`) — fix: add an integration test (tag `integration`) that publishes in a tx, starts the worker client with a recording `EmailService`, and asserts delivery after commit; record the manual Mailpit check in the PR.
- [x] (resolved in working tree on 375f6ff) AC-5 unmet (no test evidence) — no up/down verification for 000012 in this commit — fix: run `make migrate && make migrate-down && make migrate` and note the result in the PR, or add a migration up/down integration test.

### 🟡 Major
- [x] (resolved in working tree on 375f6ff) `internal/usecase/implementation/auth/forgot_password.go:119` — hard-coded production URL `https://app.skryfon.com/reset-password?token=` in a usecase — breaks self-hostable/open-source principle and makes every non-Skryfon deployment email a wrong link — fix: inject a frontend base URL from config (Viper) into the usecase and build the URL from it; add a test asserting it.
- [x] (resolved in working tree; not run, docker unavailable) `backend/docker-compose.yml` worker service — `volumes: ./:/app` mounts `backend/` at `/app`, but `working_dir: /app/backend` doesn't exist (go.mod is at `backend/go.mod`) so `go run ./cmd/worker` will fail — fix: `working_dir: /app` (or mount `../` at `/app`). Please verify with `docker compose up worker`.
- [x] (resolved in working tree; residual: `reset_url`/`invite_url` still embed the token for up to 1h after completion) `eventing/dispatcher.go:37,52` / `river_job.args` — plaintext reset/invite tokens (`plain_token`, and the URL containing it) are persisted in `river_job.args`, retained after completion — at-rest exposure that the hashed `password_reset_tokens` design was avoiding — fix: drop `plain_token` from `template_data` (the URL is enough) and configure River `CompletedJobRetentionPeriod` / periodic cleanup to be short; or document the accepted trade-off.

### 🟢 Minor
- [x] (resolved in working tree) `backend/internal/infrastructure/persistence/testutil_test.go` — comment references the removed `checkLiveDBOrSkip`; update it.
- [ ] `internal/archtest/architecture_test.go` — AC-3 guard is an identifier-name heuristic (no type checker); an `EmailService` held under another or embedded name would evade it. Acceptable.
- [x] (resolved in working tree) `container.go` `NewWorkerContainer(cfg, db, log)` — `log` param is now unused; drop it or use it.
- [x] (resolved in working tree) `config.go` `app.frontend_url` defaults to `http://localhost:5173`; unset in production it silently emails localhost links — consider failing fast when `app.environment != development`.
- [x] (resolved in working tree on 375f6ff) `container/container.go:5-19` — import groups are interleaved (infraservice placed among stdlib/third-party); run goimports grouping.
- [x] (resolved in working tree; logger unification deliberately left) `container.go` `WorkerContainer` — `Config`, `Log`, `EmailSvc` fields unused by `cmd/worker`; worker uses a separate `slog` logger alongside zerolog — consider a single logging path.
- [x] (accepted, unchanged) `outbox_integration_test.go` — integration tests silently `Skip` locally when Postgres is down; fine, but note AC-1 evidence only runs under `-tags integration`.

## Verdict
- **Score:** 99/100 (100 − 1 open 🟢)
- **Flag:** 🟢 Merge
- **Notes:** All five ACs now have implementation and test evidence, and all prior Critical/Major findings are fixed. Iteration 3 fixed three of the four 🟢 nits; only the accepted archtest heuristic remains. `go test ./...` passes after iteration 3; the `-tags integration` suite was not re-run. Caveats: the fixes are uncommitted (commit before merging), docker-compose was not exercised (no Docker here), and the manual Mailpit check is still to be recorded in the PR. Reset/invite URLs still carry the token in `river_job.args` for up to an hour after completion (accepted trade-off).

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `375f6ff` + uncommitted working tree
- **Resolved:** AC-2/3 guard test; AC-4 worker integration test; AC-5 migrate evidence (tests isolated on scratch DB); hard-coded reset URL now `APP_FRONTEND_URL`; compose `working_dir`; `plain_token` removed from job args + 1h completed-job retention; import grouping; unused `WorkerContainer` fields
- **Still open:** no Critical or Major
- **New issues:** 4 🟢 (stale comment in testutil_test.go, archtest heuristic, unused `log` param, localhost default for frontend URL)
- **Score:** n/a (AC gate, Block) → 96/100
- **Verdict:** 🟢 Merge

### Iteration 3 — 2026-09-30 — uncommitted working tree
- **Resolved:** stale `checkLiveDBOrSkip` comment in `testutil_test.go`; unused `log` param dropped from `NewWorkerContainer` (and `cmd/worker/main.go`); `Config.Validate()` now fails in production/staging when `app.frontend_url` is unset or the localhost default (`DefaultFrontendURL`), with the two affected config tests updated to set `APP_FRONTEND_URL`
- **Still open:** archtest AC-3 identifier-name heuristic (accepted, 🟢)
- **New issues:** none
- **Note:** deployments without `APP_FRONTEND_URL` in production/staging now fail at startup
- **Score:** 96 → 99 (+3)
- **Verdict:** 🟢 Merge
