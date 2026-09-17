# Review: EMPLOYEE36-6 — A5 — Frontend: API client package & env config

> Branch: ebin/feat/EPIC-A/EMPLOYEE36-6 | Last reviewed: 2026-09-17 11:35 | Iteration: 4 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-6
**State:** In Progress (state_group: started)
**Link:** N/A (not returned by Plane MCP)
**Cycle:** `plan/cycles/cycle-01-project-setup.md` (Cycle 1 — Project Setup & Scaffolding)

### Description
Implement the shared `packages/api-client` package and connect both frontend
applications to the backend health endpoint. The client should provide the
foundation for future authenticated/multi-tenant API communication, while A5 uses
only the public health endpoint as the first end-to-end smoke test. Depends on A3
(EMPLOYEE36-4) and A4 (EMPLOYEE36-5). No auth/tenant-resolution logic, no feature
APIs in this ticket — JWT and `X-Tenant-ID` interceptors are stubs only.

### Acceptance Criteria
- [x] AC-1: `packages/api-client` builds successfully.
- [x] AC-2: Axios instance uses the configured API base URL.
- [x] AC-3: JWT interceptor stub is registered.
- [x] AC-4: `X-Tenant-ID` interceptor stub is registered.
- [x] AC-5: Typed health API call successfully calls `GET /api/v1/health`.
- [x] AC-6: Admin application calls the backend health endpoint through `packages/api-client`.
- [x] AC-7: Employee application calls the backend health endpoint through `packages/api-client`.
- [x] AC-8: Both applications render the backend health response.
- [x] AC-9: Backend health response includes database ping status.
- [x] AC-10: End-to-end flow works (Admin/Employee → `packages/api-client` → Gin API → Health Handler → PostgreSQL).

## Note on review base

This branch stacks EMPLOYEE36-2 through EMPLOYEE36-5, each already reviewed at its
own tip commit (`plan/reviews/review-EMPLOYEE36-{3,4,5}.md`, all 🟢 Merge).
`origin/main` (`17bfe53`) hasn't caught up with any of those merges yet, so diffing
straight against it would re-surface ~8.4k lines of already-reviewed work. This
review instead diffs `49c2a16` (tip of EMPLOYEE36-5) to `HEAD` — i.e. only the
single commit tagged `(EMPLOYEE36-6)`.

