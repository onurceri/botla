# Fix 2.3 — Panic recovery for async analytics goroutine

- **Source:** `PROJECT_ISSUES_REPORT.md` §2.3
- **Severity:** High
- **Validation:** ✅ Confirmed TRUE
- **Type:** Concurrency bug (unrecovered panic crashes the whole process)

## Context

`trackAnalyticsAsync` spawns `go fn()` with no `defer recover()`. A panic inside `fn()` (e.g. a nil
`AnalyticsRepo`) propagates to the top of the goroutine and terminates the entire Go process. This
runs on every chat request, so it is the highest-risk unprotected goroutine. The codebase already
has a recovery convention elsewhere — apply it here (and to the other unprotected spawns).

## Validated current state

`internal/services/chat_helpers.go` **lines 215–243**; the unguarded spawn is `go fn()` at
**line 241**. `fn` calls `s.AnalyticsRepo.IncrementAnalytics(...)` and
`s.AnalyticsRepo.TrackUnansweredQuery(...)` with **no nil check** on `s.AnalyticsRepo`.

Other goroutines already use recovery (the convention to reuse): `tool_executor.go:117`,
`public.go:417`, `handoff_service.go:184`, `pool.go:61`, `refresh_scheduler.go:95`.

**Other unprotected `go func()` spawns found during validation** (recommend fixing in the same pass):
- `internal/services/privacy_service.go:55` (data export job)
- `internal/services/privacy_service.go:244` (admin export job)
- `internal/processing/queue_manager.go:77` (delayed enqueue)
- `internal/api/handlers/admin_health.go:95` and `:102` (parallel health checks + channel closer)

## Plan

1. In `trackAnalyticsAsync`, wrap the async path in recovery:
   ```go
   if s.SyncAnalytics {
       fn()
   } else {
       go func() {
           defer func() {
               if r := recover(); r != nil && s.Log != nil {
                   s.Log.Error("async_analytics_panic", map[string]any{"panic": fmt.Sprint(r)})
               }
           }()
           fn()
       }()
   }
   ```
2. Prefer a shared helper (e.g. `safeGo(func())` in a common package) so all background spawns get
   uniform recovery + logging; adopt the existing recovery style seen in `tool_executor.go`/`public.go`.
3. Apply the same recovery to the other unprotected spawns listed above (privacy service, queue
   manager, admin health parallel checks).

## Files to modify
- `internal/services/chat_helpers.go` (primary)
- `internal/services/privacy_service.go`, `internal/processing/queue_manager.go`,
  `internal/api/handlers/admin_health.go` (same-pattern hardening)
- Optional: a shared `safeGo` helper package

## Verification
- `go build ./...`, `go vet ./...`.
- Unit test: construct a `ChatService` with a stubbed/nil `AnalyticsRepo` (or inject a panicking
  repo), `SyncAnalytics=false`, call `trackAnalyticsAsync`, and assert the process survives and the
  panic is logged.
- `go test -race ./internal/services/...`.
