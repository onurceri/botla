## TASK-003 — Refactor Evidence Tests

Goal:
Refactor the specific evidence files to use the new `fixtures` package and helpers, eliminating boilerplate.

Scope:
- `internal/integration/analytics_full_coverage_test.go`
- `internal/integration/url_discovery_test.go`
- Replace manual setup code with `fixtures` calls.

Checklist:
[ ] Refactor `internal/integration/analytics_full_coverage_test.go`
[ ] Refactor `internal/integration/url_discovery_test.go`
[ ] Verify both tests pass
[ ] Lint check

Edge Cases:
- `analytics_full_coverage_test.go` does specific analytics setups (manual DB inserts). Ensure helpers don't conflict or obscure necessary details.
- `url_discovery_test.go` sets specific discovery modes. Ensure `CreateChatbot` helper accepts options or config.

Files Likely to Change:
- `internal/integration/analytics_full_coverage_test.go`
- `internal/integration/url_discovery_test.go`
