# Task 002: Implement ChatbotRepository

## Background

Implement the `ChatbotRepository` interface by wrapping existing `db.*` functions. This provides a gradual migration path - the implementation delegates to existing code initially.

**Depends on:** Task 001

## Implementation Plan

1. **Create PostgresChatbotRepo**
   - File: `internal/repository/chatbot_repo.go`
   - Struct holds `*sql.DB`
   - Constructor: `NewPostgresChatbotRepo(db *sql.DB)`

2. **Implement methods by delegating to db package**
   ```go
   func (r *PostgresChatbotRepo) GetByID(ctx context.Context, id string) (*models.Chatbot, error) {
       return db.GetChatbotByID(ctx, r.db, id)
   }
   ```

3. **Create mock implementation for tests**
   - File: `internal/repository/mock_chatbot_repo.go`
   - Use function fields for easy test setup

## Files to Create

| File | Purpose |
|------|---------|
| `internal/repository/chatbot_repo.go` | PostgreSQL implementation |
| `internal/repository/mock_chatbot_repo.go` | Mock for unit tests |

## Checklist

- [x] Create `chatbot_repo.go` with `PostgresChatbotRepo` struct
- [x] Implement all `ChatbotRepository` interface methods
- [x] Create `mock_chatbot_repo.go` with mock implementation
- [x] Add basic unit test to verify mock works
- [x] Run `go build ./...` to verify compilation
