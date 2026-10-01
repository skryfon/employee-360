# Review: EMPLOYEE36-23 — D5 — Invitation-accept page

> Branch: ebin/feat/EPIC-D/EMPLOYEE36-23 | Last reviewed: 2026-10-01 | Iteration: 3 | Verdict: 🟢 Merge

## Ticket
**Identifier:** EMPLOYEE36-23
**State:** started
**Link:** n/a (Plane response had no URL)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
Unauthenticated, token-in-URL page that consumes `POST /api/v1/invitations/accept`. Every invitee (admin or employee) sets a password before completing acceptance. Expired/already-accepted/revoked tokens get a clear, non-generic error. Open question: which app hosts the route — "don't guess silently".

### Acceptance Criteria
- [x] AC-1: An invited admin can set a password and lands able to log in via D2's login page — impl `clients/admin/.../AcceptInvitationPage.tsx` (submit → "Sign in to Admin" link, `state.invitationAccepted`); test `clients/admin/.../auth.test.tsx` "validates token on load … completes acceptance" (ends on login page with banner). Weak: no test that the login actually succeeds afterwards.
- [x] AC-2: An invited employee can set a password and lands able to log in via D4's login page — same, `clients/employee/.../AcceptInvitationPage.tsx` + employee `auth.test.tsx`.
- [x] AC-3 (resolved in 0248419): An expired, already-accepted, or revoked token shows a distinct, actionable error state — **was unmet (iteration 1)**; now met — `lookupUsableInvitation` (`common.go`) returns `ErrInvitationExpired/Revoked/Accepted` → 410/403/409 with stable codes; page maps by `error.code` to distinct states; tests `auth.test.tsx` 'INVITATION_EXPIRED/REVOKED/ACCEPTED' assert distinct text. Original gap: backend `validate_invitation.go:44-58` maps expired / revoked / accepted / unknown / user-already-active all to `ErrInvalidToken` (400 "invalid or expired invitation token"); `AcceptInvitationPage.tsx` `formatAcceptInvitationError` collapses everything to one message ("invalid, has expired, or has already been accepted"). The expired and revoked tests (`auth.test.tsx`, "expired"/"revoked") assert the *same* text, so nothing is distinct.

## Latest commit reviewed
`0248419` — fix(auth): address EMPLOYEE36-23 review, add sonner toasts for invitation actions

Scope note: branch is stacked on earlier unmerged tickets (EMPLOYEE36-19…22 etc., already reviewed); only the EMPLOYEE36-23 commit was reviewed here.

## Findings

### 🔴 Critical
- [x] `backend/internal/usecase/implementation/invitation/validate_invitation.go:44-58`, `clients/*/src/features/auth/pages/AcceptInvitationPage.tsx:13-27` — AC-3 unmet: all failure states are indistinguishable (single generic message), contradicting the ticket's "distinct, non-generic" requirement (and its note that enumeration safety doesn't apply here). Fix: have validate return a reason (e.g. `ErrInvitationExpired` / `ErrInvitationRevoked` / `ErrInvitationAccepted`, or `valid:false` + `reason`) with stable error codes; render a separate message/action per state (expired → ask admin to resend; revoked → contact admin; accepted → "Sign in"). Add tests asserting each state's distinct text. (resolved in 0248419)

### 🟡 Major
- [x] `AcceptInvitationPage.tsx:98` (admin) / employee equivalent — the open question (which app hosts the route) was resolved silently by duplicating the page in both clients, with no decision recorded in `plan/cycles/cycle-02…` or `plan/architecture/frontend.md`. The invite link uses a single `FrontendURL` (`invite_user.go:156`) and the validate response carries no role, so an employee landing on the admin app sees "Sign in to Admin", whose login rejects non-admins (`NotAdminError`). Fix: return role (or target app) from validate and route/redirect accordingly, or use one neutral entry point; document the decision. (resolved in 0248419)
- [x] `backend/internal/delivery/http/routes_test.go` — new unauthenticated route `GET /invitations/validate` has no route-level test (only a fake wired in); accept has `TestInvitationRoutes_AcceptRateLimited`. Add a route test that it works without auth and is covered by the rate limiter (token-validity oracle). (resolved in 0248419)
- [x] `AcceptInvitationPage.tsx:13-27` — error classification by substring-matching server message text (`'token'`, `'conflict'`, …) is brittle and would misclassify unrelated errors. Use HTTP status / error `code` from the envelope. (resolved in 0248419)
- [x] `plan/cycles/cycle-02-auth-onboarding.md` — new backend endpoint `GET /api/v1/invitations/validate` (outside the frontend-only D5 scope) is not recorded in the cycle scope doc; update the doc or split the backend change. (resolved in 0248419)

### 🟢 Minor
- [x] `packages/api-client/src/auth.ts` `validateInvitation` hand-rolls `apiRequest` with a dynamic import although a generated `getApiV1InvitationsValidate` hook was added in the same commit; inconsistent with `acceptInvitation`. (resolved in 0248419)
- [x] Backend `ValidateInvitationResponse.valid` is still always `true` (email now shown, client type dropped it); remove it. (resolved in 60053cd)
- [x] `validate_invitation.go` duplicates the lookup/usability logic of `accept_invitation.go`; extract a shared helper. (resolved in 0248419)
- [x] Token travels in a query string (logged by proxies); acceptable for emailed link but consider `Referrer-Policy: no-referrer` on the page. (resolved in 60053cd)
- [x] Stray trailing blank lines (resolved in 0248419)
- [ ] Admin/employee `AcceptInvitationPage` + tests remain near-identical copies; consider extracting shared pieces.
- [x] On `INVITATION_ACCEPTED` the page's "Sign in" targets the current host's login (role unknown then); normally correct since links now target the invitee's app. (resolved in 60053cd)
- [ ] Commit `0248419` bundles unrelated sonner toast work with the review fixes; prefer separate commits.

## Verdict
- **Score:** 98/100
- **Flag:** 🟢 Merge
- **Notes:** All prior critical/major findings are resolved: distinct, code-based error states (validate and accept), role returned by validate with cross-app redirect and the host decision recorded in the cycle doc, route-level unauthenticated + rate-limit tests for validate, and the endpoint documented. `go build/test ./...`, `pnpm -r typecheck` and `pnpm -r test` all pass. Remaining items are minor nits (2 open: duplicated admin/employee page, bundled commit).

## Re-review Log
### Iteration 2 — 2026-10-01 — sha `0248419`
- **Resolved:** AC-3 (distinct states), host/role placement + decision doc, validate route + rate-limit tests, code-based error classification, cycle doc update, generated hook usage, email shown, shared lookup helper, trailing blank lines.
- **Still open:** Referrer-Policy suggestion, `valid` field, duplicated admin/employee page.
- **New issues:** minor — ACCEPTED sign-in host, bundled commit.
- **Score:** n/a (hard gate) → 95 (Merge)
- **Verdict:** 🟢 Merge

### Iteration 3 — 2026-10-01 — sha `60053cd`
- **Resolved:** Referrer-Policy meta tag (both apps), removed `valid` from validate response/swagger/cycle doc/generated client, ACCEPTED-state sign-in copy points to the other portal (tests assert it).
- **Still open:** duplicated admin/employee accept page (needs `packages/ui` scaffolding), bundled sonner commit (history not rewritten).
- **New issues:** none
- **Score:** 95 → 98 (+3)
- **Verdict:** 🟢 Merge
