## TASK-001 — Setup Squirrel & Refactor Admin Audit Logs

Goal:
Introduce the `squirrel` library to the project and replace manual string concatenation in `internal/db/admin_audit.go` with fluent SQL construction.

Scope:
Included:
- Adding `github.com/Masterminds/squirrel` dependency.
- Creating integration tests for `ListAuditLogs`.
- Refactoring `ListAuditLogs` to use `squirrel`.

Excluded:
- Any changes to `admin_chatbots.go`.

Checklist:
- [ ] Add `github.com/Masterminds/squirrel` to `go.mod` (run `go get`).
- [ ] Create `internal/db/admin_audit_test.go`.
- [ ] Implement `TestListAuditLogs` covering:
    - No filters (check default ordering/limit).
    - Filter by `AdminUserID`.
    - Filter by `Action`.
    - Filter by `DateRange`.
- [ ] Run tests to ensure current implementation passes (sets the baseline).
- [ ] Refactor `ListAuditLogs` in `internal/db/admin_audit.go` to use `squirrel.StateMentBuilder`.
    - Use `sq.StatementBuilder.PlaceholderFormat(sq.Dollar)`.
    - Construct the SELECT query using `.From().Where().OrderBy().Limit().Offset()`.
- [ ] Run tests to verify the refactoring matches the baseline behavior.
- [ ] Run `make lint` to ensure no linting errors.

Edge Cases:
- Empty filter struct (should return all).
- StartDate > EndDate (logic in handler usually, but query should handle valid dates).
- SQL injection characters in string fields (should be handled by placeholders). Note: Squirrel handles this if placeholders are used correctly.

Files Likely to Change:
- go.mod
- go.sum
- internal/db/admin_audit.go
- internal/db/admin_audit_test.go (NEW)
