# CLAUDE.md

Employee360: open-source, self-hostable, multi-tenant employee platform (holidays,
policies, benefits, career, comp/tax, performance).

## Status

Planning docs only (`requirmement.md`, `plan/`) — no code/build/tests yet.

## Dispatch

| Task | Agent | Skill |
|---|---|---|
| Go backend (`backend/`) | `backend-agent` | `create-migration`, `new-backend-feature` |
| React/TS (`clients/*`, `packages/*`) | `frontend-agent` | `new-frontend-feature` |
| Review a branch/PR vs its Plane ticket | — (call directly) | `team-mate-review` |

Rules: `.claude/agents/{backend,frontend}-agent.md`. Don't ad-hoc code in these areas —
dispatch instead. Cross-layer feature: `backend-agent` first (API), then `frontend-agent`.

## Cycles (not "phases")

Each module ships across several narrowly-scoped cycles (migrations → API → frontend,
etc.), each with its own doc: `plan/cycles/cycle-NN-<name>.md` — source of truth for
current scope, read before starting work.

- Cycle 1 (active) — Project Setup & Scaffolding — `cycle-01-project-setup.md`
- Cycle 2 (planned, blocked on 1) — Holiday Calendar migrations+seeding — `cycle-02-holiday-calendar-migrations-seeding.md`
- Later modules (Onboarding, Work Status, Leave, Courses, Benefits, Career Growth, Salary/Tax, Appraisal, Policies): no file until started — don't foreclose them.

New cycle: create `plan/cycles/cycle-NN-<name>.md`, scope narrowly, break into
sub-features before coding.

## Docs

- `plan/cycles/` — current scope (authoritative)
- `requirmement.md` — original requirements (historical)
- `plan/initial-planning.md` — tech stack + early "Design Cycles" notes (unrelated naming to `plan/cycles/` — don't conflate)
- `plan/architecture/{backend,frontend}.md` — full directory trees/architecture
- `AGENTS.md` — equivalent guidance for other AI tools (`.agents/` skills/hooks/rules) — kept in sync with this file; same facts, different format

Prefer real code over these docs once it exists.

## Principles

- **Open Source** — usable/customisable/contributable by any organisation.
- **Multi-Tenant** — `tenant_id` row-scoped server-side, never trusted from client input.
- **Self-Hostable** — `docker-compose`, no cloud-provider lock-in.
- **Headless** — versioned `/api/v1` REST, client-agnostic, no cookie-only auth.
