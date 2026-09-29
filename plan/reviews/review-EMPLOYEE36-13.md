# Review: EMPLOYEE36-13 — B5 — Auth & tenant middleware (replacing Cycle 1 stubs)

> Branch: ebin/feat/EPIC-B/EMPLOYEE36-13 | Last reviewed: 2026-09-29 | Iteration: 2 | Verdict: 🟢

## Ticket
**Identifier:** EMPLOYEE36-13
**State:** started (backlog/in-progress per Plane state group)
**Link:** not returned by Plane API — reference via identifier EMPLOYEE36-13 (child of epic "EPIC-B", sequence B5)
**Cycle:** `plan/cycles/cycle-02-auth-onboarding.md` (Middleware sub-feature, `backend/internal/delivery/http/middleware/`)

### Description
Real JWT validation and tenant resolution, replacing the middleware stubs left in place by Cycle 1 (EPIC-A / A3). Every later module's multi-tenant isolation invariant depends on this. Scope: `auth.go` (real JWT validation + role extraction) and `tenant.go` (real tenant resolution from JWT claims into `context.Context`).

### Acceptance Criteria
- [x] AC-1: A request with a missing/invalid/expired access token is rejected before reaching a handler. Implementation: `backend/internal/delivery/http/middleware/auth.go:23-58` (`Auth` aborts with 401 on missing header, malformed `Bearer` scheme, and any `ValidateAccessToken` error, distinguishing expired vs. otherwise-invalid). Test: `auth_test.go` `TestAuth_MissingHeader`, `TestAuth_InvalidHeaderFormat` (5 subcases), `TestAuth_InvalidTokenSignature`, `TestAuth_ExpiredToken`, `TestAuth_RefreshTokenRejectedAsAccessToken` — all assert the downstream handler is never reached (401 returned, handler-only fields absent).
- [x] AC-2: `tenant_id` in context comes exclusively from validated JWT claims — never from a request header/body/query param. Implementation: `tenant.go:20-53` (`Tenant` only ever reads from Gin/Go context or parsed claims — no header/query/body lookups exist in the function). Test: `tenant_test.go` `TestTenant_StrictIsolation_IgnoresClientSuppliedTenantID` — sends a spoofed tenant via `X-Tenant-ID`, `Tenant-ID`, the project's own tenant header constant, a `?tenant_id=` query param, and a JSON body field simultaneously, and asserts the resolved `tenant_id` matches only the JWT's tenant, on both GET and POST.
- [x] AC-3: Role extraction supports the fixed three-role model (`super_admin`/`admin`/`employee`) for downstream role checks. Implementation: `auth.go:96-118` (`RequireRole`/`HasRole`/`HasAnyRole`) built on `entity.RoleSuperAdmin`/`RoleAdmin`/`RoleEmployee` (pre-existing constants in `domain/entity/role.go`). Test: `auth_test.go` `TestRequireRole_AccessControl` (7 subcases covering all three roles against admin-only/super-admin-only/employee-only endpoints, both allow and deny paths) + `TestRoleHelpers`.

All three ACs have both implementation and test evidence; `go test ./internal/delivery/http/middleware/...` passes (24/24), `go build ./...` and `go vet` are clean.

## Latest commit reviewed
`<pending-commit-sha>` — fix(backend): address EMPLOYEE36-13 review findings (RequireRole 401-vs-403; drop unused Require* aliases) — changes are staged/unstaged in the working tree as of this update; fill in the actual sha once committed.

Previously reviewed: `cf5d8cd` — feat(backend): implement JWT auth and tenant resolution middleware (EMPLOYEE36-13)

Note: the branch also carries unmerged commits from prior tickets (EMPLOYEE36-9 through 12, already reviewed in their own `plan/reviews/review-EMPLOYEE36-*.md` files) plus `cb96e24` ("fix"), which is a fix-up for EMPLOYEE36-12's review findings (touches `login.go`/`token_refresh.go`/`review-EMPLOYEE36-12.md`), not this ticket. This review scopes strictly to `cf5d8cd`, the only commit touching this ticket's declared files (`auth.go`, `auth_test.go`, `tenant.go`, `tenant_test.go`).

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [x] `auth.go:96-101` — `RequireRole` returns 401 "authentication required" (not 403) when an authenticated user's role list is empty — why: it can't distinguish "no auth context at all" from "authenticated but zero roles assigned," since both produce `len(roles) == 0`. Under the current three-role model every active user should always carry exactly one role, so this is unreachable in practice today, but if a user is ever left role-less (e.g. a bug in the onboarding/role-assignment path), they'd see a misleading "please authenticate" response instead of "insufficient permissions." Suggested fix: have `Auth` set an explicit "authenticated" marker (or check for `ContextKeyClaims`/`ContextKeyUserID` presence) in `RequireRole` before falling back to the "unauthenticated" branch. (resolved in `<pending-commit-sha>`)
- [x] `auth.go:85-88`, `tenant.go:67-69` — `RequireAuth`/`RequireTenant` are unused aliases for `Auth`/`Tenant` (not referenced anywhere in `backend/`, including tests). Not wrong, but until the route-wiring ticket (`plan/cycles/cycle-02-auth-onboarding.md` line 198, "Wire auth.go/tenant.go middleware onto every route") lands and picks one naming convention, this is unused exported surface. No action needed now — just something for the wiring ticket to resolve (drop whichever name isn't adopted). (resolved in `<pending-commit-sha>`)

## Verdict
- **Score:** 100/100
- **Flag:** 🟢 Merge
- **Notes:** Clean, well-scoped implementation of exactly what B5 calls for — nothing more. All three ACs are demonstrably met with targeted tests, including a strong adversarial test for tenant-spoofing (header/query/body all attempted, JWT wins). Layering is respected (middleware depends only on the `domain/service.TokenService` port, `domain/errors`, and the pre-existing `ctx`/`response` helpers — no GORM/Gin leaks into domain, no direct DB access). Route wiring is intentionally deferred to a separate cycle-02 checklist item and correctly not attempted here. Both minor findings from iteration 1 are now resolved with no open findings remaining.

## Re-review Log
### Iteration 2 — 2026-09-29 — sha `<pending-commit-sha>`
- **Resolved:** both 🟢 Minor findings (RequireRole 401-vs-403 for zero-role authenticated users; unused RequireAuth/RequireTenant aliases)
- **Still open:** none
- **New issues:** none
- **Score:** 98 → 100 (+2)
- **Verdict:** 🟢 Merge
