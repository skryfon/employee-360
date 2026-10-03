# Review: EMPLOYEE36-29 — EPIC-F: Redis Service (foundation)

> Branch: ebin/feat/EPIC-F/EMPLOYEE36-29 | Last reviewed: 2026-10-03 | Iteration: 1 | Verdict: 🔴

## Ticket
**Identifier:** EMPLOYEE36-29
**State:** backlog (Plane `state_group: backlog`), no child stories, no comments
**Link:** Plane project 84ac89ce-4980-4bcb-946f-cd252071f83a
**Cycle:** plan/cycles/cycle-05-redis-foundation.md (foundation only, "no consumers")

### Description
Introduce pinned Redis 8.x as an optional self-hostable service with a Clean-Architecture
integration: compose service, config, Cache port + go-redis adapter, tenant-namespaced keys
with mandatory TTL, health check, DI, tests.

### Acceptance Criteria
- [x] AC-1: `docker compose up` starts pinned Redis (healthy, authenticated); app connects using config only.
  - impl: `docker-compose.yml:26-45`, `backend/docker-compose.yml`, `config/config.go` (RedisConfig); test: `config/config_test.go`, `integration/redis_test.go`
- [x] AC-2: Health route reports Redis status; outage behaviour explicit and tested.
  - impl: `usecase/implementation/health_usecase.go:33-41`; test: `health_usecase_test.go` (disabled/ok/unreachable), `health_handler_test.go`
- [x] AC-3: All keys tenant-namespaced with TTLs; tests prove no cross-tenant reads.
  - impl: `domain/service/cache_key.go`, `infrastructure/service/redis.go:98-110`; test: `redis_test.go:83 TenantKeysIsolated`, `integration/redis_test.go NoCrossTenantReads`, `TestRedisCache_SetRejectsNonPositiveTTL`
- [x] AC-4: Redis types do not leak outside `internal/infrastructure/`.
  - impl: `domain/service/cache.go` (no Redis types); test: `archtest TestRedisClientOnlyImportedByInfrastructure`
- [x] AC-5: Docs updated (README, shared-context, backend architecture). The auth rate-limiting consumer was split out to EMPLOYEE36-33 and removed from this ticket's ACs.

## Latest commit reviewed
`e860291` — Merge branch 'TESTING' into ebin/feat/EPIC-F/EMPLOYEE36-29
(Redis work reviewed as `eca5132^..eca5132`; `origin/main` is stale relative to TESTING, so `origin/main...HEAD` is dominated by unrelated Department/Cycle-2/4 changes.)

## Findings

### 🔴 Critical
- [x] (resolved: AC-5 split into EMPLOYEE36-33, ticket ACs narrowed) Ticket AC-5 vs cycle scope — AC-5 (auth rate limiting consumer) is unmet; `cycle-05-redis-foundation.md` deliberately excludes consumers. Fix: either file the rate-limiting consumer as a follow-up sub-ticket and narrow EMPLOYEE36-29's ACs to the foundation (then AC-5 is out of scope and this merges), or implement the consumer. Needs a product decision; the code itself is sound.

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [x] `.env.example` sets `REDIS_PORT=6380` while compose default, config default, README and CI all use 6379 — pick one.
- [x] `infrastructure/service/redis.go:44-66` — `MaxRetries=1` + 5s dial timeout means `/health` can block ~10s when Redis is down, and `fail()` logs a Warn on every probe. Consider a short health-specific context timeout and Debug-level logging for ping.
- [x] `redis.go:80` — `fail` wraps the cause with `%v`, so `context.Canceled`/`DeadlineExceeded` can't be detected via `errors.Is`; use a second `%w` (Go 1.20+).
- [x] `Cache` port takes raw string keys, so tenant namespacing relies on callers using `CacheKey`; consider a typed key (`type CacheKey string`) when the first consumer lands. `FailFast` calling `os.Exit` from infrastructure is also better kept in `cmd/api`.
  (all four resolved in working tree, uncommitted; Cache port typed key deferred to first consumer ticket, noted in port comment)

## Verdict
- **Score:** 100/100 (all findings resolved)
- **Flag:** 🟢 Merge
- **Notes:** Implementation quality is good: clean layering, tenant-safe key helper with nil-tenant rejection, mandatory TTL, atomic Lua INCR+EXPIRE, no credential leakage in errors, archtest guard, build/vet/tests pass. The only blocker is the mismatch between the ticket's AC-5 and the cycle doc's "no consumers" scope. Resolved by splitting AC-5 into EMPLOYEE36-33; the conflict resolution in merge `e860291` (Cache + Department wiring) is correct.

## Re-review Log
_(empty — first review)_
