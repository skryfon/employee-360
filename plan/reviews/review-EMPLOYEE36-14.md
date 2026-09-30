# Review: EMPLOYEE36-14 — B6 — Auth delivery/routes

> Branch: ebin/feat/EPIC-B/EMPLOYEE36-14 | Last reviewed: 2026-09-29 | Iteration: 3 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-14
**State:** started
**Link:** not returned by Plane API — reference via identifier EMPLOYEE36-14 (child of epic EPIC-B)
**Cycle:** `plan/cycles/cycle-02-auth-onboarding.md` (Handlers/routes sub-feature)

### Description
Gin handlers and route registration for B4's auth usecases. Handlers must never call GORM/DB directly — usecase layer only. Wire B5's auth/tenant middleware on logout only.

### Acceptance Criteria
- [x] AC-1: All 5 routes registered under `/api/v1/auth/` and reachable via curl — impl `routes.go:49-61`; tests `routes_test.go:172,195,209,223,238,252`
- [x] AC-2: Only `logout` requires a valid access token; other 4 reachable unauthenticated — impl `routes.go:57`; tests `routes_test.go:172-236` (unauth 200) and `:238` (logout 401 without token), `:252` (logout OK with token)
- [x] AC-3: Handlers serialize requests/responses only — no direct DB access — impl `auth_handler.go` (usecase interfaces only, no gorm import); tests `auth_handler_test.go` use fake usecases

## Latest commit reviewed
`8746626` — fix(backend): address EMPLOYEE36-14 review findings (iteration 3 reviewed the UNCOMMITTED working tree on top of this sha; commit before merge)

(Branch is stacked on unmerged EMPLOYEE36-9..13; only commit 0887e2d reviewed here.)

## Findings

### 🔴 Critical
- [x] (resolved in 8746626) `backend/internal/delivery/http/handlers/auth_handler.go:184-191` — ForgotPassword takes the tenant from the client-supplied `X-Tenant-ID` header. The route is unauthenticated, so `middleware.GetTenantID` never succeeds and the header is the only source. This violates Invariant 1 (never trust client-supplied tenant IDs); `auth_handler_test.go:408` codifies it. If the header is absent, `tenantID` is `uuid.Nil` and no reset email is ever sent, so the real flow is also broken. Fix: resolve tenant server-side from the email's domain via the `tenant_domains` lookup (`FindTenantByDomain`, per cycle-02 "login resolves tenant from email domain") inside the usecase; drop the header and the `tenantID` parameter.
- [x] (resolved in 8746626) `backend/internal/infrastructure/persistence/user_repository.go:92` (+ `usecase/implementation/auth/login.go:73`) — new `GetByEmailWithRoles` queries `users` by email with no tenant scope. Uniqueness is only `(tenant_id, email)` (`000005_create_users.up.sql:19`), so the same email in two tenants makes `First()` return an arbitrary row: wrong-tenant login or lockout of a legitimate user. It also bypasses the designed tenant_domains resolution. Fix: resolve tenant from email domain, then call `GetByTenantAndEmailWithRoles(ctx, tenantID, email)`; remove the unscoped method. Add a two-tenants-same-email test.

### 🟡 Major
- [x] (resolved in working tree, uncommitted) `backend/internal/infrastructure/persistence/tenant_domain_repository.go:33` — new adapter has no test at all (not even an integration-tagged one). `FindTenantByDomain` is now the sole pre-auth tenant resolver for login and forgot-password, so its behaviour (unknown domain -> ErrTenantNotFound, inactive tenant excluded, case-insensitive match, correct tenant returned when two tenants have domains) should be pinned. Fix: add an integration test alongside the other persistence tests.
- [x] (resolved in 8746626) `backend/internal/types/auth/token_refresh.go:7-8`, `auth_handler.go:104-110` — Refresh accepts client-supplied `ip_address`/`user_agent` and only falls back to server values when empty, so audit metadata on refresh tokens is spoofable. This commit made Login server-derived only; do the same here (remove the fields from the request, pass server-derived values as explicit params).

### 🟢 Minor
- [x] (resolved in working tree, uncommitted: unused methods removed) `tenant_domain_repository.go:56-93` — `Create`, `GetByID`, `Delete` are unused and not tenant-scoped (`Delete`/`GetByID` by id alone). Nothing calls them yet, but before the super-admin tenant-domain cycle wires them, scope or restrict them. Consider deferring them until needed.
- [x] (resolved in 8746626) `auth_handler.go:117` — Refresh returns `err.Error()` to the client; use fixed messages like the other branches.
- [x] (resolved in 8746626) `usecase/implementation/auth/logout.go` — authenticated logout revokes whatever refresh token is presented without checking it belongs to the caller's user/tenant from context. Low risk (needs the token), but cheap to bind.
- [x] (resolved in 8746626) `routes.go:8-10` — stray comment about a blank swagger import sits above the zerolog import; move or remove.

## Verdict
- **Score:** 100/100
- **Flag:** 🟢 Merge (once the working-tree changes are committed)
- **Notes:** Iteration 3: new FindTenantByDomain integration test (4 subtests, passed against live Postgres per the fixing agent; not re-run by me) and unused unscoped tenant-domain methods removed; build/vet (incl. integration tag)/tests pass in my run. Earlier: All prior findings resolved: tenant now resolved server-side from email domain for login and forgot-password, cross-tenant email lookup removed, refresh IP/UA server-derived, logout bound to caller. Remaining: one major (no test for the new tenant-domain adapter) keeps the flag at Reviewer call; go build/vet/test pass. Original iteration-1 notes: All three ACs are met with implementation and test evidence, and routing/auth-middleware wiring is correct (only logout is protected). Two tenant-isolation problems block merge: the unauthenticated forgot-password flow trusts a client `X-Tenant-ID` header, and login uses a cross-tenant email lookup that is ambiguous when an email exists in more than one tenant. Both should be fixed by resolving the tenant from the email domain server-side. Also make refresh IP/UA server-derived.

## Re-review Log
### Iteration 2 — 2026-09-29 — sha `8746626`
- **Resolved:** 2 critical (forgot-password X-Tenant-ID, unscoped login lookup), 1 major (refresh IP/UA), 3 minor
- **Still open:** none from iteration 1
- **New issues:** 🟡 no test for new tenant_domain adapter; 🟢 unused unscoped Create/GetByID/Delete in adapter
- **Score:** 47 → 94 (+47)
- **Verdict:** 🟡 Reviewer call

### Iteration 3 — 2026-09-29 — working tree on top of `8746626` (uncommitted)
- **Resolved:** 🟡 missing tenant_domain adapter test; 🟢 unused unscoped Create/GetByID/GetByDomain/ListByTenantID/Delete removed
- **Still open:** none
- **New issues:** none
- **Score:** 94 → 100 (+6)
- **Verdict:** 🟢 Merge (commit the changes first)
