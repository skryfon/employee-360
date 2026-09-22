# Review: EMPLOYEE36-1 — EPIC-A: Cycle 1 — Project Setup & Scaffolding (Overview)

> Branch: feature/EPIC-A | Last reviewed: 2026-09-17 11:45 | Iteration: 2 | Verdict: 🟡

## Ticket
**Identifier:** EMPLOYEE36-1 (epic-anchor mode — id doesn't carry an `EPIC-` prefix, but the
ticket's own title is "EPIC-A: ... (Overview)" and it has 5 child stories)
**State:** Backlog (state_group: backlog)
**Link:** N/A (not returned by Plane MCP)
**Cycle:** `plan/cycles/cycle-01-project-setup.md` (Cycle 1 — Project Setup & Scaffolding — this
epic *is* Cycle 1)

### Description
Stand up a working, runnable full-stack skeleton so every later cycle starts from a
functioning project instead of an empty repo. Scope guardrail: no application features
this cycle — no auth, no tenant resolution, no entities/migrations beyond the migration
runner itself, no real routes besides `/health`.

Definition of Done: `go run ./cmd/api` starts cleanly; `docker-compose up -d` brings up
Postgres and the app connects; `curl localhost:PORT/api/v1/health` returns 200;
`pnpm --filter admin dev` / `pnpm --filter employee dev` boot; both frontend apps call
the backend health route via `packages/api-client` and render the result — proving
client → API → DB end to end.

### Child Issues
- **A1** — EMPLOYEE36-2 — Backend: project scaffolding & structure
- **A2** — EMPLOYEE36-3 — Backend: config, DB connectivity & migration runner
- **A3** — EMPLOYEE36-4 — Backend: middleware chain, health route & server bootstrap
- **A4** — EMPLOYEE36-5 — Frontend: monorepo & app scaffolding
- **A5** — EMPLOYEE36-6 — Frontend: API client package & env config

### Acceptance Criteria (grouped per child story — all must be `[x]` for the hard gate)

**[EMPLOYEE36-2] A1 — Backend: project scaffolding & structure**
- [x] AC-1: `go build ./...` succeeds
- [x] AC-2: `go vet ./...` passes clean
- [x] AC-3: `make dev` runs without error
- [x] AC-4: `make migrate` runs without error
- [x] AC-5: `make test` runs without error (no-op)
- [x] AC-6: `usecase/interface` / `usecase/implementation` split scaffolded, with `ucshared` in place

**[EMPLOYEE36-3] A2 — Backend: config, DB connectivity & migration runner**
- [x] AC-1: `docker-compose up -d` brings up PostgreSQL successfully
- [x] AC-2: Application connects to PostgreSQL using configured values
- [x] AC-3: Database connection failure produces exit code 1 and a stderr log line containing "database unreachable" + underlying error, no retry loop
- [x] AC-4: `cmd/migrate` connects and executes cleanly against PostgreSQL
- [x] AC-5: Migration runner handles zero migrations without error

**[EMPLOYEE36-4] A3 — Backend: middleware chain, health route & server bootstrap**
- [x] AC-1: `go run ./cmd/api` starts successfully
- [x] AC-2: Base middleware executes in required order: `request_id → logger → cors → recovery`
- [x] AC-3: `GET /api/v1/health` requires no authentication
- [x] AC-4: `curl localhost:PORT/api/v1/health` returns HTTP 200
- [x] AC-5: Health response includes database ping status
- [x] AC-6: Server shuts down gracefully on termination
- [x] AC-7: Auth and tenant-resolution middleware are not introduced in A3

**[EMPLOYEE36-5] A4 — Frontend: monorepo & app scaffolding**
- [x] AC-1: `pnpm install` completes successfully
- [x] AC-2: `pnpm --filter admin dev` boots and renders the Admin placeholder page
- [x] AC-3: `pnpm --filter employee dev` boots and renders the Employee placeholder page
- [x] AC-4: `pnpm build` completes successfully
- [x] AC-5: `pnpm lint` completes successfully
- [x] AC-6: `pnpm typecheck` completes successfully
- [x] AC-7: Tailwind CSS is configured and working in both applications
- [x] AC-8: Both applications are registered correctly within the pnpm workspace
- [x] AC-9: No API wiring or feature code is introduced

**[EMPLOYEE36-6] A5 — Frontend: API client package & env config**
- [x] AC-1: `packages/api-client` builds successfully
- [x] AC-2: Axios instance uses the configured API base URL
- [x] AC-3: JWT interceptor stub is registered
- [x] AC-4: `X-Tenant-ID` interceptor stub is registered
- [x] AC-5: Typed health API call successfully calls `GET /api/v1/health`
- [x] AC-6: Admin application calls the backend health endpoint through `packages/api-client`
- [x] AC-7: Employee application calls the backend health endpoint through `packages/api-client`
- [x] AC-8: Both applications render the backend health response
- [x] AC-9: Backend health response includes database ping status
- [x] AC-10: End-to-end flow works (Admin/Employee → `packages/api-client` → Gin API → Health Handler → PostgreSQL)

**37/37 ACs met across all 5 children. Hard gate: passed.**

## Note on review methodology

Each child story (EMPLOYEE36-2 through -6) already went through its own multi-iteration
review this cycle (`plan/reviews/review-EMPLOYEE36-{3,4,5,6}.md` are present locally;
`review-EMPLOYEE36-2.md` was committed and reached iteration 3/🟢 Merge on a prior PR
(`#3`, merged via `42798d2`) but was later deleted from git tracking — recovered its
content from that merge commit: `git show 42798d2:plan/reviews/review-EMPLOYEE36-2.md`).
Rather than re-deriving every AC from scratch, this epic-anchor review:

1. Pulled each child's final AC state and open findings from its own review file.
2. Independently re-verified the **combined** current state of `feature/EPIC-A`
   (HEAD `253619e`) rather than trusting the individual reviews were still valid:
   ran `make check` (backend `go build`/`go vet`/`go test ./...` — all pass, including
   live-DB tests against a real local Postgres; frontend `pnpm -r lint`/`pnpm -r typecheck`
   — pass) and `make check-clients` (adds `pnpm -r test` — 17/17 passing — and
   `pnpm -r build` — pass) and `make check-structure` (passes) at the current tip.
   Nothing has regressed since the individual reviews.
3. Re-confirmed each child's still-open findings are, in fact, still open on the current
   branch (not silently fixed by a later ticket's commit) — see Findings below.
4. Did **not** re-run AC-10's manual live e2e verification a second time in this pass
   (already done directly in this session for EMPLOYEE36-6's re-review, sha `253619e`,
   with real Postgres/backend/both dev servers and headless-Chromium-rendered DOM —
   see `review-EMPLOYEE36-6.md`).

`origin/main` (`17bfe53`) is 21 commits behind `feature/EPIC-A` — none of this epic's
work has been merged to `main` yet. `origin/feature/EPIC-A` (`312d8dd`) is itself behind
the local `feature/EPIC-A` (`253619e`) — the EMPLOYEE36-6 test-coverage and lockfile
fixes from this session are local-only, not yet pushed.

## Latest commit reviewed
`95b3470` — chore(reviews): stop ignoring plan/reviews and re-commit review handoffs (EMPLOYEE36-1)

## Findings

<!-- Carried forward from each child's own review; only still-open items listed. -->

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] **[EMPLOYEE36-3 / A2]** `.gitignore` (commit `5911325`, "chore: ignore review folder and contents in .gitignore") adds `plan/reviews/` and `reviews/` to `.gitignore`, bundled into a backend-connectivity ticket with no relation to its scope, deleting `review-EMPLOYEE36-2.md`/`-3.md` from tracking. **(resolved in `95b3470`)** — reverted the `.gitignore` exclusion (confirmed: `git check-ignore -v plan/reviews/review-EMPLOYEE36-1.md` now returns "not ignored"), restored `review-EMPLOYEE36-2.md` from its last committed version (`git show 42798d2:...`), and committed all six review files (`review-EMPLOYEE36-{1,2,3,4,5,6}.md`) to `feature/EPIC-A`.
- [ ] **[EMPLOYEE36-5 / A4]** Commit `730e146` ("feat(backend): wire air hot reload into make dev-api", tagged EMPLOYEE36-5) bundles unrelated backend dev-tooling into a frontend-scaffolding ticket. Low risk and additive, not a cycle-scope violation (still Cycle 1 tooling), but makes the ticket harder to review/revert independently. No retroactive action needed — process note for future tickets to keep unrelated-layer changes in their own commit/ticket.

### 🟢 Minor
- [ ] **[EMPLOYEE36-3 / A2]** `.github/workflows/backend-ci.yml`'s "verify docker-compose configurations" step only runs `docker compose config` (syntax/interpolation check), never a literal `up -d` + `pg_isready`/`ps --status running` assertion. Live-DB tests already prove a Postgres from the same image/credentials works end-to-end, so this is inferred rather than literal evidence for AC-1.
- [ ] **[EMPLOYEE36-3 / A2]** `backend/cmd/migrate/main_test.go`'s two live-DB tests (`TestMigrate_LiveDatabaseZeroMigrationsClean`, `TestMigrate_LiveDatabaseApplyAndRevert`) run sequentially against the same live database/`schema_migrations` table, implicitly depending on Go's default in-file execution order. Could flake under `-shuffle=on` or `-run` filtering.
- [ ] **[EMPLOYEE36-4 / A3]** Commit `fcecd17`'s comment-cleanup pass also stripped useful "why" rationale comments (middleware ordering invariants, e.g. why `recovery` must be outermost, why `request_id` must run first) down to one-liners across several files. Purely cosmetic; worth a follow-up pass if general comment cleanup wasn't the specific intent.
- [ ] **[EMPLOYEE36-5 / A4]** `clients/admin/eslint.config.js` / `clients/employee/eslint.config.js` and paired `tsconfig.*.json` files are byte-identical (unmodified `pnpm create vite` template output, duplicated). A later cycle could hoist a shared base config into `packages/`.
- [ ] **[EMPLOYEE36-5 / A4]** Commits `9f8894c` and `15c1402` carry no Plane ticket reference in their subject lines, unlike every other commit on this branch. Process nit only.
- [ ] **[EMPLOYEE36-6 / A5]** AC-10's evidence is a one-time manual sign-off performed directly in this session (real Postgres + backend + both dev servers + headless-Chromium-rendered DOM), not a repeatable CI check — nothing will catch a future regression of the e2e chain automatically. Suggest a fast-follow Playwright CI job.

## Verdict
- **Score:** 89/100 (100 − 1×5 major − 6×1 minor)
- **Flag:** 🟡 Reviewer call
- **Notes:** All 37 acceptance criteria across all 5 child stories (A1–A5) are met with both implementation and test/verification evidence — the hard gate passes cleanly, and this is a genuinely solid Cycle 1 skeleton: `go run ./cmd/api` boots, connects to a real Postgres, exits fast-and-loud on DB failure, `golang-migrate` runs clean against zero migrations, the Gin middleware chain and `/api/v1/health` are fully tested (including DB ping status), both frontend apps boot/lint/typecheck/build with Tailwind working, and the full client→API-client→Gin→handler→Postgres chain was independently, live-verified end to end in this session (real browser, real backend, real DB — not curl, not mocks). I independently re-ran the combined verification suite (`make check`, `make check-structure`, `make check-clients`) at the current tip and everything passes; nothing has regressed across the 5 stacked tickets. The `plan/reviews/` `.gitignore` exclusion (the more serious of the two Major findings — it was silently breaking the team-mate-review skill's persistence model for every review this cycle) is now fixed and all six review files are committed to `feature/EPIC-A`. What keeps this at Reviewer call rather than Merge is one remaining 🟡 Major, purely a commit-hygiene note: an unrelated backend `air` hot-reload change bundled into the A4 (EMPLOYEE36-5) branch (commit `730e146`) — no functional risk, no retroactive fix expected, just flagged per the rubric's "any major open caps at Reviewer call" rule. Six 🟢 Minor items (listed above) are non-blocking polish/process notes.

## Re-review Log

### Iteration 2 — 2026-09-17 — sha `95b3470`
- **Resolved:** 🟡 Major — `plan/reviews/` `.gitignore` exclusion reverted; `review-EMPLOYEE36-2.md` restored from `42798d2`; all six review files (`-1` through `-6`) committed to `feature/EPIC-A`.
- **Still open:** 🟡 Major — unrelated backend `air` hot-reload commit bundled into A4/EMPLOYEE36-5 (`730e146`, no action needed retroactively); all 6 🟢 Minor items unchanged.
- **New issues:** none
- **Score:** 84 → 89 (+5)
- **Verdict:** 🟡 Reviewer call → 🟡 Reviewer call
