# Tech Debt Task 009: Deprecate and Clean Up db Package

> **Agent Prompt**: You are performing the final cleanup of the repository consolidation effort. All services and handlers now use repository interfaces. Your task is to either (a) make `internal/db` functions package-private, or (b) remove unused functions entirely. Before removing anything, verify there are no remaining direct usages outside of tests. This is a cleanup task - do not change any behavior.

## Background

After completing Tasks 001-008:
- Repository implementations contain all SQL logic
- Services and handlers use repository interfaces
- The `internal/db` package now has redundant exported functions

## Goal

1. Audit remaining `internal/db` usages
2. Make functions private or remove them
3. Clean up unused types
4. Update or remove orphaned tests

## Integration Plan

### Step 1: Audit Remaining Usages

Run grep to find any remaining direct db calls:

```bash
# Find all db.* calls outside of internal/db/
grep -r "db\.\(Get\|Create\|Update\|Delete\|Count\|List\)" internal/ \
    --include="*.go" \
    | grep -v "internal/db/" \
    | grep -v "_test.go"
```

If any remain, they must be migrated first.

### Step 2: Categorize db Functions

For each function in `internal/db/*.go`:

| Category | Action |
|----------|--------|
| Used only by repository | Safe to make private or remove |
| Used by tests only | Consider keeping for test utilities |
| Still used directly | Must migrate first (go back to Task 007/008) |

### Step 3: Make Functions Private

For functions that should remain (e.g., test utilities):

Before:
```go
func GetChatbotByID(ctx context.Context, pool *sql.DB, id string) (*models.Chatbot, error)
```

After:
```go
func getChatbotByID(ctx context.Context, pool *sql.DB, id string) (*models.Chatbot, error)
```

### Step 4: Remove Redundant Functions

For functions now fully implemented in repositories:

```bash
# Example: Remove db.GetActions since PostgresActionRepo.List has the SQL
# Delete the function from internal/db/action.go
```

### Step 5: Clean Up Types

Some types may be duplicated between db and repository packages:
- `db.AdminChatbot` vs `repository.AdminChatbot`
- `db.ChatbotFilter` vs `repository.AdminChatbotFilter`

Remove the db package types if no longer used.

### Step 6: Update or Remove Tests

`internal/db/*_test.go` files may test functions that no longer exist:
- Move relevant tests to `internal/repository/*_test.go`
- Delete tests for removed functions

### Step 7: Consider Package Removal

If all functions are removed, consider:
- Deleting the entire `internal/db/` package
- OR keeping minimal utilities (e.g., connection pooling, migrations)

## Files Likely to Modify/Delete

| File | Likely Action |
|------|---------------|
| `internal/db/action.go` | DELETE (moved to repository) |
| `internal/db/action_logs.go` | DELETE (moved to repository) |
| `internal/db/chatbot.go` | DELETE or reduce to helpers |
| `internal/db/admin_chatbots.go` | DELETE (moved to repository) |
| `internal/db/source.go` | DELETE (moved to repository) |
| `internal/db/conversation.go` | DELETE (moved to repository) |
| `internal/db/plan.go` | DELETE (moved to repository) |
| `internal/db/db.go` | KEEP (connection setup) |
| `internal/db/*_test.go` | DELETE or move |

## Edge Cases

- **Integration test helpers**: `testdb` package may depend on db functions
- **Migration tooling**: May use db package directly
- **Circular imports**: Ensure repository doesn't accidentally need db

## Checklist

### Phase 1: Audit
- [ ] Run grep to find remaining direct db usages
- [ ] Document any functions still in use
- [ ] Go back to Task 007/008 if migrations incomplete

### Phase 2: Action Package
- [ ] Verify no direct `db.GetActions` etc. usages
- [ ] Delete `internal/db/action.go`
- [ ] Delete `internal/db/action_logs.go`
- [ ] Run `go build ./...`

### Phase 3: Chatbot Package
- [ ] Verify no direct chatbot db usages
- [ ] Delete or reduce `internal/db/chatbot.go`
- [ ] Delete `internal/db/admin_chatbots.go`
- [ ] Run `go build ./...`

### Phase 4: Source Package
- [ ] Verify no direct source db usages
- [ ] Delete `internal/db/source.go`
- [ ] Run `go build ./...`

### Phase 5: Other Files
- [ ] Review and clean up remaining db files
- [ ] Keep `db.go` for connection utilities
- [ ] Run `go build ./...`

### Phase 6: Test Cleanup
- [ ] Move useful tests to repository package
- [ ] Delete obsolete test files
- [ ] Run `make test-all`

### Verification
- [ ] Run `go build ./...`
- [ ] Run `make test-all`
- [ ] Run `make lint`
- [ ] Verify no import cycles

## Verification

```bash
# Ensure no compilation errors
go build ./...

# Ensure all tests pass
make test-all

# Check for unused imports
go mod tidy

# Verify no direct db calls remain
grep -r "internal/db\"" internal/ --include="*.go" \
    | grep -v "internal/db/" \
    | grep -v "_test.go" \
    | grep -v "testdb"
```

## Success Criteria

- [ ] `internal/db/*.go` contains only connection/utility code (or is deleted)
- [ ] No direct `db.*` calls in services or handlers
- [ ] All repository implementations contain SQL logic
- [ ] All tests pass
- [ ] No import cycles
