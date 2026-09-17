# Review: EMPLOYEE36-4 — A3 — Backend: middleware chain, health route & server bootstrap

> Branch: ebin/feat/EPIC-A/EMPLOYEE36-4 | Last reviewed: 2026-09-15 16:14 | Iteration: 2 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-4
**State:** In Progress (state_group: started)
**Link:** N/A (not returned by Plane MCP)
**Cycle:** `plan/cycles/cycle-01-project-setup.md` (Cycle 1 — Project Setup & Scaffolding)

### Description
Set up the Gin HTTP server foundation with the base middleware chain, health endpoint,
server boot, and graceful shutdown.

Auth and tenant-resolution middleware are out of scope for this ticket and will be
added in Cycle 2 once there are protected resources.

- Middleware chain, registered in this exact order: `request_id → logger → cors → recovery`
- Health endpoint `GET /api/v1/health`: no auth required, returns HTTP 200, includes
  application health status and PostgreSQL/database ping status
- `cmd/api/main.go`: load config, initialize DB connection, wire dependencies, create
  and configure the Gin server, register middleware and routes, start the HTTP server,
  handle graceful shutdown on termination signals
- Depends on A2 (EMPLOYEE36-2)

### Acceptance Criteria
- [x] AC-1: `go run ./cmd/api` starts successfully.
- [x] AC-2: Base middleware executes in the required order: `request_id → logger → cors → recovery` (resolved in `fcecd17`)
- [x] AC-3: `GET /api/v1/health` requires no authentication. (resolved in `fcecd17`)
- [x] AC-4: `curl localhost:PORT/api/v1/health` returns HTTP 200. (resolved in `fcecd17`)
- [x] AC-5: Health response includes database ping status. (resolved in `fcecd17`)
- [x] AC-6: Server shuts down gracefully on termination.
- [x] AC-7: Auth and tenant-resolution middleware are not introduced in A3.

## Note on review base

`origin/main` on GitHub (`17bfe53`) has not yet caught up with PR #3 (EMPLOYEE36-2)
or EMPLOYEE36-3, both of which are already merged into this branch's own history
(commits `42798d2`..`3afb2c5`) and already have their own reviews
(`plan/reviews/review-EMPLOYEE36-2.md` — Merge, `plan/reviews/review-EMPLOYEE36-3.md`
— Reviewer call). Diffing straight against `origin/main` would re-surface that
already-reviewed work, so this review instead diffs `3afb2c5..HEAD` — i.e. only the
commit added for EMPLOYEE36-4 on top of the EMPLOYEE36-3 base.

## Latest commit reviewed
`fcecd17` — feat(backend): add unversioned health routes, clean logs, and update root Makefile (EMPLOYEE36-4)

## Findings

### 🔴 Critical

