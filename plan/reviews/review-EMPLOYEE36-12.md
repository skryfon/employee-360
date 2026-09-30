# Review: EMPLOYEE36-12 — B4 — Auth usecases: login, refresh, logout, forgot/reset password

> Branch: ebin/feat/EPIC-B/EMPLOYEE36-12 | Last reviewed: 2026-09-29 13:40 | Iteration: 3 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-12
**State:** started (backlog/in-progress per Plane state group)
**Link:** not returned by Plane API — reference via identifier EMPLOYEE36-12 (child of epic "EPIC-B", sequence B4)
**Cycle:** `plan/cycles/cycle-02-auth-onboarding.md` (Auth Usecases sub-feature)

### Description
Application business logic in `backend/internal/usecase/{interface,implementation}/auth/`, built on B2's domain layer (EMPLOYEE36-10). `LoginUseCase`, `TokenRefreshUseCase`, `LogoutUseCase`, `ForgotPasswordUseCase` (enumeration-safe, publishes `PasswordResetRequested` via the domain `EventPublisher` inside a `Transactor` transaction instead of emailing directly), `ResetPasswordUseCase`.

### Acceptance Criteria
- [x] AC-1: `ForgotPasswordUseCase` returns an identical response for an existing vs. non-existing email — no timing or payload leak. Implementation: `backend/internal/usecase/implementation/auth/forgot_password.go:56-72` (every "doesn't exist / inactive / malformed" branch performs the same dummy `HashToken` call and returns `nil`). Test: `forgot_password_test.go` `TestForgotPasswordUseCase_EnumerationSafe` (nonexistent/invalid-format/empty email) + `TestForgotPasswordUseCase_InactiveUser`.
- [x] AC-2: Raw reset/refresh tokens are never persisted — only their hashes. Implementation: `forgot_password.go:79-91`, `login.go:113-119`, `token_refresh.go:113-119` all call `hashService.HashToken` before storing. Test: `forgot_password_test.go:124-126` and `login_test.go:250-255` assert the stored value is the hash, not the plain token.
- [x] AC-3: Successful password reset revokes every existing refresh token for that user, not just the current session's. Implementation: `reset_password.go:96-99` (`refreshTokenRepo.RevokeAllForUser`). Test: `reset_password_test.go` `TestResetPasswordUseCase_Success` (asserts two independent sessions `rt1`/`rt2` are both revoked).
- [x] AC-4: Every usecase resolves and scopes by `tenant_id` from context — never trusts a client-supplied tenant id. Implementation: `LoginUseCase.Execute` and `ForgotPasswordUseCase.Execute` now take `tenantID uuid.UUID` as an explicit parameter (`usecase/interface/auth/auth.go:10-24`, `:32-37`) and resolve the user via the new tenant-scoped `UserRepository.GetByTenantAndEmail(WithRoles)` (`login.go:68`, `forgot_password.go:72`) instead of the old ambiguous `GetByEmail`. `tenantID` is not a field on `LoginRequest`/`ForgotPasswordRequest` (`types/auth/login.go:8-18`), so it structurally cannot be client-supplied via the request body — it must come from the caller (handler), matching the cycle doc's domain-resolution design. The GORM adapter (`infrastructure/persistence/user_repository.go`) scopes every query by `tenant_id` and never resolves it itself. Test: `login_test.go` `TestLoginUseCase_TenantIsolation`, `forgot_password_test.go` `TestForgotPasswordUseCase_TenantIsolation` (two tenants sharing an email; each request only ever resolves/affects its own tenant's row). `TestGormUserRepository_TenantIsolation` in `infrastructure/persistence/user_repository_test.go` covers the same at the repository layer. Resolved in `0a3353a`.
- [x] AC-5: `ForgotPasswordUseCase` never imports `EmailService` or River directly — only the domain `EventPublisher` port. Structural evidence: `forgot_password.go`'s import block (`domain/entity`, `domain/event`, `domain/repository`, `domain/service`, `usecase/implementation/ucshared`) contains no `infrastructure/service` or River import.

## Latest commit reviewed
`{pending}` — fix(backend): log LoginUseCase/TokenRefreshUseCase best-effort errors

