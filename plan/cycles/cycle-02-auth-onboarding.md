# Cycle 2 — Auth, Email Service & Onboarding Invitations

| | |
|---|---|
| **Status** | Active |
| **Module** | Authentication (admin email+password, employee passwordless), transactional email, and onboarding invitations |
| **Depends on** | `plan/cycles/cycle-01-project-setup.md` (backend must run, DB must be reachable, `cmd/migrate`/`cmd/bootstrap` runners must work) |
| **Source** | `plan/architecture/backend.md` (entity/migration list, layering), `plan/initial-planning.md` (auth flow decisions), `requirmement.md` |

This is the scope/status doc for Cycle 2. Read this before starting or resuming work.
Cross-cutting principles (open source, multi-tenant, self-hostable, headless) live in
`CLAUDE.md`, not here.

**Renumbering note:** this cycle was originally filed as Holiday Calendar migrations +
seeding. That content moved unchanged to
`plan/cycles/cycle-03-holiday-calendar-migrations-seeding.md`. Cycle 2 was repurposed
for auth/onboarding because `plan/cycles/cycle-01-project-setup.md` already deferred
"auth/tenant middleware" to Cycle 2 ("once there's something to protect"), and every
later module needs real users, roles, and a working login before its own API is worth
building.

**Scope note:** unlike Cycle 1's scaffolding-only and Cycle 3's migrations-only slices,
this cycle runs the full stack for this one feature set — migrations through handlers,
**and now the frontend that consumes them** (admin login UI, employee OTP UI, forgot/reset
password UI, admin invitation-management UI, invitation-accept flow) — because auth
without a callable endpoint isn't testable, and the API isn't validated end-to-end until a
real client calls it. Dispatch backend work to `backend-agent` and frontend work to
`frontend-agent` per `CLAUDE.md`; frontend work should start once the corresponding backend
endpoint exists (see dependency notes in the Frontend sub-feature below).

---

## Objective

