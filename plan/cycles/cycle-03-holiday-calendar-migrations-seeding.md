# Cycle 3 — Holiday Calendar: Migrations & Seeding

| | |
|---|---|
| **Status** | Planned — blocked on Cycle 2 |
| **Module** | Holiday Calendar Management (this cycle covers schema + seed data only) |
| **Depends on** | `plan/cycles/cycle-01-project-setup.md` (backend must run, DB must be reachable, `cmd/migrate` runner must work); `plan/cycles/cycle-02-auth-onboarding.md` (tenants/users/roles schema and the `cmd/bootstrap` seeder this cycle extends are created there) |
| **Source** | `plan/architecture/backend.md` (entity list, migration file list), `requirmement.md` |

This is the scope/status doc for Cycle 3. Read this before starting or resuming work.
Cross-cutting principles (open source, multi-tenant, self-hostable, headless) live in
`CLAUDE.md`, not here.

**Renumbered:** this cycle was originally filed as Cycle 2. Cycle 2 was repurposed for
Auth, Email Service & Onboarding Invitations (`plan/cycles/cycle-02-auth-onboarding.md`)
since the auth/tenant middleware and users/roles schema needed to exist before anything
else could be protected — see that file's rationale. Holiday Calendar moved here
unchanged in scope, now correctly sequenced after auth.

**Scope note:** the Holiday Calendar module is delivered across *multiple* cycles, not
one. This cycle is **migrations + seeding only** — no handlers, no usecases, no routes,
no frontend. The next cycle(s) for this module (backend API, then frontend) will get
their own cycle files, numbered and scoped when reached — don't pull that work forward
into this one.

---

## Objective

Land the database schema (and any default seed data) the Holiday Calendar module needs,
so later cycles can build usecases/handlers/UI against a real, migrated database instead
of guessing at the shape. Tenants, departments, positions, users, roles, and audit
logging already exist by this point — landed in Cycle 2 — so this cycle only adds the
two tables specific to this module.

---

## Sub-Features

### Migrations (`backend/migrations/`, via the `create-migration` skill)

- [ ] `000012_create_holiday_categories` — `tenant_id` FK + index
- [ ] `000013_create_holidays` — `tenant_id` FK + index; FK to `holiday_categories`

Each: up + down pair, `created_at`/`updated_at` on every entity table, an index on every
FK column. See the `create-migration` skill for the full invariant checklist.

### Seeding (`backend/internal/infrastructure/database/seeder/`, run via `cmd/bootstrap`)

- [ ] Seed default holiday categories for the system tenant seeded in Cycle 2 (e.g.
      "Public Holiday", "Optional Holiday" — confirm the exact set against
      `requirmement.md` before hardcoding)
- [ ] `seeder_test.go` addition — verify the category seed step is idempotent

**Done when:** `make migrate` applies both migrations cleanly on top of Cycle 2's schema,
`make migrate-down` reverses them cleanly, and `cmd/bootstrap` seeds the default holiday
categories for the system tenant — verifiable with a direct DB query. No API endpoint
exists yet to exercise this; that's a later cycle.

---

## Out of Scope for This Cycle

- Any `internal/usecase/*`, `internal/delivery/http/*` code for these entities — no
  handlers, no routes, no auth middleware.
- Any frontend work.
- Any other planned module (Work Status, Leave Management, Courses/Certifications,
  Benefits, Career Growth, Salary/Taxation, Appraisal, Company Policies) — see
  `CLAUDE.md`'s Execution Model section. (Onboarding invitations are covered by
  Cycle 2, not this one.)

---

## Reference

- Entity list, migration file list, layering: `plan/architecture/backend.md`
- Skill to use for each migration pair: `create-migration` (`.claude/skills/`)
- Agent to use: `backend-agent` (`.claude/agents/`)
- Previous cycle: `plan/cycles/cycle-02-auth-onboarding.md`
