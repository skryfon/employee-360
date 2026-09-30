# Review: EMPLOYEE36-20 — D2 — Admin app: login + forgot/reset password pages

> Branch: ebin/feat/EPIC-D/EMPLOYEE36-20 | Last reviewed: 2026-09-30 22:10 | Iteration: 2 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-20
**State:** unstarted (Plane state group)
**Link:** n/a (not returned by Plane)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
`clients/admin/src/features/auth/`. Depends on D1 (api-client) and EPIC-B's B6 (auth routes).
Scope: login page (RHF + Zod, calls `login`); forgot-password page (email only, always generic success, enumeration-safe); reset-password page (token-in-URL, new password + confirm, calls `resetPassword`); Zustand auth store (access token, current user, roles) + TanStack Query mutations wrapping D1's methods.

### Acceptance Criteria
- [x] AC-1: A successful login redirects into the authenticated admin shell and persists the session across a page reload
  - Impl: `LoginPage.tsx:28-29`, `RequireAuth.tsx:7`, `authStore.ts:21-35` (+ api-client `session.ts` localStorage adapter)
  - Test: redirect `auth.test.tsx:48-57`; reload persistence `auth.test.tsx` `reload persistence` block (rehydrated session+user renders shell)
- [x] AC-2: An invalid login shows an inline error without leaking whether the email or the password was wrong
  - Impl: `LoginPage.tsx:12,34`, `authMutations.ts:26-30`; Test: `auth.test.tsx:59-67`, `69-78`
- [x] AC-3: The forgot-password page's success state is identical regardless of whether the email exists
  - Impl: `ForgotPasswordPage.tsx:26-27`; Test: `auth.test.tsx:82-95` (200/404/500/network)

## Latest commit reviewed
`445a01a` — fix(admin): address EMPLOYEE36-20 review findings

Scope reviewed: this commit only (21 files under `clients/admin` + lockfile). The rest of the `origin/main...HEAD` diff is earlier stacked EPIC-B/C/D work with its own reviews. `typecheck`, `lint`, `test` (13 tests) all pass.

## Findings

### 🔴 Critical
- [x] (resolved in 445a01a) `clients/admin/src/features/auth/pages/auth.test.tsx` (whole file; impl `stores/authStore.ts:21-35`) — AC-1's "persists the session across a page reload" has no test. The only login test (`:48-57`) asserts landing in the shell and in-memory state; nothing re-renders/re-imports with a pre-populated localStorage session + persisted `employee360.admin.auth` user and asserts the user lands in the shell (and that a cleared session lands on `/login`). — Per checklist §1 an AC without a scenario-specific test is unmet. — Fix: add a test that seeds `setSession(...)` + `useAuthStore.setState({ user })` (or localStorage and `useAuthStore.persist.rehydrate()`), renders `<App/>` at `/`, and expects the admin shell; plus a case with token but no user / user but no token redirecting to `/login`.

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [x] (resolved in 445a01a) `LoginPage.tsx:34` — every failure (network down, 5xx, 429 rate-limit) shows "Invalid email or password." Non-leaking but misleading; consider a separate generic "Unable to sign in, try again" for non-401/403 responses.
- [x] (resolved in 445a01a) `ResetPasswordPage.tsx:13,34-45` — reset token stays in the URL/history after success and a still-active session isn't cleared; consider `navigate(..., {replace:true})` to `/login` on success and `clearSession()` (and add `Referrer-Policy: no-referrer` for the admin app).
- [x] (resolved in 445a01a) `authStore.ts:46-48` / `RequireAuth.tsx:7` — guard checks token+user only, not `ADMIN_ROLES`; a tampered persisted `user` renders the shell chrome (server still enforces). Cheap hardening: require an admin role in `isAuthenticated` or the guard.

## Verdict
- **Score:** 100/100
- **Flag:** 🟢 Merge
- **Notes:** All four prior findings resolved in 445a01a. AC-1 now has reload-persistence tests (seeded session + rehydrated user renders shell; token-only, user-only and non-admin redirect to login). Login shows distinct messages for 401/403/non-admin vs network/5xx/429; reset success clears session and replaces history to /login; guard requires an admin role. typecheck, lint and 22 tests pass. No new issues. Branch is clear for merge.

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `445a01a`
- **Resolved:** AC-1 reload-persistence test (critical); login error message split; reset-password session clear + token strip; admin-role guard
- **Still open:** none
- **New issues:** none
- **Score:** 77 → 100 (+23)
- **Verdict:** 🟢 Merge
