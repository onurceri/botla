# Tech Debt Task 004: Implement Source Repository

> **Agent Prompt**: You are implementing the `SourceRepository` interface which already exists in `internal/repository/interfaces.go`. Create `PostgresSourceRepo` in a new file `internal/repository/source_repo.go` by moving SQL logic from `internal/db/source.go` using **Squirrel SQL builder** (`github.com/Masterminds/squirrel v1.5.4`). This is a new implementation, not a wrapper refactor. Apply strict TDD: write tests first for each method.

## Background

The `SourceRepository` interface is already defined in `internal/repository/interfaces.go`:

```go
type SourceRepository interface {
    GetByID(ctx context.Context, id string) (*models.DataSource, error)
    GetByChatbot(ctx context.Context, chatbotID string) ([]models.DataSource, error)
    GetURLSources(ctx context.Context, chatbotID string) ([]models.DataSource, error)
    Create(ctx context.Context, source *models.DataSource) (string, error)
    SoftDelete(ctx context.Context, id string) error
    Delete(ctx context.Context, id string) error
    Exists(ctx context.Context, chatbotID, url string) (bool, error)
    ExistsByHash(ctx context.Context, chatbotID, hash string) (bool, error)
    GetByHash(ctx context.Context, chatbotID, hash string) (*models.DataSource, error)
    CountByType(ctx context.Context, chatbotID, sourceType string) (int, error)
}
```

But there's no implementation yet! The `internal/db/source.go` file (14440 bytes) contains the actual SQL.

## Goal

1. Create `internal/repository/source_repo.go` with `PostgresSourceRepo` struct
2. Implement all `SourceRepository` interface methods using Squirrel
3. Move logic from `internal/db/source.go`
4. Create mock implementation for testing

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/repository/source_repo.go` | **[NEW]** Implement `PostgresSourceRepo` |
| `internal/repository/source_repo_test.go` | **[NEW]** Integration tests |
| `internal/repository/mock_source_repo.go` | **[NEW]** Mock implementation |
| `internal/db/source.go` | Move functions, then deprecate |

## Integration Plan

### Step 1: Create Empty Implementation
```go
// internal/repository/source_repo.go
package repository

import (
    "context"
    "database/sql"
    
    sq "github.com/Masterminds/squirrel"
    "github.com/onurceri/botla-co/internal/models"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type PostgresSourceRepo struct {
    pool *sql.DB
}

var _ SourceRepository = (*PostgresSourceRepo)(nil)

func NewPostgresSourceRepo(pool *sql.DB) *PostgresSourceRepo {
    return &PostgresSourceRepo{pool: pool}
}
```

### Step 2: TDD - Write Tests First
For each method, write a test before implementing:

```go
func TestPostgresSourceRepo_GetByID(t *testing.T) {
    db := testdb.OpenParallelTestDB(t)
    repo := NewPostgresSourceRepo(db)
    
    // Create test data
    user := testdb.CreateUser(t, db, testdb.UserFixture{})
    chatbot := testdb.CreateChatbot(t, db, testdb.ChatbotFixture{UserID: user.ID})
    source := testdb.CreateSource(t, db, testdb.SourceFixture{ChatbotID: chatbot.ID})
    
    // Test GetByID
    result, err := repo.GetByID(context.Background(), source.ID)
    require.NoError(t, err)
    assert.Equal(t, source.ID, result.ID)
}
```

### Step 3: Implement Methods with Squirrel

Example implementation:
```go
func (r *PostgresSourceRepo) GetByID(ctx context.Context, id string) (*models.DataSource, error) {
    query, args, err := psql.
        Select("id", "chatbot_id", "url", "type", "status", "hash", "chunk_count", "created_at", "updated_at", "deleted_at").
        From("sources").
        Where(sq.Eq{"id": id}).
        ToSql()
    if err != nil {
        return nil, err
    }
    
    var s models.DataSource
    err = r.pool.QueryRowContext(ctx, query, args...).Scan(
        &s.ID, &s.ChatbotID, &s.URL, &s.Type, &s.Status,
        &s.Hash, &s.ChunkCount, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
    )
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &s, nil
}
```

### Step 4: Create Mock Implementation
```go
// internal/repository/mock_source_repo.go
type MockSourceRepo struct {
    GetByIDFunc      func(ctx context.Context, id string) (*models.DataSource, error)
    GetByChatbotFunc func(ctx context.Context, chatbotID string) ([]models.DataSource, error)
    // ... other function fields
}

func (m *MockSourceRepo) GetByID(ctx context.Context, id string) (*models.DataSource, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return nil, nil
}
```

## Edge Cases

- **Soft delete vs hard delete**: `SoftDelete` sets `deleted_at`, `Delete` removes row
- **Exists check**: Must check for non-deleted sources only
- **Hash collisions**: `GetByHash` finds sources by content hash
- **URL normalization**: Consider how URLs are stored/compared

## Checklist

- [ ] Create `internal/repository/source_repo.go` with struct skeleton
- [ ] Add compile-time interface check
- [ ] Write test for `GetByID`
- [ ] Implement `GetByID` using Squirrel
- [ ] Write test for `GetByChatbot`
- [ ] Implement `GetByChatbot` using Squirrel
- [ ] Write test for `GetURLSources`
- [ ] Implement `GetURLSources` using Squirrel
- [ ] Write test for `Create`
- [ ] Implement `Create` using Squirrel
- [ ] Write test for `SoftDelete`
- [ ] Implement `SoftDelete` using Squirrel
- [ ] Write test for `Delete`
- [ ] Implement `Delete` using Squirrel
- [ ] Write test for `Exists`
- [ ] Implement `Exists` using Squirrel
- [ ] Write test for `ExistsByHash`
- [ ] Implement `ExistsByHash` using Squirrel
- [ ] Write test for `GetByHash`
- [ ] Implement `GetByHash` using Squirrel
- [ ] Write test for `CountByType`
- [ ] Implement `CountByType` using Squirrel
- [ ] Create `mock_source_repo.go`
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify all tests pass
- [ ] Run `make lint` to check for issues

## Verification

```bash
# Run new repository tests
go test ./internal/repository/source_repo_test.go -v

# Verify all tests pass
make test-all
```
