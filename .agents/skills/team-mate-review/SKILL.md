---
name: team-mate-review
description: >-
  Standardized code review checklist and verification skill for Employee360.
  Use when reviewing a PR, branch, or feature implementation for architectural compliance and security.
---

# Code Review & Compliance Checklist

Perform a comprehensive review against Employee360's architectural invariants:

## 1. Multi-Tenancy Review (CRITICAL)
- [ ] Every database table (except `tenants`) has a `tenant_id uuid NOT NULL` column.
- [ ] Every database table has indexes/unique constraints prefixed or composite with `tenant_id`.
- [ ] Every persistence query in `internal/infrastructure/persistence/` retrieves `tenant_id` from `ctx.TenantIDFromContext(ctx)` and scopes SQL operations by it.
- [ ] Handlers and repositories do not accept or trust client-supplied `tenant_id` parameters.

## 2. Clean Architecture Review
- [ ] No `database/sql`, GORM, or Gin imports in `internal/domain/`.
- [ ] Handlers do not call repositories directly; all operations delegate to `internal/usecase/`.
- [ ] Responses use `internal/delivery/http/response/` helper envelopes.

## 3. Frontend Architecture Review
- [ ] Feature files are placed in `clients/*/src/features/<feature>/`.
- [ ] Server state uses TanStack Query; UI state uses Zustand.
- [ ] API requests use `@employee360/api-client` (with automatic `X-Tenant-ID` and JWT interceptors).

## 4. Verification Commands
```bash
# Backend checks
cd backend && go test ./... && go vet ./...

# Frontend checks
pnpm typecheck
pnpm build
```