**Scope note:** `origin/main` is one merge behind this stacked branch; the full `origin/main...HEAD` diff also contains EMPLOYEE36-9/10/11's already-reviewed/merged commits. This review is scoped to EMPLOYEE36-12's own commits (`8a28f43`, `0a3353a`, and the pending commit, diffed against `8816b46`, the last commit before this ticket's work started).

## Findings

### 🔴 Critical
- [x] ~~`login.go:64` and `forgot_password.go:68` — ambiguous global email lookup, AC-4 unmet.~~ **Resolved in `0a3353a`.** See AC-4 above.

### 🟡 Major
- [x] ~~`forgot_password.go:110-124` — transaction error silently discarded, nothing logged.~~ **Resolved in `0a3353a`.** A `service.Logger` port (`domain/service/logger_service.go`) was added and injected into `ForgotPasswordUseCaseImpl`; `WithinTransaction`'s error is now checked and logged via `u.logger.Error(...)` before `Execute` still returns `nil` (enumeration-safety preserved). Test: `TestForgotPasswordUseCase_TransactionFailureIsLoggedNotSwallowed` asserts the logger is called exactly once with the transactor's error.
- [x] ~~`login.go:70-72` — inactive check runs before password verification, enumeration leak.~~ **Resolved in `0a3353a`.** The `!user.IsActive` check was moved below both the password-hash-presence check and `ComparePassword`, so a wrong password against an inactive account now yields `ErrInvalidCredentials`, not `ErrUserInactive`. Test: `TestLoginUseCase_InactiveUser_WrongPassword`.
- [x] ~~Missing tenant-isolation test for the AC-4 scenario.~~ **Resolved in `0a3353a`.** See AC-4 above — both usecase-level and repository-level tenant-isolation tests were added.

### 🟢 Minor
- [x] ~~`token_refresh.go:63` (`_ = u.refreshTokenRepo.RevokeFamily(...)`) and `login.go:139` (`_ = u.userRepo.Update(...)` for `LastLoginAt`) — silently discarded errors, no logging.~~ **Resolved in `{pending}`.** A `logger service.Logger` field was added to both `LoginUseCaseImpl` and `TokenRefreshUseCaseImpl` (constructor-injected, same pattern as `ForgotPasswordUseCaseImpl`); both call sites now check the error and call `u.logger.Error(ctx, ...)` before proceeding, while still returning the same outcome to the caller (best-effort side effects are not allowed to fail the primary flow). Test: `TestLoginUseCase_LastLoginUpdateFailureIsLogged`, `TestTokenRefreshUseCase_RevokeFamilyFailureIsLogged`.

## Verdict
- **Score:** 100/100
- **Flag:** 🟢 Merge
- **Notes:** All acceptance criteria are met (including AC-4, the original hard-gate blocker) and every finding from iterations 1–2 is now resolved, including the final 🟢 Minor (silent error discard in `LoginUseCase`/`TokenRefreshUseCase`'s best-effort side effects), fixed by wiring the `service.Logger` port into both usecases with dedicated regression tests. The build is clean under both the default and `integration` tags, `go vet` and `gofmt` are clean, and the full test suite passes (including the two new logging-regression tests). No open findings remain. Note: this ticket's scope is the usecase layer only (per its Description); no handler/DI wiring exists yet to actually supply `tenantID`/construct these usecases with a concrete `Logger` at runtime — that is expected to land in a later story per the cycle doc, not a gap in this review.

## Re-review Log

### Iteration 2 — 2026-09-29 — sha `0a3353a`
- **Resolved:** AC-4 (tenant-ambiguous lookup) — 🔴 Critical; transaction error swallowed — 🟡 Major; inactive-check-before-password-verify enumeration leak — 🟡 Major; missing tenant-isolation test — 🟡 Major.
- **Still open:** silent error discard in `token_refresh.go:63` / `login.go:139` — 🟢 Minor.
- **New issues:** none.
- **Score:** N/A (hard gate) → 99 (+N/A)
- **Verdict:** 🔴 Block → 🟢 Merge

### Iteration 3 — 2026-09-29 — sha `{pending}`
- **Resolved:** silent error discard in `token_refresh.go:63` / `login.go:139` — 🟢 Minor.
- **Still open:** none.
- **New issues:** none.
- **Score:** 99 → 100 (+1)
- **Verdict:** 🟢 Merge → 🟢 Merge
