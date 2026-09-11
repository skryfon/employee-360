# Employee360

An open-source, self-hostable, multi-tenant employee platform — holidays, company
policies, benefits, career growth, compensation/tax, and performance, all in one place.

Phase 1 focuses on a **Holiday Calendar Management System**: admins configure and manage
company holidays; employees log in with their company email to view them.

## Status

📋 **Planning stage — no code yet.** The repo currently holds requirements and
architecture docs only (`requirmement.md`, `plan/`). Nothing below has been
scaffolded; see [Documentation](#documentation) for what's authoritative and
[Commands](#commands-planned) for what's planned but not yet real.

Work ships in narrow **cycles**, not whole modules at once — see
[`plan/cycles/`](plan/cycles/) for current scope. Cycle 1 (Project Setup &
Scaffolding) is active.

## Table of Contents

- [Status](#status)
- [Principles](#principles)
- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Commands (planned)](#commands-planned)
- [Documentation](#documentation)

## Principles

- **Open Source** — usable, customisable, and contributable by any organisation.
- **Multi-Tenant** — `tenant_id` row-scoped server-side, never trusted from client input.
- **Self-Hostable** — `docker-compose`, no cloud-provider lock-in.
- **Headless** — versioned `/api/v1` REST, client-agnostic, no cookie-only auth.

## Tech Stack

| | |
|---|---|
| **Backend** | Go · Gin · GORM + PostgreSQL · `golang-migrate` · JWT · Viper · Zerolog · Clean Architecture (Hexagonal / Ports & Adapters) |
| **Frontend** | React + TypeScript + Vite · Tailwind CSS · TanStack Query · Zustand · React Hook Form + Zod · `pnpm` workspaces |

See [`plan/architecture/overview.md`](plan/architecture/overview.md) for diagrams,
[`plan/architecture/backend.md`](plan/architecture/backend.md) and
[`plan/architecture/frontend.md`](plan/architecture/frontend.md) for full directory
trees, and rendered diagrams in [`plan/architecture/diagrams/`](plan/architecture/diagrams/).

## Prerequisites

No `go.mod` / `package.json` exist yet to pin exact versions — these are the tools
you'll need once Cycle 1 scaffolding lands, per the target stack above:

- **Go** — for `backend/` (Gin, GORM, `golang-migrate`)
- **Node.js** + **pnpm** — for the `clients/*` and `packages/*` workspace
- **PostgreSQL** — the application database (run via `docker-compose` locally, or a
  native install)
- **Docker** + **Docker Compose** — for local infra (Postgres, and the API once
  containerised)

Exact minimum versions will be pinned in `backend/go.mod` and the root
`package.json`/`.nvmrc` as part of Cycle 1 — treat the above as the toolchain to have
installed, not a version contract yet.

## Commands (planned)

These match the target `Makefile` / workspace scripts described in the architecture
docs. They don't exist yet — treat this as the contract to build toward, not something
you can run today.

### Backend (`backend/`)

```bash
make dev            # start the dev server
make migrate        # apply DB migrations
make migrate-down   # reverse the last migration
make bootstrap       # seed system tenant, roles, and platform Super Admin
make test            # run tests
```

### Frontend (`clients/`, `packages/`)

```bash
pnpm install               # install workspace dependencies
pnpm --filter admin dev    # run the admin portal dev server
pnpm --filter employee dev # run the employee portal dev server
pnpm build                 # build all workspace packages/apps
pnpm typecheck              # type-check all workspace packages/apps
```

## Documentation

- [`shared-context.md`](shared-context.md) — facts and rules that apply regardless of
  which AI tool is reading them; start here for the full project context.
- [`plan/cycles/`](plan/cycles/) — current, authoritative per-cycle scope.
- [`requirmement.md`](requirmement.md) — original business scope and requirements.
- [`plan/architecture/`](plan/architecture/) — target backend/frontend architecture
  (directory trees, diagrams, design rationale).
- [`CLAUDE.md`](CLAUDE.md) / [`AGENTS.md`](AGENTS.md) — AI coding assistant
  configuration and dispatch rules for this repo.

Prefer real code over these docs once it exists.
