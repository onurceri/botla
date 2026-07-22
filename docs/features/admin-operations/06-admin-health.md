> Extracted from: admin-operations.md — Section: 1.6 Admin Health

## 1.6 Admin Health

- [ ] `GET /api/v1/admin/health/detailed?refresh=true` — Detailed system health (`AdminHealthHandlers.GetDetailedHealth`)
  - Checks run in parallel via goroutines with `sync.WaitGroup`
  - Checked dependencies: PostgreSQL, Redis, Qdrant, OpenAI, OpenRouter, Storage (R2/S3)
  - Redis-cached with 1-minute TTL (`healthCacheKey`)
  - Force refresh with `?refresh=true` query param
  - Latency thresholds: Postgres >100ms degraded, Redis >50ms degraded, Qdrant >200ms degraded, OpenAI >500ms degraded, OpenRouter >1000ms degraded
  - Overall status: `healthy` / `degraded` / `unhealthy` (Postgres down = unhealthy)
  - Returns HTTP 503 when unhealthy
  - Returns `version` from `APP_VERSION` env (default `dev`), `uptime`, `environment`

