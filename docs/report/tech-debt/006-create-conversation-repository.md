# Tech Debt Task 006: Create Conversation Repository

> **Agent Prompt**: You are creating a new `ConversationRepository` interface and implementation to encapsulate conversation and message database operations. Currently, the chat pipeline calls `db.GetOrCreateConversationBySessionID` and `db.CreateMessage` directly. Your task is to define the interface, implement `PostgresConversationRepo` using **Squirrel SQL builder** (`github.com/Masterminds/squirrel v1.5.4`), and prepare for consumer migration. Apply strict TDD.

## Background

Conversation-related functions are called from the chat pipeline:
- `internal/services/chat_pipeline.go`: 
  - `db.GetOrCreateConversationBySessionID`
  - `db.CreateMessage`

The `internal/db/conversation.go` file contains the implementation.

## Goal

1. Define `ConversationRepository` interface in `internal/repository/interfaces.go`
2. Create `internal/repository/conversation_repo.go` with `PostgresConversationRepo`
3. Create mock implementation for testing
4. Do NOT migrate consumers yet

## Files to Create/Modify

| File | Action |
|------|--------|
| `internal/repository/interfaces.go` | Add `ConversationRepository` interface |
| `internal/repository/conversation_repo.go` | **[NEW]** Implement |
| `internal/repository/conversation_repo_test.go` | **[NEW]** Integration tests |
| `internal/repository/mock_conversation_repo.go` | **[NEW]** Mock implementation |
| `internal/db/conversation.go` | Reference for SQL logic |

## Integration Plan

### Step 1: Analyze Existing db Functions

Review `internal/db/conversation.go` to understand:
- `GetOrCreateConversationBySessionID` - upsert pattern
- `CreateMessage` - message storage
- Any other conversation-related functions

### Step 2: Define Interface

```go
// internal/repository/interfaces.go

// ConversationRepository defines the interface for conversation data access operations.
// Conversations contain chat sessions between users and chatbots.
type ConversationRepository interface {
    // GetOrCreateBySessionID finds an existing conversation or creates a new one.
    // Uses session_id as the unique identifier within a chatbot.
    GetOrCreateBySessionID(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error)
    
    // GetByID retrieves a conversation by its unique identifier.
    GetByID(ctx context.Context, id string) (*models.Conversation, error)
    
    // CreateMessage persists a new message in a conversation.
    // Returns the generated message ID.
    CreateMessage(ctx context.Context, msg *models.Message) (string, error)
    
    // GetMessages retrieves messages for a conversation with pagination.
    GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]models.Message, error)
}
```

### Step 3: Handle Upsert Pattern

`GetOrCreateBySessionID` uses PostgreSQL's `ON CONFLICT` clause:

```go
func (r *PostgresConversationRepo) GetOrCreateBySessionID(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error) {
    // PostgreSQL UPSERT with Squirrel
    insertQuery, insertArgs, err := psql.
        Insert("conversations").
        Columns("chatbot_id", "session_id").
        Values(chatbotID, sessionID).
        Suffix("ON CONFLICT (chatbot_id, session_id) DO UPDATE SET updated_at = NOW()").
        Suffix("RETURNING id, chatbot_id, session_id, created_at, updated_at").
        ToSql()
    // ...
}
```

### Step 4: Create Mock

```go
type MockConversationRepo struct {
    GetOrCreateBySessionIDFunc func(ctx context.Context, chatbotID, sessionID string) (*models.Conversation, error)
    CreateMessageFunc          func(ctx context.Context, msg *models.Message) (string, error)
    // ...
}
```

## Edge Cases

- **Race conditions**: Multiple requests creating same session
- **Message ordering**: Messages should be ordered by created_at
- **Large conversations**: Pagination must work correctly

## Checklist

- [ ] Add `ConversationRepository` interface to `interfaces.go`
- [ ] Create `conversation_repo.go` with struct skeleton
- [ ] Write test for `GetOrCreateBySessionID`
- [ ] Implement `GetOrCreateBySessionID` using Squirrel
- [ ] Write test for `GetByID`
- [ ] Implement `GetByID` using Squirrel
- [ ] Write test for `CreateMessage`
- [ ] Implement `CreateMessage` using Squirrel
- [ ] Write test for `GetMessages`
- [ ] Implement `GetMessages` using Squirrel
- [ ] Create `mock_conversation_repo.go`
- [ ] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify all tests pass

## Verification

```bash
go test ./internal/repository/conversation_repo_test.go -v
make test-all
```
