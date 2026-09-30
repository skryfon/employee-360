# Review: EMPLOYEE36-19 — D1 — api-client: auth & invitation methods + Axios interceptors

> Branch: ebin/feat/EPIC-D/EMPLOYEE36-19 | Last reviewed: 2026-09-30 22:05 | Iteration: 3 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-19
**State:** unstarted (Plane state group)
**Link:** n/a (not returned by Plane)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
Shared `packages/api-client` layer that the rest of EPIC-D calls through. Adds API methods (login, refresh, logout, forgotPassword, resetPassword, inviteUser, listInvitations, resendInvitation, revokeInvitation, acceptInvitation), a request interceptor (access token + `X-Tenant-ID`), and a response interceptor (one silent refresh on 401, retry, else force logout).

### Acceptance Criteria
- [x] AC-1: All methods are typed against the request/response shapes the EPIC-B/EPIC-C handlers actually return — impl: `auth.ts:1-50` (Orval-generated models from the backend swagger); test: `auth.test.ts:33-153` (per-method shape tests) plus a clean `tsc -b`
- [x] AC-2: A request made with an expired access token transparently refreshes once and succeeds — impl: `client.ts:204-240`; test: `client.test.ts:98` and `auth.test.ts:155` (expired token on an invitation call)
- [x] AC-3: No component or page constructs its own Axios call — impl: `client.ts` exports the shared `apiClient` and `apiRequest` (the Orval mutator); grep of `clients/*` finds no axios/fetch use. Test evidence: none for this AC, but it is a structural rule and no violation exists in the tree (see Minor 1).

## Latest commit reviewed
`2033041` — refactor(api-client): drop non-null assertion in logout (EMPLOYEE36-19)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [x] (resolved in 1809583) `session.ts:1-20`, `session.ts:118` — refresh token is persisted to localStorage by default (readable by any script on the origin). The tradeoff is documented and there is an opt-out. Suggest confirming it with the team and pairing it with a strict CSP in the apps (D2+).
- [x] (resolved in 1809583; ESLint `no-restricted-imports` axios + `no-restricted-globals` fetch in both clients) AC-3 has no enforcement — consider an ESLint `no-restricted-imports` rule for `axios` in `clients/*` so the rule is checked automatically rather than by convention.
- [x] (resolved in 1809583; `logout()` does one best-effort refresh on 401 then retries with the rotated token, 2 new tests) `client.ts:59-66` — `/auth/logout` is in `NO_REFRESH_PATHS`, so a logout with an expired access token gets a 401 with no refresh and the server-side revoke may not happen (the local session is still cleared). Confirm that the backend logout route does not require a valid access token, or refresh first.
- [ ] Commit `612aada` (design-system skill + agent doc edits) is unrelated to this ticket. Harmless, but prefer a separate PR.
- [x] (resolved in 2033041) `auth.ts:84` — `getSession()!.refreshToken` uses a non-null assertion after a separate `getSession()?.refreshToken` guard. Capture the token in a `const` and use it instead.

## Verdict
- **Score:** 99/100
- **Flag:** 🟢 Merge
- **Notes:** Iteration 3: four of five minors fixed (only the unrelated commit remains), typecheck and 72 tests pass. The change is frontend-only (api-client), which fits Cycle 2's "frontend that consumes the backend" scope. It has no tenant or security invariant violations: the tenant comes from the login response user and is never re-derived, and the backend ignores `X-Tenant-ID`. Bearer auth keeps it headless. The refresh logic is solid: a single in-flight promise, a cross-tab Web Lock, refresh-token-reuse handling, and only auth rejections force logout while transient errors keep the session. Only minor nits remain.

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `1809583`
- **Resolved:** localStorage/CSP note, axios lint enforcement, logout with expired token
- **Still open:** unrelated design-system commit `612aada`
- **New issues:** non-null assertion in `logout()` (🟢)
- **Score:** 97 → 98 (+1)
- **Verdict:** 🟢 Merge

### Iteration 3 — 2026-09-30 — sha `2033041`
- **Resolved:** non-null assertion in `logout()`
- **Still open:** unrelated design-system commit `612aada`
- **New issues:** none
- **Score:** 98 → 99 (+1)
- **Verdict:** 🟢 Merge
