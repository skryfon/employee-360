# Review: EMPLOYEE36-23 — D5 — Invitation-accept page

> Branch: ebin/feat/EPIC-D/EMPLOYEE36-23 | Last reviewed: 2026-10-01 | Iteration: 1 | Verdict: 🔴 Block

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
- [ ] AC-3: An expired, already-accepted, or revoked token shows a distinct, actionable error state — **unmet**: backend `validate_invitation.go:44-58` maps expired / revoked / accepted / unknown / user-already-active all to `ErrInvalidToken` (400 "invalid or expired invitation token"); `AcceptInvitationPage.tsx` `formatAcceptInvitationError` collapses everything to one message ("invalid, has expired, or has already been accepted"). The expired and revoked tests (`auth.test.tsx`, "expired"/"revoked") assert the *same* text, so nothing is distinct.

## Latest commit reviewed
`0a510cd6e0cd4c420671a8bc1530abad45fc855b` — feat(auth): implement invitation acceptance and validation flow (EMPLOYEE36-23)

Scope note: branch is stacked on earlier unmerged tickets (EMPLOYEE36-19…22 etc., already reviewed); only the EMPLOYEE36-23 commit was reviewed here.

## Findings

### 🔴 Critical
- [ ] `backend/internal/usecase/implementation/invitation/validate_invitation.go:44-58`, `clients/*/src/features/auth/pages/AcceptInvitationPage.tsx:13-27` — AC-3 unmet: all failure states are indistinguishable (single generic message), contradicting the ticket's "distinct, non-generic" requirement (and its note that enumeration safety doesn't apply here). Fix: have validate return a reason (e.g. `ErrInvitationExpired` / `ErrInvitationRevoked` / `ErrInvitationAccepted`, or `valid:false` + `reason`) with stable error codes; render a separate message/action per state (expired → ask admin to resend; revoked → contact admin; accepted → "Sign in"). Add tests asserting each state's distinct text.

### 🟡 Major
- [ ] `AcceptInvitationPage.tsx:98` (admin) / employee equivalent — the open question (which app hosts the route) was resolved silently by duplicating the page in both clients, with no decision recorded in `plan/cycles/cycle-02…` or `plan/architecture/frontend.md`. The invite link uses a single `FrontendURL` (`invite_user.go:156`) and the validate response carries no role, so an employee landing on the admin app sees "Sign in to Admin", whose login rejects non-admins (`NotAdminError`). Fix: return role (or target app) from validate and route/redirect accordingly, or use one neutral entry point; document the decision.
- [ ] `backend/internal/delivery/http/routes_test.go` — new unauthenticated route `GET /invitations/validate` has no route-level test (only a fake wired in); accept has `TestInvitationRoutes_AcceptRateLimited`. Add a route test that it works without auth and is covered by the rate limiter (token-validity oracle).
- [ ] `AcceptInvitationPage.tsx:13-27` — error classification by substring-matching server message text (`'token'`, `'conflict'`, …) is brittle and would misclassify unrelated errors. Use HTTP status / error `code` from the envelope.
- [ ] `plan/cycles/cycle-02-auth-onboarding.md` — new backend endpoint `GET /api/v1/invitations/validate` (outside the frontend-only D5 scope) is not recorded in the cycle scope doc; update the doc or split the backend change.

### 🟢 Minor
- [ ] `packages/api-client/src/auth.ts` `validateInvitation` hand-rolls `apiRequest` with a dynamic import although a generated `getApiV1InvitationsValidate` hook was added in the same commit; inconsistent with `acceptInvitation`.
- [ ] `ValidateInvitationResponse.valid` is always true on success (failures are errors) and `email` is never shown on the page — drop `valid` or display the email ("Setting password for new@acme.com").
- [ ] `validate_invitation.go` duplicates the lookup/usability logic of `accept_invitation.go`; extract a shared helper.
- [ ] Token travels in a query string (logged by proxies); acceptable for emailed link but consider `Referrer-Policy: no-referrer` on the page.
- [ ] Stray trailing blank lines added in `authMutations.ts`, `authSchemas.ts`, `routes.tsx`, `auth.ts`, `auth.test.tsx`; admin/employee page+test files are near-identical copies.

## Verdict
- **Score:** n/a — hard gate (AC-3 unmet)
- **Flag:** 🔴 Block
- **Notes:** Validate-then-set-password flow, tenant derivation from the invitation row, Zod/RHF form, session clearing and backend handler/usecase tests are solid. The ticket's core differentiator — distinct error states for expired/accepted/revoked — is not delivered, and the app-placement open question was decided silently with a role-blind "Sign in to Admin" path for employees. Fix AC-3 and the placement/role handoff, add the route test, then re-review.

## Re-review Log
_(empty — first review)_