Stand up real authentication (replacing Cycle 1's middleware stubs), a reusable
transactional email service, and an admin-driven onboarding-invitation flow, so that:

1. Tenant Admins and the platform Super Admin can log in with email + password, via a real
   admin login page.
2. Employees can log in passwordlessly via email, via a real OTP login page.
3. Admin users can recover access via "forgot password", via real forgot/reset password
   pages.
4. Tenant Admins can invite new users (admin or employee) by email; invitees land as
   pending users until they accept, via a real admin invitation-management UI and an
   invitation-accept page.

---

## Decisions made for this cycle

`plan/initial-planning.md` left the employee login mechanism as an open question
("magic link vs OTP vs company SSO"). This cycle decides: **OTP over email**, not magic
links. This is a free choice, not a constraint: `requirmement.md` only requires
"Employee authentication using company email" (`requirmement.md:44-45`) and never
specifies a mechanism — `plan/initial-planning.md:55`'s "magic link" mention was an early
narrative aside, not a locked-in requirement, and line 80 explicitly flagged the mechanism
as open. Reasoning for OTP: it's a 6-digit code typed into the employee client, so it needs no
deep-link route or token-in-URL handling on the frontend — simpler to build and revisit
later without touching the token/email plumbing built here. If product wants magic links
or SSO instead, that's a follow-up cycle, not a blocker to this one.

Email delivery is **synchronous SMTP**, not an outbox/queue pattern. An event
outbox + async worker is more reliable at scale, but Employee360 has no job queue yet
(Cycle 1 didn't scaffold one), so this cycle sends mail directly and synchronously.
Revisit async/outbox delivery in a later cycle if synchronous send proves unreliable in
practice.

**Default provider: Resend, via its SMTP interface** (`smtp.resend.com`), not its
REST API/SDK. Calling a provider through its proprietary SDK would hard-code
`mail_service.go` to one vendor. Going through Resend's SMTP credentials instead
(username `resend`, password = the Resend API key) keeps `mail_service.go` a plain,
provider-agnostic SMTP client — same code path works against Resend, Amazon SES,
Postmark, or a self-hosted relay, just by changing
`SMTP_HOST`/`SMTP_PORT`/`SMTP_USERNAME`/`SMTP_PASSWORD` in config. This satisfies the
**Self-Hostable** principle (`CLAUDE.md`: no cloud-provider lock-in) while still
defaulting to a widely-used, cost-effective provider (Resend: 3,000 emails/month free,
then usage-based) instead of requiring self-hosted SMTP infrastructure this project
doesn't have yet. Local dev points `SMTP_HOST` at Mailpit instead so email is
verifiable without hitting Resend at all.

Roles stay the fixed three-role model from `CLAUDE.md`'s Platform Roles & Governance
Model (`super_admin`, `admin`, `employee`) — no granular permissions table. A
resource/action permission system solves a different (multi-role, per-resource ACL)
problem than this project's fixed role set, so it's not part of this cycle.

`plan/initial-planning.md` also left tenant resolution as an open question ("email
domain, subdomain, or explicit tenant selection at login"). This cycle decides:
**email domain**, resolved via a separate `tenant_domains` table (one tenant to many
domains) rather than a single column on `tenants` — a company invited on `acme.com` may
also send mail from `acme-hr.com` or acquire a second brand's domain later, and a join
table avoids a schema change when that happens. Reasoning for email domain over the
alternatives: subdomain routing needs wildcard DNS configured by whoever is hosting the
instance, which conflicts with the **Self-Hostable** principle (`CLAUDE.md`:
`docker-compose`, no extra infra assumptions) — a self-hoster on a bare IP or a single
domain can't cleanly offer `*.example.com`. Email domain needs no extra infra and
matches the parenthetical already in `plan/initial-planning.md:55`.

DNS-based domain-ownership verification is still out of scope for this cycle: tenant
(and tenant-domain) provisioning is a **Platform Super Admin** action (`CLAUDE.md`'s
governance model — no self-service tenant signup yet), so domains are set by a trusted
operator, not claimed by an untrusted party, and don't need a verification workflow to
prevent spoofing at this stage. Revisit if self-service signup becomes a real
requirement.

---

## Sub-Features

### Migrations (`backend/migrations/`, via the `create-migration` skill)

In dependency order, per `plan/architecture/backend.md` (`000002` and `000008`–`000011`
are new, added by this cycle; `000003`–`000007` keep their original names but shift by
one position to make room for `000002_create_tenant_domains`):

- [ ] `000001_create_tenants` — no `tenant_id` column (this is the one table that doesn't get one). Columns: `id` (uuid, PK), `name` (text), `is_active` (boolean, default true), `created_at`, `updated_at`
- [ ] `000002_create_tenant_domains` — `tenant_id` FK + index; `domain` (text, unique across all tenants — this is what login resolution matches against); `created_at`, `updated_at`
- [ ] `000003_create_departments` — `tenant_id` FK + index
- [ ] `000004_create_positions` — `tenant_id` FK + index
- [ ] `000005_create_users` — `tenant_id` FK + index; FKs to department/position (both nullable — a user can exist before being assigned either); `password_hash` nullable (employees are passwordless); `email_verified_at`, `last_login_at`, `is_active`
- [ ] `000006_create_roles` — `tenant_id` FK + index
- [ ] `000007_create_user_roles` — join table, FKs to users + roles
- [ ] `000008_create_audit_logs` — `tenant_id` FK + index; nullable `actor_user_id` FK (system-initiated actions have no actor)
- [ ] `000009_create_password_reset_tokens` — `tenant_id` FK, `user_id` FK, `token_hash` (never store the raw token), `expires_at`, `used_at`
- [ ] `000010_create_refresh_tokens` — `tenant_id` FK, `user_id` FK, `token_hash`, `family` (uuid, for rotation/revocation), `revoked_at`, `expires_at`, `ip_address`, `user_agent` — access tokens stay fully stateless per `CLAUDE.md`; refresh tokens are tracked so logout/revocation is possible
- [ ] `000011_create_user_invitations` — `tenant_id` FK, `email`, `role_id` FK, nullable `department_id`/`position_id` FK, `invited_by` FK (users), `token_hash`, `expires_at`, `accepted_at`, `revoked_at`

Each: up + down pair, `created_at`/`updated_at` on every entity table, an index on every
FK column. See the `create-migration` skill for the full invariant checklist.

### Domain Layer (`backend/internal/domain/`)

- [ ] Entities: `tenant_domain.go`, `user.go`, `role.go`, `user_role.go`, `password_reset_token.go`, `refresh_token.go`, `user_invitation.go`, `audit_log.go`
- [ ] Repository interfaces: `tenant_domain_repository.go` (includes a `FindTenantByDomain` lookup — how login resolves `tenant_id` from an email's domain), `user_repository.go`, `role_repository.go`, `user_role_repository.go`, `password_reset_repository.go`, `refresh_token_repository.go`, `user_invitation_repository.go`, `audit_repository.go`
- [ ] `service/token_service.go` — JWT issue/verify (access + refresh claims carrying `tenant_id`, `user_id`, roles)
- [ ] `service/hash_service.go` — password hashing (bcrypt) + generic secret-token hashing (sha256, for reset/invitation/refresh token storage — never store raw tokens)
- [ ] `service/email_service.go` — `EmailService` interface (`Send(ctx, EmailMessage) error`) + `EmailMessage`/`EmailTemplateName` types, infrastructure-agnostic (depends only on domain constructs, not SMTP/Resend specifics). Template set for this cycle: `PasswordReset`, `OTPCode`, `UserInvitation`.

### Email Service (`backend/internal/infrastructure/service/`)

- [ ] `mail_service.go` — plain SMTP implementation of `EmailService` (Viper config: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`; TLS as needed) — no Resend SDK, no vendor-specific code
- [ ] `mail/templates/` — subject + text + HTML template per `EmailTemplateName` (`PasswordReset`, `OTPCode`, `UserInvitation`), loaded via Go's `html/template`/`text/template`
- [ ] `.env.example` — default production values pointed at Resend's SMTP endpoint: `SMTP_HOST=smtp.resend.com`, `SMTP_PORT=465` (or `587`), `SMTP_USERNAME=resend`, `SMTP_PASSWORD=<RESEND_API_KEY>`
- [ ] Local dev: `.env` points `SMTP_HOST` at a local catcher (e.g. Mailpit) instead, so email is verifiable without hitting Resend

### Auth Usecases (`backend/internal/usecase/{interface,implementation}/auth/`)

- [ ] `LoginUseCase` — admin/super_admin email + password → access + refresh token pair
- [ ] `RequestOTPUseCase` / `VerifyOTPUseCase` — employee passwordless login (send code, verify code → token pair); reuses `VerifyEmailUseCase` naming already planned in `plan/architecture/backend.md` if it fits, otherwise add alongside it
- [ ] `TokenRefreshUseCase` — rotate refresh token (per `plan/architecture/backend.md`)
- [ ] `LogoutUseCase` — revoke the presented refresh token
- [ ] `ForgotPasswordUseCase` — admin/super_admin only (employees have no password); enumeration-safe (always returns success regardless of whether the email exists), generates + emails a reset token (random bytes, hashed before storage, short expiry)
- [ ] `ResetPasswordUseCase` — consumes a reset token, sets new password, invalidates the token and all existing refresh tokens for that user

### Onboarding Invitation Usecases (`backend/internal/usecase/{interface,implementation}/invitation/`)

- [ ] `InviteUserUseCase` — admin invites by email + role (+ optional department/position); creates a pending `users` row (`is_active = false`, no password) and an invitation row; emails `UserInvitation`. Tenant-scoped: an admin can only invite into their own tenant.
- [ ] `AcceptInvitationUseCase` — consumes the invitation token; for an invited admin, sets a password and activates the user; for an invited employee, just activates the user (they'll use OTP login going forward)
- [ ] `ResendInvitationUseCase` — reissues token + re-sends the email, only while pending
- [ ] `RevokeInvitationUseCase` — admin cancels a pending invitation
- [ ] `ListInvitationsUseCase` — tenant-scoped list for the admin UI (later cycle)

### Middleware (`backend/internal/delivery/http/middleware/`)

- [ ] `auth.go` — real JWT validation + role extraction, replacing Cycle 1's stub
- [ ] `tenant.go` — real tenant resolution from JWT claims into `context.Context`, replacing Cycle 1's stub

### Delivery / Routes (`backend/internal/delivery/http/`)

- [ ] `auth_handler.go` — `POST /api/v1/auth/login`, `POST /api/v1/auth/otp/request`, `POST /api/v1/auth/otp/verify`, `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`, `POST /api/v1/auth/forgot-password`, `POST /api/v1/auth/reset-password` — all unauthenticated except logout
- [ ] `invitation_handler.go` (or fold into `user_handler.go`) — `POST /api/v1/users/invitations` (admin-only), `POST /api/v1/users/invitations/:id/resend`, `DELETE /api/v1/users/invitations/:id`, `GET /api/v1/users/invitations`, and an unauthenticated `POST /api/v1/invitations/accept`
- [ ] Wire `auth.go`/`tenant.go` middleware onto every route above except login/OTP-request/OTP-verify/refresh/forgot-password/reset-password/invitation-accept

### Seeding (`backend/internal/infrastructure/database/seeder/`, run via `cmd/bootstrap`)

- [ ] Seed one system tenant
- [ ] Seed default roles (`super_admin`, `admin`, `employee`)
- [ ] Seed the platform Super Admin user (password from env config, never hardcoded)
- [ ] `seeder_test.go` — verify bootstrap is idempotent (running it twice doesn't duplicate the tenant/roles/admin)

### Frontend (`clients/admin/`, `clients/employee/`, `packages/api-client/`, via the
`new-frontend-feature` skill)

Depends on the corresponding backend endpoint from the sections above; don't start a
frontend item before its backend endpoint is callable.

- [ ] `packages/api-client` — auth + invitation API methods (login, OTP request/verify,
  refresh, logout, forgot/reset password, invite/accept/resend/revoke/list invitations);
  Axios interceptors for attaching the access token and `X-Tenant-ID`, and for silent
  refresh-on-401 using `TokenRefreshUseCase`'s endpoint
- [ ] `clients/admin/src/features/auth/` — login page (email + password), forgot-password
  page, reset-password page; Zustand store for the access token / auth state, TanStack
  Query mutations via the API client
- [ ] `clients/admin/src/features/invitations/` — invite-user form (email + role +
  optional department/position), invitations list (pending/accepted/revoked), resend and
  revoke actions
- [ ] `clients/employee/src/features/auth/` — OTP login page: request-code step, then
  verify-code step; Zustand store for auth state
- [ ] Invitation-accept page — a token-in-URL route (unauthenticated) that calls
  `POST /api/v1/invitations/accept`; for an invited admin, prompts to set a password; for
  an invited employee, activates directly. Lives wherever `plan/architecture/frontend.md`
  places shared/public routes — confirm before building if that's not yet decided.

**Done when:** `make migrate` applies all 11 migrations cleanly, `make migrate-down`
reverses them cleanly, `cmd/bootstrap` seeds a working system tenant + super admin, and
end-to-end **through the UI**: an admin can log in on the admin login page, request a
password reset and complete it via the forgot/reset password pages, invite a new user from
the admin invitation-management UI (invitation email visible in a local SMTP catcher), and
that invitee can complete the invitation-accept page to become an active user who can then
log in (via the admin login page if invited as admin, via the employee OTP page if invited
as employee).

---

## Out of Scope for This Cycle

- SSO / OAuth login for employees — not decided yet, see "Decisions made for this
  cycle".
- Async/outbox-based email delivery and a job queue — deferred; synchronous SMTP is
  this cycle's scope.
- Granular permissions (resource/action ACLs) — out of scope; this project uses the
  fixed three-role model.
- Holiday Calendar, or any other planned module (Work Status, Leave Management,
  Courses/Certifications, Benefits, Career Growth, Salary/Taxation, Appraisal, Company
  Policies) — see `CLAUDE.md`'s Execution Model section.

---

## Reference

- Entity list, migration file list, layering: `plan/architecture/backend.md`
- Auth flow decisions and open questions: `plan/initial-planning.md`
- Skill to use for each migration pair: `create-migration` (`.claude/skills/`)
- Skill to use for the usecase/handler/route scaffolding: `new-backend-feature`
  (`.claude/skills/`)
- Agent to use: `backend-agent` (`.claude/agents/`)
- Previous cycle: `plan/cycles/cycle-01-project-setup.md`
- Next cycle: `plan/cycles/cycle-03-holiday-calendar-migrations-seeding.md`
