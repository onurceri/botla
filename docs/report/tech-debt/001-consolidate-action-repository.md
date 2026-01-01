# Tech Debt Task 001: Consolidate Action Repository

> **Agent Prompt**: You are refactoring the Action data access layer to eliminate redundant abstraction. The `PostgresActionRepo` currently delegates all calls to static functions in `internal/db/action.go`. Your task is to move the SQL logic directly into the repository implementation using **Squirrel SQL builder** (`github.com/Masterminds/squirrel v1.5.4`), then deprecate/remove the db package functions. Apply strict TDD: ensure existing tests pass throughout, and add tests for edge cases. Do not break any existing functionality.

## Background

The codebase has a redundant data access abstraction where:
- `internal/repository/action_repo.go` contains `PostgresActionRepo` which implements `ActionRepository`
- Each method simply delegates to a function in `internal/db/action.go`
- This creates double maintenance burden and cognitive overhead

**Example of current redundancy:**
```go
// internal/repository/action_repo.go
func (r *PostgresActionRepo) List(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
    return db.GetActions(ctx, r.pool, chatbotID)
}

// internal/db/action.go
func GetActions(ctx context.Context, db *sql.DB, chatbotID string) ([]*models.ChatbotAction, error) {
    query := `SELECT id, chatbot_id, ... FROM chatbot_actions WHERE chatbot_id = $1`
    // actual SQL logic here
}
```

## Goal

Move all SQL logic from `internal/db/action.go` and `internal/db/action_logs.go` into `internal/repository/action_repo.go` using **Squirrel** for query building, eliminating the middleman.

## Squirrel Usage

Use `github.com/Masterminds/squirrel` for building SQL queries. Example pattern:

```go
import sq "github.com/Masterminds/squirrel"

// Create PostgreSQL-compatible builder (uses $1, $2 placeholders)
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *PostgresActionRepo) List(ctx context.Context, chatbotID string) ([]*models.ChatbotAction, error) {
    query, args, err := psql.
        Select("id", "chatbot_id", "name", "description", "action_type", "config", "parameters", "tool_name", "enabled", "version", "created_at", "updated_at").
        From("chatbot_actions").
        Where(sq.Eq{"chatbot_id": chatbotID}).
        OrderBy("created_at DESC").
        ToSql()
    if err != nil {
        return nil, err
    }
    
    rows, err := r.pool.QueryContext(ctx, query, args...)
    // ... scanning logic
}
```

## Files to Modify

| File | Action |
|------|--------|
| `internal/repository/action_repo.go` | Add SQL logic directly |
| `internal/repository/interfaces.go` | Move `ErrVersionConflict` here (if not already) |
| `internal/db/action.go` | Deprecate or remove |
| `internal/db/action_logs.go` | Deprecate or remove |

## Integration Plan

### Step 1: Verify Existing Tests Pass
```bash
go test ./internal/repository/... -v
go test ./internal/db/action*_test.go -v
```

### Step 2: Move SQL Logic for Each Method

For each method in `PostgresActionRepo`:

1. **List()** - Move `db.GetActions` logic
2. **ListEnabled()** - Move `db.GetEnabledActions` logic
3. **GetByID()** - Move `db.GetActionByID` logic
4. **GetByToolName()** - Move `db.GetActionByToolName` logic
5. **Create()** - Move `db.CreateAction` logic
6. **Update()** - Move `db.UpdateAction` logic (handle `ErrVersionConflict`)
7. **Delete()** - Move `db.DeleteAction` logic
8. **GetLogs()** - Move `db.GetActionLogs` logic
9. **CreateLog()** - Move `db.CreateActionLog` logic

### Step 3: Handle Error Types
- Move `ErrVersionConflict` to `internal/repository/interfaces.go` (if not there)
- Update any code that imports `db.ErrVersionConflict` to use the repository version

### Step 4: Update Imports
- Search for any files importing `internal/db` that use action-related functions
- Update them to use the repository interface instead

### Step 5: Deprecate db Functions
- Add deprecation comments to `internal/db/action.go` functions
- Or remove the file if no longer needed

## Edge Cases

- **Version conflict handling**: Ensure `Update()` correctly returns `ErrVersionConflict` when optimistic locking fails
- **Nil returns**: `GetByID` and `GetByToolName` should return `nil, nil` for not-found cases
- **Empty result sets**: `List()` and `ListEnabled()` should return empty slices, not nil

## Checklist

- [ ] Run existing tests to establish baseline
- [ ] Move `GetActions` SQL to `PostgresActionRepo.List`
- [ ] Move `GetEnabledActions` SQL to `PostgresActionRepo.ListEnabled`
- [ ] Move `GetActionByID` SQL to `PostgresActionRepo.GetByID`
- [ ] Move `GetActionByToolName` SQL to `PostgresActionRepo.GetByToolName`
- [ ] Move `CreateAction` SQL to `PostgresActionRepo.Create`
- [ ] Move `UpdateAction` SQL to `PostgresActionRepo.Update`
- [ ] Move `DeleteAction` SQL to `PostgresActionRepo.Delete`
- [ ] Move `GetActionLogs` SQL to `PostgresActionRepo.GetLogs`
- [ ] Move `CreateActionLog` SQL to `PostgresActionRepo.CreateLog`
- [ ] Consolidate `ErrVersionConflict` in repository package
- [ ] Update any external references to db action functions
- [ ] Remove import of `internal/db` from `action_repo.go`
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify all tests pass
- [ ] Run `make lint` to check for issues

## Verification

```bash
# Verify no compilation errors
go build ./...

# Run all tests
make test-all

# Verify no direct db.GetActions calls remain (outside db package tests)
grep -r "db\.GetActions\|db\.GetEnabledActions\|db\.CreateAction" internal/ --include="*.go" | grep -v "_test.go" | grep -v "internal/db/"
```
