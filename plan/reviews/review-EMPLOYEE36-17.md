# Review: EMPLOYEE36-17 — C2 — Invitation delivery/routes

> Branch: ebin/feat/EPIC-C/EMPLOYEE36-17 | Last reviewed: 2026-09-30 19:30 | Iteration: 3 | Verdict: 🟢

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
- [x] AC-2: End-to-end via curl: an admin invites a new user, the invitation email appears in a local SMTP catcher, and accepting it produces an active user — impl `routes.go:63-80` + C1 usecases + worker; test `backend/integration/invitation_flow_test.go` `TestInvitationFlow_InviteDeliverAccept` (invite → worker delivery → token → accept → active user + login). Test executed against a throwaway Postgres DB: PASS (iteration 3); manual curl/Mailpit run not performed (no Docker), covered by the integration test

## Latest commit reviewed
`5a56a82234490be9bba57f15e59f84a705887f86` — feat(backend): add invitation handlers and routes with tests (EMPLOYEE36-17)

## Findings

### 🔴 Critical
- [x] AC-2 unmet — no end-to-end test (resolved in 3e5dee8: `integration/invitation_flow_test.go`)

### 🟡 Major
- [x] `backend/integration/invitation_flow_test.go` — test never executed (resolved in iteration 3: `go test -tags=integration ./integration -run TestInvitationFlow -v` PASS on a throwaway DB; full integration suite also PASS)
- [x] `backend/docs/swagger.{json,yaml}`, `docs.go` — stale Swagger (resolved in 3e5dee8: invitation routes now present)

### 🟢 Minor
- [x] `Accept` inline error mapping (resolved in 3e5dee8: folded into `writeInvitationError`)
- [x] Hard-coded "8 characters" message (resolved in 3e5dee8: uses `err.Error()`; also applied to `auth_handler.go` ResetPassword, same text)

## Verdict
- **Score:** 100/100
- **Flag:** 🟢 Merge
- **Notes:** All findings resolved and both ACs have implementation and executed-test evidence. The invitation flow integration test passes against a real Postgres (invite → worker delivery → accept → active user → login), the full integration suite passes, and `go vet`/`go test ./...` are clean. The manual curl/Mailpit run was not performed (no Docker), but the integration test covers the same path with a capturing mailer.

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `3e5dee8`
- **Resolved:** AC-2 missing e2e test (test added), stale Swagger, Accept error-mapping inconsistency, hard-coded password message
- **Still open:** none of the original findings
- **New issues:** 🟡 integration test not yet executed
- **Score:** 73 → 95 (+22)
- **Verdict:** 🟡 Reviewer call

### Iteration 3 — 2026-09-30 — sha `3e5dee8` (no code change)
- **Resolved:** integration test not executed — ran on a throwaway DB: PASS; full integration suite PASS
- **Still open:** none
- **New issues:** none
- **Score:** 95 → 100 (+5)
- **Verdict:** 🟢 Merge
