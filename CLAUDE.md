# CLAUDE.md

Guidance for Claude Code in this repository. Employee360 is an open-source,
self-hostable, multi-tenant employee platform (holidays, policies, benefits, career
development, comp/tax, performance — one place).

## Project Status

Only planning docs exist (`requirmement.md`, `plan/`) — no source code, build tooling,
or tests yet. No commands to run. Update this file once implementation lands.

## Agent Dispatch

| Task | Agent |
|---|---|
| Go backend (`backend/`) | `backend-agent` |
| React/TS (`clients/admin/`, `clients/employee/`, `packages/*`) | `frontend-agent` |

Full rules live in each agent file (`.claude/agents/{backend,frontend}-agent.md`) — don't
ad-hoc code in these areas without dispatching. For a feature spanning both layers,
dispatch `backend-agent` first (API contract), then `frontend-agent` (wires against it).

## Project Skills

Skills in `.claude/skills/` encode project-specific patterns. Match the task to a skill
before doing it ad-hoc:

| Skill | When to use |
|---|---|
| `create-migration` | Adding a new DB migration (up/down pair) |
| `new-backend-feature` | A new backend resource/endpoint — scaffolds every Clean Architecture layer |
| `new-frontend-feature` | A new page/screen in the admin or employee app — scaffolds the feature folder |
| `team-mate-review` | Reviewing a teammate's branch/PR against its Plane ticket, or preparing work for review |

## Execution Model: Cycles

Work ships as a sequence of **cycles**, not "phases." A module is typically delivered
across several cycles (e.g. migrations, then backend API, then frontend), each narrowly
scoped. Each cycle has its own doc, `plan/cycles/cycle-NN-<name>.md` — read it before
starting or resuming work; it's the source of truth for current scope, not this file.

- **Cycle 1 — Project Setup & Scaffolding** (active) — `plan/cycles/cycle-01-project-setup.md`
- **Cycle 2 — Holiday Calendar: Migrations & Seeding** (planned, blocked on Cycle 1) — `plan/cycles/cycle-02-holiday-calendar-migrations-seeding.md`
- Further modules (Onboarding, Work Status, Leave Management, Courses/Certifications, Benefits, Career Growth, Salary/Taxation, Appraisal, Company Policies): no cycle file until work starts on them — don't build earlier cycles in a way that forecloses them.

New cycle: create `plan/cycles/cycle-NN-<name>.md` (next number), scope it narrowly,
break it into sub-features before writing code.

## Planning Docs

- `plan/cycles/` — authoritative per-cycle scope/status.
- `requirmement.md` — original requirements doc (historical source).
- `plan/initial-planning.md` — tech stack tables + early "Design Cycles" notes (unrelated to `plan/cycles/` despite similar naming — don't conflate).
- `plan/architecture/backend.md` / `frontend.md` — full directory trees and detailed architecture.

Once code exists, prefer reading it over these docs where they diverge.

## Key Principles

- **Open Source** — usable/customisable/contributable by any organisation.
- **Multi-Tenant** — every tenant-owned table row-scoped by `tenant_id`, resolved server-side from the verified JWT/domain, never trusted from client input.
- **Self-Hostable** — `docker-compose` for local/self-hosted deploy, no cloud-specific managed services.
- **Headless** — versioned REST API (`/api/v1/...`), client-agnostic (works for web, mobile, other clients — no cookie-only auth).
