# TASK-002 — Implement Batch Plan Validation in PlanService

Goal:
Add a service-level method to fetch all active plans from the database and run the strict validation logic on each of them.

Scope:
- Modify `internal/services/plan_service.go` to add `ValidateAllPlans()`.
- Use existing `fetchAllPlans` logic to load configurations.
- Return a combined error if any plan fails validation.
- Add integration tests for plan-wide validation.

Checklist:
[ ] Understand `PlanService` architecture and cache interaction.
[ ] Identify failure scenarios (database connection error, malformed JSON in DB, invalid plan data).
[ ] Create `internal/services/plan_service_validation_test.go`.
[ ] Write integration tests that mock/use a test DB containing a plan with an invalid JSON config.
[ ] Implement `ValidateAllPlans(ctx context.Context) error` in `internal/services/plan_service.go`.
[ ] Ensure the method catches the validation error from Task 001.
[ ] Run `go test ./internal/services/...` and ensure all pass.
[ ] Run `golangci-lint run ./internal/services/...`.

Edge Cases:
- Database query returns no plans (application should probably warn or fail).
- A plan exists in the DB but has a `null` config blob.
- Multiple plans fail validation (return an aggregate error).

Files Likely to Change:
- `internal/services/plan_service.go`
- `internal/services/plan_service_validation_test.go` (New)
