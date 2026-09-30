# Review: EMPLOYEE36-16 — C1 — Invitation domain & usecases

> Branch: ebin/feat/EPIC-C/EMPLOYEE36-16 | Last reviewed: 2026-09-30 18:05 | Iteration: 4 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-16
**State:** started
**Link:** n/a (Plane MCP response had no URL)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
Domain entity/repository plus application logic for admin-driven onboarding, in `backend/internal/usecase/{interface,implementation}/invitation/`. Depends on B1 (migrations), B2 (hash/email/event services), B3 (email infra), B8 (async event publishing/outbox). Scope: `UserInvitation` entity + repository interface; `InviteUserUseCase`, `AcceptInvitationUseCase`, `ResendInvitationUseCase`, `RevokeInvitationUseCase`, `ListInvitationsUseCase`.

Note: the branch is stacked on earlier EPIC-A/B work; only commit `0bfaef4` belongs to this ticket, and this review covers that commit.

### Acceptance Criteria
- [x] AC-1: Admin cannot invite into another tenant even if a different `tenant_id` is supplied in the body — impl `invite_user.go:66-80` (tenant from ctx only; `InviteUserRequest` has no tenant field); test `invitation_test.go` `TestInvite_TenantFromContextOnly` (resolved in 2bb0fd0)
- [x] AC-2: Tokens stored only as hashes — impl `invite_user.go:96`, `resend_invitation.go:78`; test `TestTokensStoredOnlyAsHashes` (resolved in 2bb0fd0)
- [x] AC-3: Resend/Revoke reject accepted or revoked invitations — impl `resend_invitation.go:62`, `revoke_invitation.go:38`; test `TestResendRevoke_RejectNonPending` (resolved in 2bb0fd0)
- [x] AC-4: Accept requires a password for every role — impl `accept_invitation.go:44`; test `TestAccept_RequiresPasswordForEveryRole` (resolved in 2bb0fd0)
- [x] AC-5: Invite/Resend use only `EventPublisher`; rollback leaves nothing — impl `invite_user.go:138`, `resend_invitation.go:92`; tests `TestInviteAndResend_RollbackLeavesNothing` (fake transactor), `archtest` `TestUsecaseLayerImportsNoInfrastructure` / `TestEmailServiceSendOnlyCalledFromWorker` (resolved in 2bb0fd0)

## Latest commit reviewed
`30278ee` — fix(backend): address EMPLOYEE36-16 iteration 3 review findings

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] `invite_user.go:77-114` — `DepartmentID`/`PositionID` from the request are written to the `users` and `user_invitations` rows without checking they belong to the caller's tenant. FKs only check existence, so an admin can attach another tenant's department/position ID (cross-tenant reference). Fix: validate both against the tenant-scoped repos (or a tenant-scoped exists check) before creating, and add a test. (resolved in 2bb0fd0)
- [x] `revoke_invitation.go:41` — Revoke only marks the invitation; the pending inactive `users` row and its role assignment stay. `invite_user.go:88` then rejects any re-invite of that email with `ErrEmailAlreadyExists`, so a revoked (or long-expired) invitee can never be re-invited. Fix: on revoke, delete the still-inactive pending user + user_role in the same transaction, or let Invite reuse a pending user. Add a test for revoke → re-invite. (resolved in 2bb0fd0)
- [x] `accept_invitation.go:49-79` — Token lookup and usability check happen outside the transaction, and `MarkAccepted` is not specified as conditional (`accepted_at IS NULL AND revoked_at IS NULL`). Two concurrent accepts, or an accept racing a revoke, can both pass. Fix: do the lookup inside the tx, or make `MarkAccepted` a conditional update that returns `ErrInvalidToken` on 0 rows affected. Document that in the repo interface. (resolved in 2bb0fd0)
- [x] `invite_user.go` / `resend_invitation.go` / `revoke_invitation.go` — Admin mutations write no audit log entry (`audit` usecase / `audit_logs` table exist). Add audit calls (in the same tx for invite/resend). (resolved in 2bb0fd0)

