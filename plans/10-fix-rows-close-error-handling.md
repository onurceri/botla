# Fix 4.1 — Handle `rows.Close()` errors

- **Source:** `PROJECT_ISSUES_REPORT.md` §4.1
- **Severity:** Low-Medium
- **Validation:** ✅ Confirmed TRUE
- **Type:** Code quality / resource-leak hygiene

## Context

Many repository functions discard the error from `rows.Close()`. `rows.Close()` rarely fails, but
ignoring it trips `errcheck` and can hide connection/transaction-state problems in high-volume query
loops. Capture and log the error using the codebase's existing logger.

## Validated current state

- **47 instances** of `defer func() { _ = rows.Close() }()` across **16 files** under
  `internal/repository/`. Representative sites:
  - `action_repo.go:47`
  - `admin_repo.go:103, 202, 348, 421, 526`
  - `analytics_repo.go:102, 267, 408, 445`
  - `chatbot_repo.go:181, 211, 233, 572`
  - `conversation_repo.go:126, 185`
  - `source_repo.go:125, 159, 184, 346, 526, 567`
- Logging convention exists: `pkg/logger` (already imported in `organization_repo.go:12`,
  `privacy_repo.go:11`) but not yet used for `rows.Close()` errors.

## Plan

1. Introduce a small shared helper to avoid repeating the closure, e.g. in a repo-internal util:
   ```go
   func closeRows(log *logger.Logger, rows *sql.Rows) {
       if err := rows.Close(); err != nil && log != nil {
           log.Warn("rows_close_error", map[string]any{"error": err.Error()})
       }
   }
   ```
   (Match the actual `pkg/logger` API; repos that already hold a logger field should use it.)
2. Replace each `defer func() { _ = rows.Close() }()` with `defer closeRows(r.log, rows)` (or the
   appropriate logger reference) across the 16 files.
3. Where a repo has no logger, either thread one through or log via a package-level logger consistent
   with existing usage.

> This is mechanical and wide; do it as a single sweep and rely on `go vet`/`errcheck` to confirm no
> sites are missed.

## Files to modify
- 16 files under `internal/repository/` (see list above) + a shared helper location

## Verification
- `go build ./...` and `go vet ./...`.
- Run `errcheck ./internal/repository/...` and confirm the `rows.Close()` warnings are gone.
- Existing repo tests still pass: `go test ./internal/repository/...`.
