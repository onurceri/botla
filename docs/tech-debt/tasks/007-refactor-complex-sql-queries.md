# TASK-007 — Refactor Complex SQL Queries with Query Builder

## Goal
Replace manual string concatenation in `AdminListChatbots` with a type-safe SQL builder (Squirrel) to improve maintainability and safety.

## Scope
- `internal/db/admin_chatbots.go`

## Checklist
- [ ] Add `github.com/Masterminds/squirrel` dependency.
- [ ] Create a reproduction test for `AdminListChatbots` (if not exists) to ensure baseline behavior.
- [ ] Refactor `AdminListChatbots` to use `squirrel.Select(...)`.
    - Handle dynamic `Where` clauses for filters.
    - Handle Pagination.
    - Handle Sorting.
- [ ] Verify the generated SQL is correct and comparable to original.
- [ ] Run existing tests for Admin Chatbots.

## Edge Cases
- Empty filters.
- SQL injection prevention (builder handles this, but verify usage).
- Complex joins or subqueries (Squirrel supports these, but check syntax).

## Files Likely to Change
- `internal/db/admin_chatbots.go`
- `go.mod`