- [x] `persistence/role_repository.go`, `user_role_repository.go`, `audit_repository.go` (new in 4b7a638) — these adapters scope by the tenant in context (their ports carry no tenantID) but have no dedicated tests. They are only hit indirectly by the invitation rollback test, so cross-tenant reads/deletes (`DeleteByUserID`, `GetRolesByUserID`, `RoleRepo.GetByID`) and the "no tenant in context matches nothing" behaviour are unproven. Add tenant-isolation tests for each. (resolved in 30278ee)

### 🟢 Minor
- [x] `persistence/user_invitation_repository.go:24` — stale comment: says `pendingScope`, the constant is `pendingWhere`. (resolved in 30278ee)
- [x] `invite_user.go:73` — Malformed email returns `ErrInvalidCredentials`, which is the wrong semantic (likely maps to 401). Use a validation error. (resolved in 2bb0fd0)
- [x] Commit also touches `ONBOARDING.md` and `backend/docker-compose.yml` (`./:/app` → `./backend:/app`), unrelated to this ticket. Split out or mention in the PR. (ONBOARDING.md removed in 9fcbb70; docker-compose change intentionally kept per author, 0414369)
- [x] No `UserInvitationRepository` GORM implementation, container wiring or repo tenant-isolation test yet, and AC-5's rollback is proven only with a fake transactor. Confirm a follow-up ticket (C2+) owns this, and that it also implements the new `OrgReferenceRepository` and the conditional-update contract on `MarkAccepted`/`MarkRevoked`/`UpdateToken`. (still open) Also note the raw token necessarily sits in the `river_job` args until job cleanup; consider a short retention for completed email jobs. (resolved in 4b7a638: UserInvitationRepository + OrgReferenceRepository implemented with conditional updates, wired in auth_container.go, live-DB isolation/conditional/rollback tests + integration test `TestInviteUser_RollbackLeavesNothing`)

## Verdict
- **Score:** 100/100
- **Flag:** 🟢 Merge
- **Notes:** Iteration 2: all four majors and the email nit are fixed with tests; one minor (repo implementations deferred) remains. Original notes: All five ACs have both implementation and test evidence, and tenant is derived from context only with no tenant field in the DTOs. Layering is clean and the archtest guards the usecase → infrastructure boundary. Open items are the un-validated department/position tenant ownership, the revoke/re-invite dead end, a non-atomic accept, and missing audit logging. None is a hard block, but the department/position cross-tenant reference should be fixed before merge.

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `0414369`
- **Resolved:** dept/position tenant check (OrgReferenceRepository + tests), revoke cleanup + re-invite (tx, guarded delete), atomic accept (in-tx lookup, conditional MarkAccepted, race test), audit logging on invite/resend/revoke (tx-scoped, rollback test), ErrInvalidEmail, unrelated ONBOARDING.md/compose (compose kept intentionally)
- **Still open:** 🟢 GORM repo implementations (UserInvitationRepository, OrgReferenceRepository), wiring, tenant-isolation tests deferred to a later ticket
- **New issues:** none. build, vet and usecase/archtest/domain tests pass
- **Score:** 77 → 99 (+22)
- **Verdict:** 🟢 Merge

### Iteration 3 — 2026-09-30 — sha `4b7a638`
- **Resolved:** missing GORM repos/wiring/tenant-isolation and rollback tests for UserInvitationRepository and OrgReferenceRepository
- **Still open:** none from before
- **New issues:** 🟡 role/user_role/audit adapters lack tenant-isolation tests; 🟢 stale `pendingScope` comment. Conditional updates, tenant scoping and container wiring verified by reading; build, vet and `go test ./internal/...` pass
- **Score:** 99 → 94 (-5)
- **Verdict:** 🟡 Reviewer call

### Iteration 4 — 2026-09-30 — sha `30278ee`
- **Resolved:** tenant-isolation tests for role/user_role/audit adapters (own tenant, foreign tenant, no-tenant context); stale `pendingWhere` comment
- **Still open:** none
- **New issues:** none. build, vet and `go test ./internal/...` pass; live-DB tests reported passing by the implementing agent (not re-run here)
- **Score:** 94 → 100 (+6)
- **Verdict:** 🟢 Merge
