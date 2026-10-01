# Review: refactor-identity-flow — Explicit identity flow, audit columns, DB-verified auth middleware

> Branch: ebin/refactor/1.0 (vs origin/TESTING) | Last reviewed: 2026-10-01 | Iteration: 2 | Verdict: 🟢

## Ticket
**Identifier:** none (no Plane ticket; reviewed against `plan/architecture/refactor-identity-flow-handler-to-repo.md` and its "Verification" section)
**State:** n/a
**Cycle:** `plan/cycles/cycle-02-auth-onboarding.md` (Cycle 2, active)

### Acceptance Criteria (from the plan doc's Verification section)
- [x] AC-1: no `middleware.*` in handlers/usecases/infrastructure (only a comment match in `invitation/common.go:51`); `Get*` helpers unexported
- [x] AC-2: no `ctx.*FromContext` / `ctx.With*` in usecase/infrastructure (grep empty)
- [x] AC-3: usecases/repos take explicit `tenantID`/`actorID` (role, audit, invitation, user_role repos); tests updated (`role_repository_test`, `audit_repository_test`, `user_role_repository_test`, `invitation_test`)
- [x] AC-4: handlers resolve identity via `handlers/identity.go`, 401 on missing/invalid (`invitation_handler_test`, `role_handler_test`)
- [x] AC-5: unauthenticated / unusable identity → 401 (`auth_identity_test.go`), DB failure → 503
- [x] AC-6: `go build`, `go vet`, `go test ./internal/...` green (persistence/integration tests were not run — no DB here)

## Latest commit reviewed
`501eec7` — fix(backend): address identity-flow review (single identity query, soft-delete filters, safe 000015 down)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] (resolved in 501eec7) `internal/infrastructure/persistence/tenant_domain_repository.go:41` — `FindTenantByDomain` filters `is_active` but not `tenants.deleted_at` / `tenant_domains.deleted_at`, even though this commit adds those columns and `tenant_repository.go` filters them. A soft-deleted tenant can still start login / forgot-password (reset emails sent, tokens minted); only the new middleware stops it afterwards. Fix: add `tenants.deleted_at IS NULL AND tenant_domains.deleted_at IS NULL` + a test.
- [x] (resolved in 501eec7) `migrations/000015_add_audit_columns.down.sql:5` — down runs `DELETE FROM users WHERE deleted_at IS NOT NULL`: silent data loss, cascades to `user_roles` and tokens, and fails if the user is referenced by `user_invitations.invited_by` (no ON DELETE). Fix: document in the up file why it's irreversible, and null/guard `invited_by` references (or rename live-duplicate emails instead of deleting).
- [x] (resolved in 501eec7) Scope: the plan doc lists "changing the auth/tenant middleware logic" and "schema changes" as **out of scope** and is still `Status: proposed`. This commit adds DB-verified auth + migration 000015 + soft-delete semantics for users. Reasonable and documented in `shared-context.md`, but update the plan doc (scope + status) or split into separate commits/ticket so the history is traceable.

### 🟢 Minor
- [x] (resolved in 501eec7) `000015_add_audit_columns.up.sql` — `deleted_at/deleted_by` added to tenant_domains, departments, positions, roles, user_invitations but only `users`/`tenants` reads filter on it; add the filters when those get soft-delete behaviour (roles still hard-`Delete`).
- [x] (resolved in 501eec7) `middleware/auth.go` — Auth now costs up to 3 queries (tenant, user, roles) per request, with no cache. Fine for now; note for load/perf later.
- [x] (resolved in 501eec7, caution comment only) `middleware/identity_fake_test.go` — `registeringTokenService` auto-registers every minted token as a valid identity, which can mask missing-verifier cases in older tests; only the explicit `mintRaw` tests exercise failures.

## Verdict
- **Score:** 100/100 (0 open findings)
- **Flag:** 🟢 Merge
- **Notes:** All prior findings are resolved. Identity verification is now one joined query (`persistence/identity_reader.go`) scoped by tenant and user, with soft-deleted tenants/users/roles excluded and a uniform 401 / 503 fail-closed behaviour preserved. `deleted_at` filters now cover tenant-domain, role, user-role, invitation and org-reference reads. The 000015 down migration is lossless. Build, vet and unit tests pass (integration suite was reported passing by the fixing agent against local Postgres; not re-run by the reviewer).

## Re-review Log
### Iteration 2 — 2026-10-01 — sha `501eec7`
- **Resolved:** tenant-domain `deleted_at` filter; non-destructive 000015 down; plan doc status/scope; `deleted_at` filters on role/invitation/org-reference reads; single-query identity check; test-wrapper caution comment
- **Still open:** none
- **New issues:** none
- **Score:** 82 → 100 (+18)
- **Verdict:** 🟢 Merge
