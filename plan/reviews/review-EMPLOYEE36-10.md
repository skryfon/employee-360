# Review: EMPLOYEE36-10 — B2 — Auth domain layer: entities, repo interfaces & core services

> Branch: ebin/feat/EPIC-B/EMPLOYEE36-10 | Last reviewed: 2026-09-29 11:05 | Iteration: 1 | Verdict: 🟡

## Ticket
**Identifier:** EMPLOYEE36-10
**State:** started (backlog/in-progress per Plane state group)
**Link:** not returned by Plane API — reference via identifier EMPLOYEE36-10 (child of epic "EPIC-B: Cycle 2 — Auth Core, Tenant Resolution & Email Service", sequence B2)
**Cycle:** `plan/cycles/cycle-02-auth-onboarding.md`

### Description
Pure domain layer (`backend/internal/domain/`) for auth: entities, repository interfaces, and the cross-cutting services (token, hash, email, event publisher). Zero dependencies on Gin/GORM/SMTP/River — those live in infrastructure (see B3, B8). `user_invitation.go` and its repository interface belong to EPIC-C (C1), not here.

### Acceptance Criteria
- [x] AC-1: Domain package has zero imports of Gin, GORM, an SMTP library, or River
- [x] AC-2: All repository interfaces are consumed (not yet implemented — that's infrastructure) by B4's usecases *(see note below — interpreted as a scope/design-boundary statement, not literally testable in this diff since B4 doesn't exist yet)*
- [x] AC-3: Token service supports both access and refresh claim shapes with `tenant_id` + role claims
- [x] AC-4: `EventPublisher`'s signature takes only domain types (`context.Context`, `event.Event`) — no `*gorm.DB`/`*sql.Tx`/River types leak into this package

**Note on AC-2:** B4 (Auth usecases) is a separate, not-yet-started sibling ticket in this epic, so no usecase in this diff literally "consumes" these interfaces yet — that's inherent to B2 shipping before B4. Marked met on structural evidence instead: every repository interface's method set lines up with what `plan/cycles/cycle-02-auth-onboarding.md`'s Auth/Invitation Usecases section will need (e.g. `RefreshTokenRepository.GetByTokenHash`/`Revoke`/`RevokeFamily` for `TokenRefreshUseCase`/`LogoutUseCase`; `UserRepository.GetByEmailWithRoles` for `LoginUseCase`), and everything compiles. Flag this reasoning if the team wants literal consumption evidence instead.

## Latest commit reviewed
`ba5111f` — feat(auth): implement pure domain layer, cross-cutting services, and unit tests

**Scope note:** `origin/main` is currently at `8da8bff`, one merge behind this stacked branch — the full `origin/main...HEAD` diff also contains EMPLOYEE36-9's already-reviewed/merged commits (`76fddc9`, `511f6c0`, `4f429d8`). This review is scoped to `ba5111f` only, EMPLOYEE36-10's actual contribution.

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [ ] `backend/internal/domain/entity/audit_log.go:22-33` vs `backend/migrations/000008_create_audit_logs.up.sql` — the `AuditLog` entity doesn't match the already-migrated `audit_logs` table: entity has `EntityName`/`EntityID string`/`OldValues string`/`NewValues string`/`IPAddress`/`UserAgent`; the table has `entity_type`/`entity_id UUID`/a single `metadata JSONB` column, and no `ip_address`/`user_agent` columns at all. Whoever implements `AuditRepository` next (B4/audit usecase) will either silently drop `IPAddress`/`UserAgent`/`OldValues`/`NewValues` data or need an unplanned follow-up migration. Fix: align the entity to the existing schema (fold old/new into one `Metadata` field, rename `EntityName`→`EntityType`, type `EntityID` as `uuid.UUID`) or file a migration to add the missing columns before this lands elsewhere.
- [ ] `backend/internal/infrastructure/service/jwt_service.go` — hand-rolled JWS/JWT implementation (manual base64url + HMAC-SHA256 sign/verify) instead of a vetted library (e.g. `golang-jwt/jwt/v5`, already the de-facto standard for Go). No obvious exploit was found in this review (HMAC compare is constant-time via `hmac.Equal`, and the code ignores the token's own `alg` header rather than trusting it, so there's no classic `alg:none`/confusion bypass), but reinventing signing/verification for the auth token stack is unnecessary risk for security-critical code and diverges from `CLAUDE.md`'s stated "Authentication: JWT" stack assumption. Recommend swapping in a maintained library before B4/B5 build on top of this.
- [ ] Ticket-scope creep: this commit also adds `backend/internal/infrastructure/service/{hash_service.go,hash_service_test.go,jwt_service.go,jwt_service_test.go}` — concrete **infrastructure** implementations. B2's own description says infra implementations "live in infrastructure (see B3, B8)", `plan/cycles/cycle-02-auth-onboarding.md`'s "Domain Layer" sub-feature list only covers `backend/internal/domain/`, and the epic's child-issue list (B1–B8) has no ticket that explicitly owns hash/token infra implementations — B3 only covers email/SMTP infra. Not blocking (still Cycle 2, layering is correct: infra depends on domain, not the reverse), but worth reconciling against the epic's ticket breakdown so B4/B5 don't duplicate or re-scope this work.

### 🟢 Minor
- [ ] `backend/internal/domain/entity/role.go:20` vs `backend/migrations/000006_create_roles.up.sql` — `Role.Description` has no matching column in the `roles` table (only `id`, `tenant_id`, `name`). Currently unused elsewhere, so low impact, but will silently fail to persist if a GORM model derives directly from this entity.
- [ ] `backend/internal/domain/entity/base.go` — `BaseEntity`/`TenantBaseEntity` are defined but never embedded by any of the actual entities (each one hand-declares `ID`/`TenantID`/`CreatedAt`/`UpdatedAt` itself) — dead code as of this commit.
- [ ] No unit tests for the small business-logic methods added in this diff: `PasswordResetToken.IsValid()`, `RefreshToken.IsActive()`, `User.FullName()` (`backend/internal/domain/entity/`). Worth a quick edge-case test each (exact-expiry boundary, empty first/last name).
- [ ] `backend/internal/infrastructure/service/jwt_service_test.go:56-102` (`TestJWTService_RefreshTokenFlow`) asserts `UserID`/`TenantID`/`TokenID`/`Family` round-trip but never asserts `Roles`, even though `RefreshTokenClaims.Roles` is part of AC-3's claim shape.
- [ ] `NewJWTService` (`backend/internal/infrastructure/service/jwt_service.go:59`) accepts any `secret` including an empty string with no validation — an empty/misconfigured `JWT_SECRET` would silently produce weak, predictable signatures rather than failing fast at construction.

## Verdict
- **Score:** 80/100
- **Flag:** 🟡 Reviewer call
- **Notes:** All four stated ACs are met (with a documented interpretation call on AC-2, which describes future consumption by a not-yet-started sibling ticket). The domain layer is correctly framework-free (verified by grep and a clean `go build`/`go test`), and `TokenService`/`EventPublisher` shapes match the AC requirements with passing unit tests. The open items are all real but non-blocking: two domain-entity/migration-schema mismatches (`AuditLog`, `Role.Description`) that will bite whoever implements those repositories next, a hand-rolled JWT implementation worth replacing with a vetted library before more of the auth stack is built on it, and infra code (`hash_service.go`/`jwt_service.go`) shipped a ticket ahead of where the epic's breakdown assigns it. None of these block this specific diff from working correctly today, but the schema mismatches and the custom JWT code should be resolved before B4/B5 build on top of them, to avoid rework.

## Re-review Log
_(empty — first review)_