- [x] `backend/internal/delivery/http/routes.go` — AC-3/AC-4 unmet: health route was registered at `/health`, not `/api/v1/health`. **(resolved in `fcecd17`)** `registerRoutes` now mounts a versioned route under `shared.APIVersionPrefix` (`v1.GET("/health", ...)`) alongside unversioned `/health` and `/healthz` probes. `main_test.go`'s `healthURL` and `routes_test.go`/`container_test.go` both assert against `/api/v1/health`.
- [x] `backend/internal/delivery/http/handlers/health_handler.go` — AC-5 unmet: health response omitted database ping status. **(resolved in `fcecd17`)** New `service.DatabasePinger` port + `database.GormDatabasePinger` adapter + `usecaseimpl.HealthUseCase` cleanly thread a DB ping through domain → usecase → handler (proper layering, no GORM in the handler). Response now includes `"database": "ok"|"unreachable"`. `main_test.go:TestAPI_LiveDatabaseSuccess` was inverted to assert `"database":"ok"` against a live Postgres (previously it asserted the field's absence); `health_handler_test.go` and `health_usecase_test.go` cover both the healthy and unreachable-DB paths with fakes.
- [x] `backend/internal/delivery/http/middleware/{request_id,logger,cors,recovery}.go`, `routes.go` — AC-2 unmet: no test evidence for middleware order/behavior. **(resolved in `fcecd17`)** New `request_id_test.go` (generation + client-supplied propagation), `logger_test.go` (request_id present in log line, proving order after RequestID), `cors_test.go` (allowlist, wildcard, disallowed-origin, preflight), `recovery_test.go` (panic → 500 envelope, log contains stack info, normal requests unaffected), and `routes_test.go` (full chain wired via `NewRouter`, asserting `X-Request-ID` header + CORS header both present on the real health route) together demonstrate the chain executes in the required order with the required behavior.

### 🟡 Major

- [x] `backend/internal/delivery/http/middleware/cors.go` — Reflected CORS origin combined with `Allow-Credentials: true`, no allowlist. **(resolved in `fcecd17`)** `CORS` now takes `allowedOrigins []string` (from new `config.CORSConfig`, Viper-backed via `cors.allowed_origins` / `CORS_ALLOWED_ORIGINS`). Wildcard (`*`) mode never sets `Allow-Credentials`; a configured allowlist only reflects+credentials an origin that's actually on the list, with `Vary: Origin` set, and disallowed origins get no ACAO header at all. Covered by `cors_test.go`'s four scenarios and `config_test.go:TestConfig_CORSAllowedOriginsEnvOverride`.
- [x] No unit tests were added for any of the five new/changed packages. **(resolved in `fcecd17`)** Every touched package now has a `_test.go`: handlers, all four middleware, routes, container, server, plus the new health usecase/domain-service/infrastructure-pinger trio.

### 🟢 Minor

- [x] `backend/internal/delivery/http/handlers/health_handler.go:11` doc comment claiming no DB dependency. **(resolved in `fcecd17`)** — handler rewritten to take a `usecaseinterface.HealthUseCase`; stale comment is gone.
- [x] Cycle doc ambiguity between `/api/v1/health` and `/healthz`. **(resolved in `fcecd17`)** — both are now served (plus bare `/health`), so either reading of the cycle doc is satisfied.
- [ ] **New:** This commit also strips detailed rationale comments down to one-liners across `ctx.go`, `logger.go`, `request_id.go`, `logger.go` (middleware), `recovery.go`, `server.go`, `pkg/logger/logger.go`, and `bootstrap/main.go` (e.g. `recovery.go` no longer states *why* it must be outermost in the chain; `request_id.go` no longer states *why* it must run first). These comments previously documented real, non-obvious invariants (middleware ordering constraints, "never trust a client-supplied tenant id") that aren't otherwise obvious from the code — the project's own comment guidance ("only add a comment when the WHY is non-obvious") argues for keeping that class of comment rather than trimming it to a WHAT-only one-liner. Purely cosmetic, no functional risk; worth a follow-up pass if the intent was general comment cleanup rather than removing invariant documentation.

## Verdict
- **Score:** 99/100 (100 − 1×1)
- **Flag:** 🟢 Merge
- **Notes:** All seven acceptance criteria are now met with both implementation and test evidence. The health route is correctly versioned at `/api/v1/health` (with `/health` and `/healthz` also served for infra probes), the response now reports live database ping status via a cleanly layered domain-service/usecase/handler chain, and every middleware in the required chain (`request_id → logger → cors → recovery`) has dedicated unit tests plus an integration-level test through the real `NewRouter`. The reflected-CORS-with-credentials issue is fixed with a proper Viper-backed allowlist and is well tested. `go build ./...`, `go vet ./...`, and `go test ./...` all pass locally (the one live-DB test skips locally due to no reachable Postgres, as expected outside CI). The only remaining note is a very low-severity comment-quality regression (see 🟢 Minor) — some previously-useful "why" comments about middleware ordering invariants were trimmed to one-liners in this same commit's cleanup pass. Clear to merge.

## Re-review Log

### Iteration 2 — 2026-09-15 — sha `fcecd17`
- **Resolved:** AC-2 (middleware order test evidence), AC-3/AC-4 (route mounted at `/api/v1/health`), AC-5 (DB ping status in response), reflected-CORS-with-credentials (Major), missing unit tests across all touched packages (Major), stale "no DB dependency" doc comment (Minor), cycle-doc route-path ambiguity (Minor)
- **Still open:** none
- **New issues:** one new 🟢 Minor — this commit's comment-cleanup pass also stripped useful "why" rationale comments (middleware ordering invariants) down to one-liners in several files
- **Score:** 28 → 99 (+71)
- **Verdict:** 🔴 Block → 🟢 Merge
