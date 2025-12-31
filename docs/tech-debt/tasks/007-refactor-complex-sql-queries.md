# TASK-007 — Refactor Complex SQL Queries with Query Builder

## Goal
Replace manual string concatenation in `AdminListChatbots` with a type-safe SQL builder (Squirrel) to improve maintainability and safety.

## Scope
- `internal/db/admin_chatbots.go`

## Checklist
- [x] Add `github.com/Masterminds/squirrel` dependency.
- [x] Create a reproduction test for `AdminListChatbots` (if not exists) to ensure baseline behavior.
- [x] Refactor `AdminListChatbots` to use `squirrel.Select(...)`.
    - [x] Handle dynamic `Where` clauses for filters.
    - [x] Handle Pagination (with validation).
    - [x] Handle Sorting.
- [x] Verify the generated SQL is correct and comparable to original.
- [x] Run existing tests for Admin Chatbots.

## Implementation Notes

### Changes Made
1. Added `github.com/Masterminds/squirrel v1.5.4` to go.mod
2. Refactored `AdminListChatbots` function to use Squirrel query builder
3. Added pagination validation (limit > 0 defaults to 20, offset < 0 defaults to 0)
4. Used `sq.Dollar` placeholder format for PostgreSQL ($1, $2...)
5. Preserved ILIKE for case-insensitive name filtering
6. Maintained all subqueries, joins, and sorting logic
7. Created comprehensive unit tests with 13 test cases covering all edge cases

### Test Results - All 13 Tests Pass
- ✅ EmptyFilters: Returns all non-deleted chatbots
- ✅ NameFilterCaseInsensitive: Case-insensitive name matching works
- ✅ OrganizationFilter: Filters by organization_id correctly
- ✅ OwnerFilter: Filters by user_id correctly
- ✅ AllFilters: Combined filters work correctly
- ✅ Pagination: LIMIT and OFFSET work correctly
- ✅ DefaultPagination: Default values (20, 0) applied correctly
- ✅ SubqueryResults: SourceCount and MessageCount calculated correctly
- ✅ JoinsCorrect: Organization and user joins populate fields correctly
- ✅ NullFields: Optional NULL fields handled correctly
- ✅ SQLInjectionPrevention: SQL injection attempts prevented
- ✅ SQLInjectionSpecialCharacters: Special characters in names handled safely
- ✅ EmptyResult: Returns empty slice for no matches
- ✅ DeletedExcluded: Soft-deleted chatbots excluded from results
- ✅ NegativeLimit: Defaults to 20
- ✅ ZeroLimit: Defaults to 20
- ✅ NegativeOffset: Defaults to 0
- ✅ SortByCreatedAtDesc: Results sorted by created_at DESC

### Existing Tests
- ✅ Repository tests: All pass
- ✅ Handler tests: All pass

### Code Quality
- ✅ `make fmt`: Code formatted
- ✅ `make vet`: No vet issues
- ✅ `make lint`: No linting issues

### SQL Comparison
- Original: Raw SQL string with manual concatenation
- New: Squirrel-built queries with same semantics
- Both use: `ILIKE` for case-insensitive search
- Both use: `::uuid` and `::text` type casts
- Both use: `$1, $2...` PostgreSQL placeholders
- Both generate: Identical query structure and semantics

## Edge Cases - All Handled
- ✅ Empty filters: Returns all chatbots
- ✅ SQL injection prevention: Builder safely parameterizes all inputs
- ✅ Complex joins: Supported via Squirrel (LEFT JOIN with organizations, JOIN with users)
- ✅ Subqueries: Supported via Squirrel (COUNT subqueries in SELECT)
- ✅ Pagination validation: Negative/zero values defaulted to safe values

## Files Changed
- `go.mod`: Added `github.com/Masterminds/squirrel v1.5.4`
- `internal/db/admin_chatbots.go`: Refactored `AdminListChatbots` function (replaced manual SQL with Squirrel)
- `internal/db/admin_chatbots_test.go`: Created comprehensive unit tests (13 test cases, all passing)
- `docs/tech-debt/tasks/007-refactor-complex-sql-queries.md`: Updated checklist and added implementation notes
