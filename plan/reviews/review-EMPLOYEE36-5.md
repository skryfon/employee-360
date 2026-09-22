# Review: EMPLOYEE36-5 — A4 — Frontend: monorepo & app scaffolding

> Branch: ebin/feat/EPIC-A/EMPLOYEE36-5 | Last reviewed: 2026-09-15 19:45 | Iteration: 2 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-5
**State:** In Progress (state_group: started)
**Link:** N/A (not returned by Plane MCP)
**Cycle:** `plan/cycles/cycle-01-project-setup.md` (Cycle 1 — Project Setup & Scaffolding)

### Description
Set up the frontend monorepo structure for Employee360 using pnpm workspaces. Scaffold
the Admin and Employee applications as Vite + React + TypeScript applications with
Tailwind CSS configured and a basic placeholder page rendered.

Scaffolding only — no feature implementation, authentication, API client, React Query,
Zustand stores, or backend/API wiring. `packages/api-client` and `packages/ui` are
placeholder directories only; `packages/api-client` implementation is explicitly
deferred to A5 (EMPLOYEE36-6). Depends on A3 (EMPLOYEE36-4).

### Acceptance Criteria
- [x] AC-1: `pnpm install` completes successfully. — verified directly (`pnpm install` → "Already up to date", exit 0).
- [x] AC-2: `pnpm --filter admin dev` boots and renders the Admin placeholder page. — verified directly: booted the Vite dev server and curled it; served page has `<title>Employee360 Admin</title>` and `clients/admin/src/App.tsx:1-9` renders "Employee360 Admin".
- [x] AC-3: `pnpm --filter employee dev` boots and renders the Employee placeholder page. — verified directly, same method; `clients/employee/src/App.tsx:1-9` renders "Employee360 Employee Portal".
- [x] AC-4: `pnpm build` completes successfully. — verified directly: `pnpm -r build` (`tsc -b && vite build`) succeeded for both apps, producing `dist/` bundles.
- [x] AC-5: `pnpm lint` completes successfully. — verified directly: `pnpm -r lint` (`eslint .`) passed for both apps.
- [x] AC-6: `pnpm typecheck` completes successfully. — verified directly: `pnpm -r typecheck` (`tsc -b`) passed for both apps.
- [x] AC-7: Tailwind CSS is configured and working in both applications. — `clients/{admin,employee}/vite.config.ts:1-8` wires `@tailwindcss/vite`; `src/index.css:1` is `@import "tailwindcss"`; `App.tsx` uses Tailwind utility classes. Confirmed *working* (not just configured) — the `pnpm build` output emitted a real generated stylesheet (`dist/assets/index-*.css`, ~5 kB) for both apps.
- [x] AC-8: Both applications are registered correctly within the pnpm workspace. — `pnpm-workspace.yaml:1-3` lists `clients/*` and `packages/*`; `pnpm -r` commands above correctly picked up both `admin` and `employee` as workspace projects.
- [x] AC-9: No API wiring or feature code is introduced. — confirmed by diff read: `packages/api-client/package.json` and `packages/ui/package.json` are placeholder manifests only (no source files), and `clients/*/src/` contains only the default Vite scaffold (`App.tsx`, `main.tsx`, `index.css`).

## Note on review base
This branch stacks EMPLOYEE36-2/3/4, each already reviewed at its own tip commit
(`plan/reviews/review-EMPLOYEE36-{3,4}.md`, both 🟢 Merge). `origin/main` (`17bfe53`)
hasn't caught up with any of those merges yet, so diffing straight against it would
re-surface already-reviewed work. This review instead diffs from `15c1402` (the last
commit before EMPLOYEE36-5 work started) to `HEAD`.

Two commits in that range — `9f8894c` (Makefile consolidation) and `15c1402` (Swagger
docs) — predate the three commits actually tagged `(EMPLOYEE36-5)` (`974baa5`,
`1737aea`, `730e146`) and carry no ticket reference. They're pre-existing, already-
functioning chores, not part of this ticket's scope, so they weren't re-reviewed in
depth here (see 🟢 Minor note below).

