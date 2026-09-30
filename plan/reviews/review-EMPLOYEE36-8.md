# Review: EMPLOYEE36-8 — EPIC-C: Cycle 2 — Onboarding Invitations (Overview)

> Branch: feature/EPIC-C | Last reviewed: 2026-09-30 19:20 | Iteration: 2 | Verdict: 🟡

## Ticket
**Identifier:** EMPLOYEE36-8 (epic; children EMPLOYEE36-16 = C1, EMPLOYEE36-17 = C2)
**State:** backlog
**Link:** n/a (Plane MCP response had no URL)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

Note: the only branch matching `EMPLOYEE36-8` (`ebin/feat/EPIC-B/EMPLOYEE36-8`) carries EPIC-B migration work, so at the user's direction the epic was reviewed on `feature/EPIC-C`. `origin/main...HEAD` is ~200 files (includes unmerged EPIC-B work); the epic's own scope is `0820328..HEAD` (35 backend files, ~3.3k lines: C1 `0bfaef4`..`30278ee`, C2 `5a56a82`..`3e5dee8`). Both children already have per-ticket reviews at 🟢 (`review-EMPLOYEE36-16.md`, `review-EMPLOYEE36-17.md`); this review re-verifies them at epic level.

### Description
Admin-driven onboarding: Tenant Admins invite new users (admin or employee) by email; invitees are pending until they accept. Builds on EPIC-B (users/roles schema, auth/tenant middleware, B8 EventPublisher/outbox). Scope guardrail: no frontend; tenant-scoped only.

### Acceptance Criteria
Epic Definition of Done:
- [x] DoD-1: Admin invites by email + role (+ optional dept/position); pending `users` row (`is_active=false`, no password) + invitation row created — impl `usecase/implementation/invitation/invite_user.go:57-165`; test `invitation_test.go` invite tests, `integration/invitation_flow_test.go`
- [x] DoD-2: Invitation email delivered asynchronously by `cmd/worker` (not in request cycle) — impl `invite_user.go:158` (`eventPublisher.Publish` in tx, no email import); test `integration/invitation_flow_test.go` `TestInvitationFlow_InviteDeliverAccept` (capturing mailer, not a real SMTP catcher), archtest `TestEmailServiceSendOnlyCalledFromWorker`
- [x] DoD-3: `POST /api/v1/invitations/accept` (unauthenticated) consumes token, password set for every invitee — impl `routes.go:63`, `accept_invitation.go:36-85`; test `TestAccept_RequiresPasswordForEveryRole`, `routes_test.go` unauthenticated accept check
- [x] DoD-4: Admin can resend and revoke a pending invitation — impl `resend_invitation.go`, `revoke_invitation.go`, `routes.go:74-75`; test `TestResendRevoke_RejectNonPending`, `routes_test.go`, handler tests
- [x] DoD-5: Admin can list their tenant's invitations — impl `list_invitations.go`, `user_invitation_repository.go:List` (tenant-scoped), `routes.go:73`; test `user_invitation_repository_test.go` isolation, handler/routes tests

C1 (EMPLOYEE36-16):
- [x] C1-AC-1: No cross-tenant invite even with body `tenant_id` — impl `invite_user.go:66-80` (tenant from ctx; DTO has no tenant field); test `TestInvite_TenantFromContextOnly`
- [x] C1-AC-2: Tokens stored only as hashes — impl `invite_user.go:96/118`, `resend_invitation.go`; test `TestTokensStoredOnlyAsHashes`
- [x] C1-AC-3: Resend/Revoke reject accepted or revoked — impl `resend_invitation.go`, `revoke_invitation.go`, conditional `pendingWhere` update; test `TestResendRevoke_RejectNonPending`, repo conditional-update tests
- [x] C1-AC-4: Accept requires a password for every role — impl `accept_invitation.go:44`; test `TestAccept_RequiresPasswordForEveryRole`
- [x] C1-AC-5: Invite/Resend use only `EventPublisher`; rollback leaves nothing — impl `invite_user.go:138-160`; test `TestInviteAndResend_RollbackLeavesNothing`, `integration/invitation_rollback_test.go`, archtest

