# TASK-003 — Integrate Fail-Fast Validation into Server Startup

Goal:
Wire the plan validation logic into the application's boot sequence so the server fails to start if any plan configuration is invalid.

Scope:
- Modify `cmd/server/main.go` inside `newApplication()`.
- Inject `PlanService` and call `ValidateAllPlans()`.
- Log clear error messages and exit the process on failure.

Checklist:
[ ] Inspect `cmd/server/main.go` to find the correct insertion point in `newApplication`.
[ ] Write an integration test in `internal/integration/startup_test.go` (or similar) that verifies the server fails to start with invalid plan data.
[ ] Implement the validation call in `newApplication`.
[ ] Add structured logging for validation success/failure.
[ ] Ensure the Postgres connection is ready before validation.
[ ] Run `make be-run-no-pdf` (or equivalent) to verify manual failure injection.
[ ] Run `golangci-lint run cmd/server/main.go`.

Edge Cases:
- Server fails to connect to DB before validation starts.
- Validation fails during a production deploy (captured by logs).

Files Likely to Change:
- `cmd/server/main.go`
- `internal/integration/plan_startup_test.go` (New)