## Latest commit reviewed
`49c2a16` — ci(frontend): add GitHub Actions workflow and make check-clients (EMPLOYEE36-5)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] `.github/workflows/` — only `backend-ci.yml` existed; there was no frontend CI workflow gated on `clients/**`/`packages/**`. **(resolved in `49c2a16`)** New `.github/workflows/frontend-ci.yml` triggers on `clients/**`, `packages/**`, `package.json`, `pnpm-workspace.yaml`, `pnpm-lock.yaml`; installs with `pnpm install --frozen-lockfile`, runs the new `make check-clients` target (`lint-clients` → `typecheck` → `test-clients` → `build-clients`, mirroring `check-backend`'s composition), and adds a smoke-test step that boots both Vite dev servers and curls them for the expected placeholder content. Verified locally: `make check-clients` passes end-to-end, the smoke-test logic (background both servers, poll with curl, assert page content, `pkill` cleanup) was exercised directly against real running dev servers and correctly detected both apps serving their titles, and the YAML parses cleanly.
- [ ] `backend/.air.toml`, `Makefile:28-34` (commit `730e146`, tagged `EMPLOYEE36-5`) — wires Go hot-reload (`air`) into `make dev-api`. This is backend dev tooling unrelated to EMPLOYEE36-5's stated scope ("Frontend: monorepo & app scaffolding") and to any of its ACs. Not a cycle violation (it's still Cycle 1 tooling) and it's low-risk/additive, but bundling an unrelated backend concern into a frontend-scaffolding ticket's commit history makes the ticket harder to review/revert independently. Still open — this is about a commit already on the branch and wasn't addressed by the new commit. Suggested fix: land backend-tooling changes like this under their own ticket/commit going forward (no action needed retroactively; not worth rewriting history for).

### 🟢 Minor
- [ ] `clients/admin/eslint.config.js` / `clients/employee/eslint.config.js` and the paired `tsconfig.*.json` files are byte-identical (both are the unmodified `pnpm create vite` template output run twice). Not wrong for a scaffolding ticket, but a later cycle could hoist a shared base config into `packages/` to avoid the two copies drifting apart.
- [ ] Commits `9f8894c` and `15c1402` (see "Note on review base") carry no Plane ticket reference in their subject line, unlike every other commit on this branch. Purely a process nit — doesn't affect this ticket's mergeability.

## Verdict
- **Score:** 93/100 (100 − 1×5 − 2×1)
- **Flag:** 🟢 Merge
- **Notes:** All nine acceptance criteria remain met (unaffected by this iteration's changes). The frontend CI gap flagged in the initial review is now closed: `frontend-ci.yml` runs `make check-clients` (lint, typecheck, test, build) plus a live dev-server smoke test on every push/PR touching the frontend workspace, matching `backend-ci.yml`'s pattern and giving AC-1/4/5/6 real, ongoing enforcement instead of one-time manual verification. The remaining 🟡 Major (unrelated backend `air` hot-reload change bundled into commit `730e146`) is a historical commit-hygiene note, not something this iteration's diff could or needed to address — it stays open as a non-blocking process note. The two 🟢 Minor notes (duplicated eslint/tsconfig scaffolding, two untagged pre-EMPLOYEE36-5 chore commits) are unchanged. Clear to merge.

## Re-review Log

### Iteration 2 — 2026-09-15 — sha `49c2a16`
- **Resolved:** 🟡 Major — missing frontend CI workflow (added `.github/workflows/frontend-ci.yml` + `make check-clients`/`build-clients`)
- **Still open:** 🟡 Major — unrelated backend `air` hot-reload change bundled into commit `730e146`; 🟢 Minor — duplicated eslint/tsconfig scaffolding between admin/employee; 🟢 Minor — untagged `9f8894c`/`15c1402` commits predating this ticket
- **New issues:** none
- **Score:** 88 → 93 (+5)
- **Verdict:** 🟢 Merge → 🟢 Merge
