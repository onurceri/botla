# Tech Debt Task 002: Consolidate Chatbot Repository

> **Agent Prompt**: You are refactoring the Chatbot data access layer to eliminate redundant abstraction. The `PostgresChatbotRepo` currently delegates all calls to static functions in `internal/db/chatbot.go`. Your task is to move the SQL logic directly into the repository implementation using **Squirrel SQL builder** (`github.com/Masterminds/squirrel v1.5.4`), then deprecate/remove the db package functions. Apply strict TDD: ensure existing tests pass throughout. This is a larger file (~527 lines) so proceed methodically.

## Background

The `internal/db/chatbot.go` file contains 527 lines of SQL logic that is wrapped by `PostgresChatbotRepo`. The repository methods are:

| Repository Method | DB Function |
|-------------------|-------------|
| `GetByID` | `db.GetChatbotByID` |
| `GetByUserID` | `db.GetChatbotsByUserID` |
| `GetByWorkspace` | `db.GetChatbotsByWorkspace` |
| `Create` | `db.CreateChatbot` |
| `Update` | `db.UpdateChatbot` |
| `SoftDelete` | `db.SoftDeleteChatbot` |
| `CountByUserID` | `db.CountChatbotsByUserID` |
| `CountByWorkspace` | `db.CountChatbotsByWorkspace` |
| `UpdateSuggestedQuestions` | `db.UpdateChatbotSuggestedQuestions` |

## Goal

Move all SQL logic from `internal/db/chatbot.go` into `internal/repository/chatbot_repo.go` using Squirrel.

## Squirrel Usage

```go
import sq "github.com/Masterminds/squirrel"

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *PostgresChatbotRepo) GetByID(ctx context.Context, id string) (*models.Chatbot, error) {
    query, args, err := psql.
        Select("id", "user_id", "name", /* ... all fields */).
        From("chatbots").
        Where(sq.Eq{"id": id}).
        Where(sq.Eq{"deleted_at": nil}).
        ToSql()
    // ...
}
```

## Files to Modify

| File | Action |
|------|--------|
| `internal/repository/chatbot_repo.go` | Add SQL logic using Squirrel |
| `internal/db/chatbot.go` | Deprecate or remove |

## Integration Plan

### Step 1: Verify Existing Tests Pass
```bash
go test ./internal/repository/chatbot_repo_test.go -v
go test ./internal/db/chatbot_test.go -v
```

### Step 2: Move Helper Functions First
The db package has helper functions that should be moved:
- `scanChatbots(rows *sql.Rows)` - row scanning helper
- `normalizeLocale(code string)` - locale normalization

Consider making these private methods on `PostgresChatbotRepo` or package-level functions in repository.

### Step 3: Move Each Method's SQL Logic
Migrate in this order (simpler to complex):
1. `CountByUserID` - simple count query
2. `CountByWorkspace` - simple count query
3. `GetByUserID` - list with scanning
4. `GetByWorkspace` - list with scanning
5. `GetByID` - single row with complex scanning
6. `Create` - insert with RETURNING
7. `Update` - update with many fields
8. `UpdateSuggestedQuestions` - simple update
9. `SoftDelete` - transaction with source cleanup

### Step 4: Handle Transactions
`SoftDelete` uses a transaction to:
1. Get source IDs before deletion
2. Set `deleted_at` on chatbot
3. Return source IDs for vector cleanup

Ensure transaction handling is preserved:
```go
tx, err := r.pool.BeginTx(ctx, nil)
// ... operations
tx.Commit()
```

## Edge Cases

- **Nil handling for optional fields**: Workspace ID, Organization ID can be nil
- **JSON arrays**: `suggested_questions`, `initial_messages` are JSON arrays
- **Locale normalization**: Preserve the `normalizeLocale` logic
- **Transaction rollback**: Ensure proper rollback on `SoftDelete` failure

## Checklist

- [ ] Run existing tests to establish baseline
- [ ] Move `scanChatbots` helper to repository package
- [ ] Move `normalizeLocale` helper to repository package
- [ ] Move `CountByUserID` SQL using Squirrel
- [ ] Move `CountByWorkspace` SQL using Squirrel
- [ ] Move `GetByUserID` SQL using Squirrel
- [ ] Move `GetByWorkspace` SQL using Squirrel
- [ ] Move `GetByID` SQL using Squirrel
- [ ] Move `Create` SQL using Squirrel
- [ ] Move `Update` SQL using Squirrel
- [ ] Move `UpdateSuggestedQuestions` SQL using Squirrel
- [ ] Move `SoftDelete` SQL with transaction handling
- [ ] Remove import of `internal/db` from `chatbot_repo.go`
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify all tests pass
- [ ] Run `make lint` to check for issues

## Verification

```bash
# Verify no compilation errors
go build ./...

# Run all tests
make test-all

# Verify no direct db.GetChatbot* calls remain
grep -r "db\.GetChatbot\|db\.CreateChatbot\|db\.UpdateChatbot" internal/ --include="*.go" | grep -v "_test.go" | grep -v "internal/db/"
```