C2 (EMPLOYEE36-17):
- [x] C2-AC-1: Management routes need valid admin token; accept unauthenticated — impl `routes.go:59-77` (Auth → Tenant → RequireRole(admin, super_admin)); test `routes_test.go` 401/403/2xx matrix
- [x] C2-AC-2: End-to-end invite → email → accept → active user — impl routes + usecases + worker; test `integration/invitation_flow_test.go` (see note in Notes: not re-executed here)

## Latest commit reviewed
`f454e82` — docs(review): EMPLOYEE36-8 epic review (rate limiter in `e91bab7`)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [ ] `backend/config/config.go` (`Validate`) / `routes.go:35` — new `rate_limit` settings are not validated. With `enabled=true`, `RATE_LIMIT_BURST=0` (or negative) makes `rate.NewLimiter` reject every request, and `requests_per_second<=0` gives an infinite `Retry-After`; a typo in env silently 429s the whole API except health. Fix: in `Validate()` require `requests_per_second > 0` and `burst >= 1` when enabled (and idle_ttl/cleanup_interval >= 0), with a config test.
- [x] (withdrawn per maintainer: explicit `tenantID` param is accepted when the value comes from ctx via handler/usecase) `persistence/user_invitation_repository.go:41-130` (also `org_reference_repository.go`) — repository methods take `tenantID` as a parameter (`GetByID`, `MarkAccepted`, `List`, `DepartmentExists`, …) instead of reading it from `context.Context`. Under the project's literal Invariant 1 / review checklist §3 this is 🔴; I downgraded to 🟡 because the usecases always derive the value from context (never the request), the pattern matches the pre-existing EPIC-B repositories (`UserRepository`), and tenant isolation tests exist. Fix: either accept as project convention and record it in `shared-context.md`/`backend.md`, or move to ctx-derived scoping across repos. Reviewer call.

### 🟢 Minor
- [ ] `backend/config/config_test.go` — no tests for the new config: `RATE_LIMIT_*` env/default loading and `applyTrustedProxiesEnvOverride` (empty, CSV with spaces). Add cases.
- [ ] `packages/api-client/src/generated/hooks/` — Swagger now documents the invitation routes but the generated API client has no invitation hooks (only `auth`). Regenerate when the invitation UI cycle starts (not in this epic's scope per guardrail).
- [x] (resolved: global per-IP rate-limit middleware, `middleware/ratelimit.go`, wired in `routes.go`; uncommitted) `invitation_handler.go` `Accept` / `routes.go:63` — unauthenticated endpoint that verifies tokens and hashes passwords has no rate limiting/throttle. Token entropy makes guessing impractical, but a per-IP limit would protect against hash-cost DoS. Track for a hardening ticket.

## Verdict
- **Score:** 93/100
- **Flag:** 🟡 Reviewer call
- **Notes:** All 12 ACs (5 epic DoD + 5 C1 + 2 C2) have implementation and test evidence, and both child tickets carry prior 🟢 reviews. Independently verified this session: `go build ./...`, `go vet ./...` and `go test ./internal/...` all pass (usecase, archtest, handlers, routes, container). Route wiring, tenant-from-context in usecases, hash-only token storage, conditional (pending-only) updates, transactional accept/revoke and audit logging were confirmed by reading. The DB-backed integration tests (`integration/invitation_*_test.go`) and live-DB repository tests were NOT re-run — no database is configured in this environment; they rely on the child reviews' recorded passes. The one major raised (repo methods taking a `tenantID` parameter) was withdrawn: the maintainer confirmed that is acceptable when the value is taken from context in the handler/usecase, which holds here. No frontend or later-cycle scope creep found in the epic's range.

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `f454e82`
- **Resolved:** 🟢 no rate limiting on unauthenticated accept (`e91bab7`: per-IP token bucket, 429 + Retry-After, health exempt, trusted proxies default none; middleware/routes tests pass, build/vet clean)
- **Still open:** 🟢 api-client lacks invitation hooks
- **New issues:** 🟡 rate-limit config not validated (burst/rps <= 0 → self-inflicted 429s); 🟢 no config tests for new settings
- **Score:** 98 → 93 (-5)
- **Verdict:** 🟡 Reviewer call
