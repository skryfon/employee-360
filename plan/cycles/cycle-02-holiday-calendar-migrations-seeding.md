# Cycle 2 — Holiday Calendar: Migrations & Seeding

| | |
|---|---|
| **Status** | Planned — blocked on Cycle 1 |
| **Module** | Holiday Calendar Management (this cycle covers schema + seed data only) |
| **Depends on** | `plan/cycles/cycle-01-project-setup.md` (backend must run, DB must be reachable, `cmd/migrate` runner must work) |
| **Source** | `plan/architecture/backend.md` (entity list, migration file list), `requirmement.md` |

This is the scope/status doc for Cycle 2. Read this before starting or resuming work.
Cross-cutting principles (open source, multi-tenant, self-hostable, headless) live in
`CLAUDE.md`, not here.

**Scope note:** the Holiday Calendar module is delivered across *multiple* cycles, not
one. This cycle is **migrations + seeding only** — no handlers, no usecases, no routes,
no frontend. The next cycle(s) for this module (backend API, then frontend) will get
their own cycle files, numbered and scoped when reached — don't pull that work forward
into this one.

---

## Objective

Land the database schema and bootstrap seed data the Holiday Calendar module needs, so
later cycles can build usecases/handlers/UI against a real, migrated database instead of
guessing at the shape.

---

## Sub-Features

### Migrations (`backend/migrations/`, via the `create-migration` skill)

In dependency order, per `plan/architecture/backend.md`:

- [ ] `000001_create_tenants` — no `tenant_id` column (this is the one table that doesn't get one)
- [ ] `000002_create_departments` — `tenant_id` FK + index
- [ ] `000003_create_positions` — `tenant_id` FK + index
- [ ] `000004_create_users` — `tenant_id` FK + index; FKs to department/position
- [ ] `000005_create_roles` — `tenant_id` FK + index
- [ ] `000006_create_user_roles` — join table, FKs to users + roles
- [ ] `000007_create_holiday_categories` — `tenant_id` FK + index
- [ ] `000008_create_holidays` — `tenant_id` FK + index; FK to holiday_categories
- [ ] `000009_create_audit_logs` — `tenant_id` FK + index

Each: up + down pair, `created_at`/`updated_at` on every entity table, an index on every
FK column. See the `create-migration` skill for the full invariant checklist.

### Seeding (`backend/internal/infrastructure/database/seeder/`, run via `cmd/bootstrap`)

- [ ] Seed one system tenant
- [ ] Seed default roles (platform `super_admin`, plus baseline tenant roles referenced by `plan/architecture/backend.md`)
- [ ] Seed the platform Super Admin user, credentials from env config (never hardcoded)
- [ ] `seeder_test.go` — verify bootstrap is idempotent (running it twice doesn't duplicate the tenant/roles/admin)

**Done when:** `make migrate` applies all 9 migrations cleanly, `make migrate-down`
reverses them cleanly, and `cmd/bootstrap` (or `make bootstrap`) seeds a working system
tenant + super admin — verifiable with a direct DB query. No API endpoint exists yet to
exercise this; that's the next cycle.

---

## Out of Scope for This Cycle

- Any `internal/usecase/*`, `internal/delivery/http/*` code for these entities — no
  handlers, no routes, no auth middleware.
- Any frontend work.
- Any other planned module (Onboarding, Work Status, Leave Management,
  Courses/Certifications, Benefits, Career Growth, Salary/Taxation, Appraisal, Company
  Policies) — see `CLAUDE.md`'s Execution Model section.

---

## Reference

- Entity list, migration file list, layering: `plan/architecture/backend.md`
- Skill to use for each migration pair: `create-migration` (`.claude/skills/`)
- Agent to use: `backend-agent` (`.claude/agents/`)
- Previous cycle: `plan/cycles/cycle-01-project-setup.md`
