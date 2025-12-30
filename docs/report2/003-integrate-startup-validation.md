# TASK-003 — Integrate Fail-Fast Validation into Server Startup

Goal:
Wire the plan validation logic into the application's boot sequence so the server fails to start if any plan configuration is invalid.

Scope:
- Modify `cmd/server/main.go` inside `newApplication()`.
- Inject `PlanService` and call `ValidateAllPlans()`.
- Log clear error messages and exit the process on failure.

Checklist:
[x] Inspect `cmd/server/main.go` to find the correct insertion point in `newApplication`.
[x] Write an integration test in `internal/integration/startup_test.go` (or similar) that verifies the server fails to start with invalid plan data.
[x] Implement the validation call in `newApplication`.
[x] Add structured logging for validation success/failure.
[x] Ensure the Postgres connection is ready before validation.
[x] Run `make be-run-no-pdf` (or equivalent) to verify manual failure injection.
[x] Run `golangci-lint run cmd/server/main.go`.

Edge Cases:
- Server fails to connect to DB before validation starts.
- Validation fails during a production deploy (captured by logs).

Files Likely to Change:
- `cmd/server/main.go`
- `internal/integration/plan_startup_test.go` (New)
