# Review: EMPLOYEE36-22 — Employee login page + auth store (title not fetched; Plane unavailable)

> Branch: feature/EPIC-D | Last reviewed: 2026-10-01 07:59 | Iteration: 1 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-22
**State:** unknown (Plane MCP returned ECONNRESET twice; details pasted by user)
**Link:** n/a
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
`clients/employee/src/features/auth/`. Depends on D1 and EPIC-B's B6 (login route).
Scope: single-step email + password login page → `login`; Zustand store for employee auth state, shared shape with D2's admin store where it makes sense.

### Acceptance Criteria
- [x] AC-1: A successful login redirects into the authenticated employee shell and persists the session across a page reload
  - Impl: `LoginPage.tsx:38` (navigate on success), `authStore.ts` (persist `user`, token via api-client), `RequireAuth.tsx`
  - Test: `auth.test.tsx:49` (login → shell), `auth.test.tsx:140` (shell rendered from persisted session + user)
- [x] AC-2: Invalid credentials show an inline error and allow retry without losing the entered email
  - Impl: `LoginPage.tsx` (`InlineAlert`, uncontrolled RHF inputs not remounted on error)
  - Test: `auth.test.tsx:61` (401, email retained, retry succeeds), plus 403 / 500 / 429 / network variants

## Latest commit reviewed
`19f59fecfecaa5048c6e61731b2cc05425cac952` — feat(employee): add auth feature, employee shell, and auth store

Scope of review: this commit only (no ticket-named branch exists; work is on the epic branch). Verified locally: `vitest run` 23/23 pass, `tsc` clean.

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [x] `features/auth/queries/authMutations.ts:27` — on a non-employee role, `logout().catch(() => undefined)` swallows failure, so if the logout call fails the api-client session (tokens) stays persisted although the store has no user — fix: call `clearSession()` in the catch/finally, and assert `getSession()` is null in the `auth.test.tsx:95` test. (fixed, uncommitted)
- [x] `stores/authStore.ts:44` — `EMPLOYEE_ROLES` includes `admin`/`super_admin`, so admins can sign into the employee portal; ticket/cycle only say "employee". Confirm intentional and note it in the cycle doc. (fixed, uncommitted: comment in authStore.ts + note in cycle-02 doc)
- [x] `stores/authStore.ts` — near-verbatim copy of the admin store (only key + role list differ). Matches "shared shape" but consider a shared factory in a package as a follow-up. (acknowledged / deferred follow-up)
- [x] `auth.test.tsx:140` — reload test seeds localStorage by hand rather than logging in then remounting, so it doesn't prove the real login path writes the persisted key; add one login → rehydrate → shell test. (fixed, uncommitted)
- [x] Commit also adds forgot/reset password pages + shell/dashboard beyond this ticket's stated scope (login + store). They are within Cycle 2's employee-auth frontend list, so not scope creep; just confirm they belong to a sibling ticket and aren't left unreviewed. (acknowledged / deferred follow-up; within Cycle 2)

## Verdict
- **Score:** 95/100
- **Flag:** 🟢 Merge
- **Notes:** Both ACs have implementation and specific test evidence. Auth state is Zustand-only (token mirrored from api-client, profile persisted), forms use RHF + Zod, server calls go through `packages/api-client`, and no raw fetch/axios was found. Tenant resolution is left to the api-client. Only minor hardening and clarity items remain; the token-cleanup edge case at authMutations.ts:27 is worth fixing before or soon after merge. Plane ticket details were user-pasted, not fetched.

## Re-review Log
_(empty — first review)_
