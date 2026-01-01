# Tech Debt Task 003: Consolidate AdminChatbot Repository

> **Agent Prompt**: You are refactoring the AdminChatbot data access layer to eliminate redundant abstraction. The `PostgresAdminChatbotRepo` currently delegates all calls to static functions in `internal/db/admin_chatbots.go`. Your task is to move the SQL logic directly into the repository implementation using **Squirrel SQL builder** (`github.com/Masterminds/squirrel v1.5.4`), then deprecate/remove the db package functions. Pay special attention to the type conversion between `db.AdminChatbot` and `repository.AdminChatbot`.

## Background

The `internal/db/admin_chatbots.go` file contains admin-specific chatbot queries with complex joins. These are wrapped by `PostgresAdminChatbotRepo`.

| Repository Method | DB Function |
|-------------------|-------------|
| `ListChatbots` | `db.AdminListChatbots` |
| `GetByID` | `db.AdminGetChatbot` |
| `ResetSources` | `db.AdminResetChatbotSources` |
| `GetSourceIDs` | `db.AdminGetChatbotSourceIDs` |
| `DeleteVectors` | `db.AdminDeleteChatbotVectors` |

## Goal

Move all SQL logic from `internal/db/admin_chatbots.go` into `internal/repository/admin_chatbot_repo.go` using Squirrel, eliminating the type conversion layer.

## Current Type Duplication

Currently there are two identical types:
- `db.AdminChatbot` 
- `repository.AdminChatbot`

And a conversion function `toRepoAdminChatbot()`. After consolidation, only `repository.AdminChatbot` should exist.

## Files to Modify

| File | Action |
|------|--------|
| `internal/repository/admin_chatbot_repo.go` | Add SQL logic using Squirrel |
| `internal/repository/interfaces.go` | Keep `AdminChatbot` type here |
| `internal/db/admin_chatbots.go` | Deprecate or remove |

## Integration Plan

### Step 1: Verify Existing Tests Pass
```bash
go test ./internal/repository/admin_chatbot_repo_test.go -v
go test ./internal/db/admin_chatbots_test.go -v
```

### Step 2: Move Filter Type
The `db.ChatbotFilter` type is used by `AdminListChatbots`. Either:
- Rename `repository.AdminChatbotFilter` to match or
- Consolidate into one type in the repository package

### Step 3: Move SQL Logic with Squirrel

Example for `ListChatbots` with optional filters:
```go
func (r *PostgresAdminChatbotRepo) ListChatbots(ctx context.Context, filter AdminChatbotFilter, limit, offset int) ([]AdminChatbot, int, error) {
    builder := psql.
        Select(
            "c.id", "c.name", "c.user_id", "c.workspace_id",
            "o.id AS organization_id", "o.name AS organization_name",
            "u.email AS owner_email",
            // ... subqueries for counts
        ).
        From("chatbots c").
        LeftJoin("users u ON c.user_id = u.id").
        LeftJoin("organizations o ON c.organization_id = o.id").
        Where(sq.Eq{"c.deleted_at": nil})

    // Apply optional filters
    if filter.Name != nil {
        builder = builder.Where(sq.ILike{"c.name": "%" + *filter.Name + "%"})
    }
    if filter.OrganizationID != nil {
        builder = builder.Where(sq.Eq{"c.organization_id": *filter.OrganizationID})
    }
    // ...
}
```

### Step 4: Handle Complex Queries
`AdminListChatbots` has:
- Multiple JOINs (users, organizations)
- Subqueries for source_count and message_count
- Optional filter conditions
- Pagination with total count

For subqueries in SELECT, you may need raw SQL fragments:
```go
sq.Expr("(SELECT COUNT(*) FROM sources WHERE chatbot_id = c.id) AS source_count")
```

## Edge Cases

- **Optional filter handling**: All filter fields are pointers, only apply WHERE when non-nil
- **NULL organization fields**: Handle nil organization_id, organization_name
- **Pagination edge cases**: offset=0, limit=0, empty result set
- **Total count**: Must return accurate total regardless of limit/offset

## Checklist

- [ ] Run existing tests to establish baseline
- [ ] Remove `db.AdminChatbot` type dependency
- [ ] Remove `db.ChatbotFilter` type dependency
- [ ] Remove `toRepoAdminChatbot` conversion function
- [ ] Move `AdminListChatbots` SQL using Squirrel
- [ ] Move `AdminGetChatbot` SQL using Squirrel
- [ ] Move `AdminResetChatbotSources` SQL using Squirrel
- [ ] Move `AdminGetChatbotSourceIDs` SQL using Squirrel
- [ ] Move `AdminDeleteChatbotVectors` SQL using Squirrel
- [ ] Remove import of `internal/db` from `admin_chatbot_repo.go`
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify all tests pass
- [ ] Run `make lint` to check for issues

## Verification

```bash
# Verify no compilation errors
go build ./...

# Run all tests
make test-all

# Verify no direct db.Admin* calls remain
grep -r "db\.Admin" internal/ --include="*.go" | grep -v "_test.go" | grep -v "internal/db/"
```
