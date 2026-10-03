# Cycle 5 — Redis Foundation

| | |
|---|---|
| **Status** | Active |
| **Module** | Redis infrastructure (backend only) — EPIC-F, ticket EMPLOYEE36-29 |
| **Depends on** | Cycle 1 (compose, config, health route), Cycle 2 (DB-verified auth middleware, in-memory rate limiter) |
| **Source** | Plane ticket EMPLOYEE36-29 |

This is the scope doc for Cycle 5. It **reverses Cycle 2's "no Redis" non-goal**
(`cycle-02-auth-onboarding.md`): Redis is introduced as optional infrastructure only.

## Objective

Add Redis as a self-hostable infra service with a Clean-Architecture integration, so
later tickets can build on it. **Foundation only: no consumers.**

## In scope

- [ ] `redis` service in `docker-compose.yml` and `backend/docker-compose.yml` (pinned
      `redis:8.10.2-alpine`, `--requirepass`, AOF, healthcheck, resource limits, named
      volume `employee360_redis_data`, `${REDIS_PORT:-6379}` mapping)
- [ ] `RedisConfig` (`REDIS_*`, optional `REDIS_URL`), defaults, validation (port range,
      pool size, password required in production/staging when enabled), `.env.example`,
      `config.yaml.example`
- [ ] `service.Cache` port (`Get/Set/Delete/Incr/Ping`; ttl required; `ErrCacheMiss`,
      `ErrCacheUnavailable`) and `service.CachePinger`
- [ ] `service.CacheKey` / `GlobalCacheKey` tenant-namespaced key helpers
- [ ] go-redis adapter + `NoopCache` in `internal/infrastructure/service/`
- [ ] Health: `redis` field (`ok` / `unreachable` / `disabled`); never changes HTTP status
- [ ] DI (`container.New` takes an optional cache), startup connect + shutdown close in `cmd/api`
- [ ] Archtest: go-redis importable only under `internal/infrastructure/`
- [ ] CI: `redis` service in the integration job; integration tests (`-tags=integration`)
- [ ] Docs: README, `shared-context.md`, `plan/architecture/backend.md`

## Failure-mode policy

- **Disabled** (`REDIS_ENABLED=false`, default): `NoopCache`; health says `disabled`.
- **Enabled but down at startup**: fail fast (exit non-zero).
- **Down at runtime**: cache calls return errors; health says `unreachable` (HTTP 200).
  Each consumer decides: rate limiting falls back to the in-memory limiter
  (`middleware/ratelimit.go`); identity caching fails closed.

## Non-goals

- Any consumer: rate limiting, OTP counters, JWT revocation/identity caching (follow-up tickets).
- Redis Cluster/Sentinel, pub/sub, frontend changes.

## Invariants

- Tenant-owned keys must use `CacheKey(tenantID, ...)` with a tenantID sourced from the
  authenticated context by the handler (never client input); a nil tenant is rejected.
- The domain `Cache` port exposes no Redis types.