I verified directly rather than trusting the diff alone: ran `go build ./...`,
`go vet ./...` and `go test ./...` on the backend (all pass — the `health_handler.go`
type rename doesn't break the existing DB-ping test coverage from EMPLOYEE36-4);
ran `pnpm -r build`, `pnpm -r lint`, `pnpm -r typecheck` on the frontend workspace
(all pass, including the new `packages/api-client`); regenerated the Orval client
from `backend/docs/swagger.json` and confirmed zero drift from the committed
`src/generated/` output; and ran `pnpm -r test` directly, which exits 0 having run
**zero** tests (no package in the workspace defines a `test` script — see 🟡 Major
below). I attempted to verify the end-to-end chain (AC-10) live by starting
Postgres, but Docker isn't available in this environment and the only reachable
Postgres on `localhost:5432` rejected the project's default dev credentials
(`postgres`/`postgres`), so AC-10 could not be confirmed live either.

## Latest commit reviewed
`253619e` — fix(frontend): regenerate lockfile for admin/employee test deps (EMPLOYEE36-6)

## Findings

### 🔴 Critical
<!-- Each entry: file:line — what — why — suggested fix -->
- [x] AC-2 unmet — `packages/api-client/src/client.ts:313-327` (`configureApiClient`) sets `apiClient.defaults.baseURL`, but no test asserts that calling it actually changes the base URL used by outgoing requests. **(resolved in 2281bac)** — `packages/api-client/src/client.test.ts:15-34` now asserts `apiClient.defaults.baseURL` updates on a valid call and stays untouched on an empty string; ran `pnpm -r test` and confirmed it passes.
- [x] AC-3 unmet — `packages/api-client/src/client.ts:339-342` (JWT interceptor stub) has no test proving the interceptor is registered or a no-op today. **(resolved in 2281bac)** — `client.test.ts:47-61` asserts ≥2 request interceptor handlers are registered and that no `Authorization` header is attached (proving today's no-op stub). Verified passing.
- [x] AC-4 unmet — `packages/api-client/src/client.ts:353-356` (`X-Tenant-ID` interceptor stub), same gap as AC-3. **(resolved in 2281bac)** — `client.test.ts:63-70` mirrors the AC-3 test for the `X-Tenant-ID` header. Verified passing.
- [x] AC-5 unmet — `packages/api-client/src/health.ts` (`fetchHealth`/`useHealth`) had no test proving the typed call hits `GET /api/v1/health` or that malformed envelopes are rejected. **(resolved in 2281bac)** — `packages/api-client/src/health.test.ts` covers `fetchHealth` (success unwrap, missing-`data` rejection, wrong-type-field rejection) and `useHealth` (success settle, error surface) via `axios-mock-adapter` + `renderHook`. Verified passing.
- [x] AC-6 unmet — `clients/admin/src/App.tsx` calls `useHealth()` with no automated coverage; the curl-based CI smoke test never executes client JS. **(resolved in 2281bac)** — `clients/admin/src/App.test.tsx` renders `<App>` against a mocked `apiClient` and asserts a real `GET /api/v1/health` request is made (`mock.history.get`) and the response is rendered. Verified passing.
- [x] AC-7 unmet — identical gap for Employee. **(resolved in 2281bac)** — `clients/employee/src/App.test.tsx`, same approach mirrored for Employee. Verified passing.
- [x] AC-8 unmet — none of loading/error/success had a DOM assertion. **(resolved in 2281bac)** — both `App.test.tsx` files now assert all three states: `"Checking API health…"` while pending, `/API health check failed/i` on a mocked 500, and the rendered `status`/`app`/`database` values on success. Verified passing.
- [x] AC-10 unmet — no automated or live-verified test proves the full chain (Admin/Employee → `packages/api-client` → Gin API → Health Handler → PostgreSQL) actually works end to end. **(resolved — manual sign-off, 2026-09-17 11:35, sha `253619e`)** — live-verified in this environment (Postgres became reachable via the project's own `.env` credentials, unlike iteration 1's attempt): started Postgres (already running on `localhost:5432`, confirmed via `psql` with the root `.env` `DATABASE_*` creds), ran `go run ./cmd/api` from `backend/` (log: `"connected to database"`, then `GET /api/v1/health` → `200`), and started both `pnpm --filter admin dev --port 5183` and `pnpm --filter employee dev --port 5184` (both already point at `http://localhost:8080` via their own `.env`'s `VITE_API_BASE_URL`). Rendered each app with a real headless Chromium (`chromium-browser --headless=new --dump-dom --virtual-time-budget=8000`, executing actual client JS, not curl) and confirmed the DOM: Admin shows `Status: ok`, `App: employee360`, `Database: ok`; Employee shows the same three values under its own heading. Backend request log shows two distinct `GET /api/v1/health` 200s beyond my own manual curl, one per app. This is a one-time manual verification (no CI job asserts this on every push — see the still-open suggestion below), all dev processes were stopped afterward. Recommend still filing a fast-follow ticket for a real Playwright e2e CI job so this doesn't silently regress; not a blocker for this PR.
- [x] `pnpm-lock.yaml` was not regenerated for `clients/admin` and `clients/employee`'s newly-added devDependencies (`vitest`, `jsdom`, `@testing-library/*`, `axios-mock-adapter`), breaking `pnpm install --frozen-lockfile` in CI (`.github/workflows/frontend-ci.yml:36`). **(resolved in 253619e)** — ran `pnpm install` to regenerate the lockfile (all three importer blocks now list the new deps), then verified from a clean `node_modules` that both `pnpm install --frozen-lockfile` and `make check-clients` (the exact CI job: lint, typecheck, test, build) pass end to end.

### 🟡 Major
- [x] No frontend package defined a `test` script, so `pnpm -r test` silently ran zero tests. **(resolved in 2281bac)** — `clients/admin/package.json`, `clients/employee/package.json`, and `packages/api-client/package.json` now all define `"test": "vitest run"`, and since `make check-clients` already chained `test-clients` → `pnpm -r test` (from EMPLOYEE36-5), CI's existing job now runs real tests as soon as the lockfile issue above is fixed.
- [x] `.github/workflows/frontend-ci.yml:41-55`'s curl-based smoke test never executes client JS, so it couldn't serve as evidence for AC-6/7/8. **(resolved in 2281bac)** — the underlying concern (no automated coverage that actually executes the client-side wiring) is now addressed by the new RTL component tests, which do execute the real `App` component and assert on the actual HTTP call. The curl step itself is unchanged and still isn't proof of a real end-to-end run (see AC-10, still open), but it's no longer the *only* thing resembling coverage.

### 🟢 Minor
- [ ] AC-10's evidence is a one-time manual sign-off (this review, iteration 4), not a repeatable CI check — nothing will catch a future regression of the e2e chain automatically. Suggest filing a fast-follow ticket for a Playwright CI job (boot `docker-compose` Postgres + `make dev-api` + both Vite dev servers, drive a real browser, assert rendered health data) so this becomes self-verifying. Non-blocking for this PR.

## Verdict
- **Score:** 99/100 (100 − 1 for the one 🟢 Minor)
- **Flag:** 🟢 Merge
- **Notes:** All 10 ACs are now met with real evidence, and the hard gate clears. AC-2 through AC-8 have genuine Vitest/RTL tests (mocking at the HTTP-transport layer, not the units under test) that I ran and confirmed passing (17/17). The `pnpm-lock.yaml` drift that broke CI's `pnpm install --frozen-lockfile` step is fixed and verified from a clean install (`make check-clients` passes end to end). AC-10 is now backed by a live manual verification I performed directly in this session: real Postgres, real `go run ./cmd/api`, real Vite dev servers for both apps, and real headless-Chromium-rendered DOM (not curl) showing `status: ok`, `app: employee360`, `database: ok` on both Admin and Employee, with the backend's own request log confirming two distinct calls beyond my own probe. The only open item is a Minor suggestion to turn that one-time manual check into a repeatable CI job — non-blocking, doesn't affect the verdict. Clear to merge.

## Re-review Log

### Iteration 2 — 2026-09-17 — sha `2281bac`
- **Resolved:** AC-2, AC-3, AC-4, AC-5, AC-6, AC-7, AC-8 (all via new Vitest/RTL tests, verified passing locally); both prior 🟡 Major findings (missing test scripts; curl smoke test not proving wiring).
- **Still open:** AC-10 (no e2e coverage or documented manual sign-off).
- **New issues:** 🔴 `pnpm-lock.yaml` not regenerated for `clients/admin`/`clients/employee` new devDependencies — breaks `pnpm install --frozen-lockfile` in CI (frontend-ci.yml:36), confirmed locally.
- **Score:** N/A → N/A (hard gate both iterations)
- **Verdict:** 🔴 Block → 🔴 Block

### Iteration 3 — 2026-09-17 — sha `253619e`
- **Resolved:** `pnpm-lock.yaml` drift (regenerated; `pnpm install --frozen-lockfile` and `make check-clients` verified passing from a clean install).
- **Still open:** AC-10 (no e2e coverage or documented manual sign-off) — the sole remaining blocker.
- **New issues:** none
- **Score:** N/A → N/A (hard gate persists on AC-10)
- **Verdict:** 🔴 Block → 🔴 Block

### Iteration 4 — 2026-09-17 — sha `253619e` (no new commit; manual verification only)
- **Resolved:** AC-10 — live manual verification performed in this session (real Postgres + backend + both Vite apps + headless-Chromium-rendered DOM), superseding iteration 1's failed attempt (Docker unavailable, default creds rejected — this time the project's own `.env` creds worked and Postgres was already reachable).
- **Still open:** none (AC checklist)
- **New issues:** 🟢 Minor — AC-10's evidence is manual, not CI-enforced; suggest a fast-follow Playwright job.
- **Score:** N/A → 99/100
- **Verdict:** 🔴 Block → 🟢 Merge
