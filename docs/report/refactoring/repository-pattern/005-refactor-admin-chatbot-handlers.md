# Task 005: Refactor AdminChatbotHandlers

## Background

Refactor `AdminChatbotHandlers` to use repository interfaces. This handler orchestrates complex logic and is the second primary handler identified in the issue.

**Depends on:** Task 002, Task 003

**File:** `internal/api/handlers/admin_chatbots.go`

## Current State

```go
type AdminChatbotHandlers struct {
    DB           *sql.DB
    AdminService *services.AdminService
    RagService   *services.RAGService
    Queue        *services.Queue
}
```

## Target State

```go
type AdminChatbotHandlers struct {
    ChatbotRepo  repository.ChatbotRepository
    SourceRepo   repository.SourceRepository
    AdminService *services.AdminService
    RagService   *services.RAGService
    Queue        *services.Queue
}
```

## Implementation Plan

1. **Add SourceRepository interface** (may need to extend Task 001)
2. **Update struct definition**
3. **Update handler methods**
   - `ForceRefreshChatbot` - uses multiple db.Admin* functions
   - `ListChatbots` - uses `db.AdminListChatbots`
   - `GetChatbot` - uses `db.AdminGetChatbot`
4. **Update wire-up**

## Special Considerations

The `ForceRefreshChatbot` method uses admin-specific db functions:
- `db.AdminGetChatbot`
- `db.AdminDeleteChatbotVectors`
- `db.AdminResetChatbotSources`
- `db.AdminGetChatbotSourceIDs`

These may need a separate `AdminChatbotRepository` interface or be added to `ChatbotRepository`.

## Checklist

- [x] Decide on interface structure (extend or separate admin interface)
- [x] Update `AdminChatbotHandlers` struct
- [x] Update `ListChatbots` method
- [x] Update `GetChatbot` method
- [x] Update `ForceRefreshChatbot` method
- [x] Update wire-up in `internal/api/router/router.go`
- [x] Update wire-up in `internal/integration/testserver.go`
- [x] Run existing tests to verify no regressions
- [x] Add comprehensive unit tests for handlers
- [x] Add comprehensive unit tests for repository mock

## Implementation Notes

Implemented as a **separate `AdminChatbotRepository` interface** because admin operations have elevated privileges and different return types (`AdminChatbot` vs `models.Chatbot`). This follows the Interface Segregation Principle.

### New Files Created

- `internal/repository/admin_chatbot_repo.go` - PostgreSQL implementation
- `internal/repository/mock_admin_chatbot_repo.go` - Mock implementation for testing
- `internal/repository/admin_chatbot_repo_test.go` - Unit tests for repository

### Modified Files

- `internal/repository/interfaces.go` - Added `AdminChatbotRepository`, `AdminChatbot`, `AdminChatbotFilter`
- `internal/api/handlers/admin_chatbots.go` - Updated to use repository interface
- `internal/api/router/router.go` - Updated wire-up
- `internal/integration/testserver.go` - Updated wire-up
- `internal/api/handlers/admin_chatbots_test.go` - Comprehensive handler tests

