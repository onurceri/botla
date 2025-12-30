## TASK-002 — Refactor Admin Chatbots List

Goal:
Refactor the complex chatbot listing query in `internal/db/admin_chatbots.go` to use `squirrel`, eliminating manual string concatenation and argument indexing management.

Scope:
Included:
- Refactoring `AdminListChatbots`.
- Creating specific tests for chatbot filtering logic.

Excluded:
- Other methods in `admin_chatbots.go` unless they share private query builders.

Checklist:
- [ ] Create `internal/db/admin_chatbots_test.go`.
- [ ] Implement `TestAdminListChatbots` covering:
    - Filter by `Name` (partial match).
    - Filter by `OrganizationID`.
    - Filter by `OwnerID`.
    - Pagination.
- [ ] Run tests to set baseline.
- [ ] Refactor `AdminListChatbots` to use `squirrel`.
    - CAUTION: This query involves `LEFT JOIN`, `JOIN`, and subqueries/counts.
    - Use `sq.Select(...)` with `.LeftJoin(...)` and `.Join(...)`.
    - Handle the `WHERE` clauses for filters carefully (`sq.Or` for the complex text search if needed, or just `sq.Eq`/`sq.Like`).
    - Note that the original query uses `ILIKE` for name. Squirrel supports `Like` maps or `sq.Expr("name ILIKE ?", val)`.
- [ ] Refactor the `Count` query in the same function to use `squirrel` as well.
- [ ] Run tests to verify correctness.
- [ ] Run `make lint`.

Edge Cases:
- Name filter containing `%` or `_` (ensure proper escaping or acceptance of behavior).
- Joining logic correctness (ensure joins are preserved correctly in the builder).

Files Likely to Change:
- internal/db/admin_chatbots.go
- internal/db/admin_chatbots_test.go (NEW)
