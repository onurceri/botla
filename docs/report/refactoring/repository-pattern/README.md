# Repository Pattern Refactoring

This folder contains incremental tasks to decouple HTTP handlers from direct database access by implementing the Repository Pattern.

## Problem Statement

HTTP Handlers in `internal/api/handlers` directly inject `*sql.DB` and execute data access logic. This makes handlers untestable without a real database and couples business logic to database implementation.

**Evidence:**
- `internal/api/handlers/action.go` - Direct `*sql.DB` dependency
- `internal/api/handlers/admin_chatbots.go` - Complex `db.*` orchestration
- `internal/api/handlers/chatbot_context.go` - Direct `db.*` calls

## Tasks

| # | Task | Status |
|---|------|--------|
| 001 | Define Repository Interfaces | `[x]` |
| 002 | Implement ChatbotRepository | `[x]` |
| 003 | Implement ActionRepository | `[x]` |
| 004 | Refactor ActionHandlers | `[x]` |
| 005 | Refactor AdminChatbotHandlers | `[x]` |

## Benefits

1. **Testability**: Unit test handlers with mock repositories
2. **Separation of Concerns**: Data access logic isolated from HTTP handling
3. **Flexibility**: Easy to swap database implementations
