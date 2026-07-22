# Botla Backend — Audit Fix Plans

These plans correspond to the findings in [`PROJECT_ISSUES_REPORT.md`](../PROJECT_ISSUES_REPORT.md).
Every finding was **validated against the current codebase** before its plan was written
(line numbers below reflect the code as it exists today, not the report).

**Validation result: 11 / 11 findings confirmed TRUE.** Two plans carry bonus discoveries
flagged during validation (2.1 and 3.2).

## Index

| # | Finding | Severity | Plan |
|---|---------|----------|------|
| 2.1 | Email unique constraint blocks soft-deleted re-signup | Medium-High | [01](01-fix-users-email-soft-delete-constraint.md) |
| 2.2 | `CircuitBreakerManager` unsynchronized map (race/panic) | High | [02](02-fix-circuit-breaker-manager-race.md) |
| 2.3 | Async analytics goroutine has no panic recovery | High | [03](03-fix-async-analytics-goroutine-recovery.md) |
| 2.4 | `BrowserPool` stuck at size 1 after idle reaping | Medium | [04](04-fix-browser-pool-stuck-at-one.md) |
| 2.5 | Redis rate limiter under-counts (second-precision `now`) | Medium | [05](05-fix-redis-ratelimit-precision.md) |
| 3.1 | DB connection churn (low `MaxIdleConns`, no lifetimes) | Medium-High | [06](06-fix-db-connection-pool-tuning.md) |
| 3.2 | N-roundtrip insert in `SaveMessageSources` | Medium | [07](07-fix-save-message-sources-bulk-insert.md) |
| 3.3 | New `http.Client` per admin health check | Medium | [08](08-fix-admin-health-shared-http-client.md) |
| 3.4 | Over-broad `FOR UPDATE` lock in feedback query | Medium | [09](09-fix-feedback-for-update-lock.md) |
| 4.1 | Unchecked `rows.Close()` errors (47 sites) | Low-Medium | [10](10-fix-rows-close-error-handling.md) |
| 4.2 | DB/Qdrant/Redis ports exposed in compose files | Low (dev) / High (prod) | [11](11-fix-docker-compose-exposed-ports.md) |

## Suggested execution order (from the report's roadmap)

**Phase 1 — Concurrency & crash prevention (do first):**
- [02](02-fix-circuit-breaker-manager-race.md) CircuitBreakerManager mutex
- [03](03-fix-async-analytics-goroutine-recovery.md) goroutine panic recovery

**Phase 2 — Infrastructure & scaling:**
- [04](04-fix-browser-pool-stuck-at-one.md) BrowserPool scale-up
- [05](05-fix-redis-ratelimit-precision.md) Redis limiter precision
- [01](01-fix-users-email-soft-delete-constraint.md) partial email index

**Phase 3 — Performance & DB tuning:**
- [06](06-fix-db-connection-pool-tuning.md) pool tuning
- [07](07-fix-save-message-sources-bulk-insert.md) bulk insert
- [08](08-fix-admin-health-shared-http-client.md) shared HTTP client
- [09](09-fix-feedback-for-update-lock.md) `FOR UPDATE OF m`

**Phase 4 — Code quality / hygiene:**
- [10](10-fix-rows-close-error-handling.md) rows.Close error handling
- [11](11-fix-docker-compose-exposed-ports.md) compose port lockdown

## Validation nuances worth knowing up front

- **2.1**: The registration handler's pre-check *also* fails to exclude soft-deleted users, so the
  migration and the handler must be fixed together.
- **3.2**: `SaveMessageSources` inserts a `source_type` column that is **not present in the
  `message_sources` migration** — possible schema drift; verify before converting to bulk insert.
- **2.2**: In steady state the map is mostly read (writes happen at startup), but the health endpoint
  and a lazy-init path still create real races; the fix is still required.
- **3.3**: Real-world impact is softened by a 1-minute result cache and on-demand usage, but the
  pattern is still worth fixing.
