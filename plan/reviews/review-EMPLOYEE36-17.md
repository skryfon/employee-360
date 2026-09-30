# Review: EMPLOYEE36-17 — C2 — Invitation delivery/routes

> Branch: ebin/feat/EPIC-C/EMPLOYEE36-17 | Last reviewed: 2026-09-30 19:30 | Iteration: 2 | Verdict: 🟡

## Ticket
**Identifier:** EMPLOYEE36-17
**State:** started
**Link:** n/a (Plane MCP response had no URL)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
Gin handlers and route registration for C1's invitation usecases (EMPLOYEE36-16). Depends on C1 and EPIC-B's B5 (auth/tenant middleware). Scope: `POST /api/v1/users/invitations` (admin-only), `POST /api/v1/users/invitations/:id/resend`, `DELETE /api/v1/users/invitations/:id`, `GET /api/v1/users/invitations`, and unauthenticated `POST /api/v1/invitations/accept`; wire auth/tenant middleware onto every route except accept.

Note: the branch is stacked on unmerged EPIC-B/EPIC-C work (`origin/main...HEAD` is 198 files); only commit `5a56a82` belongs to this ticket, and this review covers that commit (`0bd19c2..HEAD`).

### Acceptance Criteria
- [x] AC-1: All invitation-management routes require a valid admin access token; `invitations/accept` is reachable unauthenticated — impl `routes.go:63-80` (Auth → Tenant → RequireRole(admin, super_admin); accept registered outside the group); test `routes_test.go` 401/403/2xx matrix across all 4 management routes + unauthenticated accept check
- [x] AC-2: End-to-end via curl: an admin invites a new user, the invitation email appears in a local SMTP catcher, and accepting it produces an active user — impl `routes.go:63-80` + C1 usecases + worker; test `backend/integration/invitation_flow_test.go` `TestInvitationFlow_InviteDeliverAccept` (invite → worker delivery → token → accept → active user + login). Test compiles (`go vet -tags=integration`) but has **not been executed**; curl/Mailpit run not performed (see Major)

## Latest commit reviewed
`5a56a82234490be9bba57f15e59f84a705887f86` — feat(backend): add invitation handlers and routes with tests (EMPLOYEE36-17)

## Findings

### 🔴 Critical
- [x] AC-2 unmet — no end-to-end test (resolved in 3e5dee8: `integration/invitation_flow_test.go`)

### 🟡 Major
- [ ] `backend/integration/invitation_flow_test.go` — the new integration test has never been executed (and the manual curl/Mailpit run was not done; no Docker/disposable DB available), so AC-2 is backed by an unverified test — run `go test -tags=integration ./integration -run TestInvitationFlow -v` against a disposable DB (`DATABASE_*`) and record the result on the ticket
- [x] `backend/docs/swagger.{json,yaml}`, `docs.go` — stale Swagger (resolved in 3e5dee8: invitation routes now present)

### 🟢 Minor
- [x] `Accept` inline error mapping (resolved in 3e5dee8: folded into `writeInvitationError`)
- [x] Hard-coded "8 characters" message (resolved in 3e5dee8: uses `err.Error()`; also applied to `auth_handler.go` ResetPassword, same text)

## Verdict
- **Score:** 95/100
- **Flag:** 🟡 Reviewer call
- **Notes:** All prior findings are resolved and all ACs now have implementation and test evidence. `go vet` (with and without the `integration` tag) and `go test ./...` pass. One major remains open: the new integration test that backs AC-2 has not been run against a real database, so the end-to-end behaviour is still unverified. Run it once (and ideally the curl/Mailpit flow) to reach 🟢 Merge.

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `3e5dee8`
- **Resolved:** AC-2 missing e2e test (test added), stale Swagger, Accept error-mapping inconsistency, hard-coded password message
- **Still open:** none of the original findings
- **New issues:** 🟡 integration test not yet executed
- **Score:** 73 → 95 (+22)
- **Verdict:** 🟡 Reviewer call
