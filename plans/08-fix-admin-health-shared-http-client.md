# Fix 3.3 — Shared `http.Client` for admin health checks

- **Source:** `PROJECT_ISSUES_REPORT.md` §3.3
- **Severity:** Medium (mitigated in practice — see nuance)
- **Validation:** ✅ Confirmed TRUE
- **Type:** Network performance / resource preservation

## Context

Each external health probe (`checkQdrant`, `checkOpenAI`, `checkOpenRouter`) constructs a fresh
`http.Client` per call. Fresh clients don't share connection pools, so frequent checks can exhaust
ephemeral ports. The fix is a single shared, pre-configured client on the handler struct.

## Validated current state

`internal/api/handlers/admin_health.go`
- Struct at **lines 44–49** — no `http.Client` field:
  ```go
  type AdminHealthHandlers struct {
      DB        *sql.DB
      Redis     *redis.Client
      Cfg       *config.Config
      StartTime time.Time
  }
  ```
- Constructor at **lines 52–59** (`NewAdminHealthHandlers`).
- Per-call clients created at:
  - `checkQdrant` **line 244**
  - `checkOpenAI` **line 299**
  - `checkOpenRouter` **line 352**
  each: `client := &http.Client{Timeout: 5 * time.Second}`.

Nuance: endpoint is `/api/v1/admin/health/detailed` (`routes_admin.go:44`), on-demand only, and
results are cached for 1 minute (`healthCacheTTL`, line 22). Real-world impact is therefore limited,
but the pattern is wasteful and worth fixing.

## Plan

1. Add an `HTTP *http.Client` field to `AdminHealthHandlers`.
2. Initialize it once in `NewAdminHealthHandlers` with a shared transport + timeout:
   ```go
   HTTP: &http.Client{
       Timeout: 5 * time.Second,
       Transport: &http.Transport{
           MaxIdleConns:        20,
           IdleConnTimeout:     90 * time.Second,
           DisableCompression:  false,
       },
   },
   ```
3. Replace the three `client := &http.Client{...}` locals with `h.HTTP` (or a short alias), keeping
   the existing requests unchanged.

## Files to modify
- `internal/api/handlers/admin_health.go`

## Verification
- `go build ./...`.
- Hit `/api/v1/admin/health/detailed` repeatedly; confirm probes still report correct status.
- (Optional) Under repeated calls, confirm ephemeral-port usage stays flat (`ss`/`lsof`) instead of
  growing.
