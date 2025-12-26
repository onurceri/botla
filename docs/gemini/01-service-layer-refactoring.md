# Task 01: Service Layer Refactoring - Extract Orchestration Logic from HTTP Handlers

## Priority
**High** - Core architectural improvement affecting testability and extensibility

## Problem Statement

The HTTP handlers in `internal/api/handlers/` are performing high-level business orchestration, including:
- Subscription/plan validation
- Database persistence
- External API coordination
- Token quota enforcement

This "Fat Handler" anti-pattern makes the core logic impossible to unit test without mocking the entire HTTP stack. If the business decides to add a CLI tool or a Slack integration, the logic currently in handlers would have to be duplicated or undergo painful extraction.

## Evidence

### Current State in `internal/api/handlers/chat.go`

```go
func (h *ChatHandlers) Chat(w http.ResponseWriter, r *http.Request) {
    // Handler directly manages:
    // 1. Plan retrieval from database
    plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
    
    // 2. Model validation against allowed models
    for _, m := range plan.Config.Chat.AllowedModels {
        if m == cbot.Model {
            allowed = true
            break
        }
    }
    
    // 3. Token quota checking
    if plan.Config.Chat.MaxMonthlyTokens > 0 {
        used, errUsage := db.GetMonthlyTokenUsage(r.Context(), h.DB, userID)
        if errUsage == nil && used >= plan.Config.Chat.MaxMonthlyTokens {
            // Return error
        }
    }
    
    // 4. Only then delegates to service
    result, err := h.ChatService.ProcessChat(ctx, chatReq, cbot, ragConfig)
}
```

## Refactoring Goals

1. **Single Responsibility**: Handlers should only:
   - Parse HTTP request
   - Call a single service method
   - Format HTTP response

2. **Testability**: Business logic should be testable without HTTP mocking

3. **Reusability**: Logic should be callable from CLI, Slack bots, or other interfaces

## Implementation Plan

### Phase 1: Extend ChatService

**File**: `internal/services/chat_service.go`

```go
// ChatService should encapsulate all chat business logic
type ChatService struct {
    db           *sql.DB
    queries      *db.Queries
    ragClient    rag.VectorClient
    llmClient    rag.LLMClient
    embedder     rag.Embedder
}

// ProcessChatWithValidation handles the complete chat flow including validation
func (s *ChatService) ProcessChatWithValidation(ctx context.Context, req ChatRequestWithUser) (*ChatResult, error) {
    // 1. Get and validate plan
    plan, err := s.validateUserPlan(ctx, req.UserID)
    if err != nil {
        return nil, err
    }
    
    // 2. Check token quota
    if err := s.checkTokenQuota(ctx, req.UserID, plan); err != nil {
        return nil, err
    }
    
    // 3. Validate and adjust model
    model := s.validateModel(req.Chatbot.Model, plan)
    
    // 4. Process chat
    return s.ProcessChat(ctx, req.ChatRequest, req.Chatbot, plan.Config.Chat.RAG)
}
```

### Phase 2: Simplify Chat Handler

**File**: `internal/api/handlers/chat.go`

```go
func (h *ChatHandlers) Chat(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    cbot, _, ok := getChatbotContext(w, r, h.DB, h.WorkspaceService, h.OrgService)
    if !ok {
        return
    }

    userID, _ := middleware.UserIDFromContext(r.Context())
    
    var req chatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    
    // Simple validation only
    req.Message = strings.TrimSpace(req.Message)
    if req.Message == "" || strings.TrimSpace(req.SessionID) == "" {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), chatTimeout())
    defer cancel()

    // Single service call for all business logic
    result, err := h.ChatService.ProcessChatWithValidation(ctx, services.ChatRequestWithUser{
        UserID:      userID,
        Chatbot:     cbot,
        ChatRequest: models.ChatRequest{Message: req.Message, SessionID: req.SessionID},
    })
    if err != nil {
        h.handleChatError(w, err, cbot.LanguageCode)
        return
    }

    api.WriteJSON(w, http.StatusOK, result.ToResponse())
}
```

### Phase 3: Define Domain Errors

**File**: `internal/services/errors.go`

```go
package services

import "errors"

var (
    ErrTokenQuotaExceeded = errors.New("monthly token quota exceeded")
    ErrModelNotAllowed    = errors.New("model not allowed for plan")
    ErrPlanNotFound       = errors.New("user plan not found")
    ErrInvalidRequest     = errors.New("invalid chat request")
)
```

## Affected Files

| File | Action | Description |
|------|--------|-------------|
| `internal/services/chat_service.go` | MODIFY | Add `ProcessChatWithValidation`, plan validation, quota checking |
| `internal/services/errors.go` | NEW | Define domain-specific errors |
| `internal/api/handlers/chat.go` | MODIFY | Simplify to request parsing and response formatting |
| `internal/api/handlers/sources.go` | MODIFY | Apply same pattern if orchestration logic exists |

## Testing Strategy

### Unit Tests for ChatService

```go
func TestChatService_ProcessChatWithValidation_TokenQuotaExceeded(t *testing.T) {
    // Setup mock dependencies
    mockDB := setupMockDB()
    mockDB.SetMonthlyUsage(100000) // Over quota
    
    svc := NewChatService(mockDB, ...)
    
    _, err := svc.ProcessChatWithValidation(ctx, req)
    
    assert.ErrorIs(t, err, ErrTokenQuotaExceeded)
}
```

### Integration Tests

Existing integration tests in `internal/integration/` should continue to pass.

## Acceptance Criteria

- [ ] ChatService contains all business orchestration logic
- [ ] Chat handler is less than 50 lines of code
- [ ] Handler has no direct database calls for business logic
- [ ] All existing tests pass
- [ ] New unit tests cover plan validation and quota checking
- [ ] Service layer is reusable from non-HTTP contexts

## Estimated Effort

**Size**: Medium (2-3 days)
- Phase 1: 1 day
- Phase 2: 0.5 day
- Phase 3: 0.5 day
- Testing: 1 day

## Dependencies

None - this is a refactoring task with no external dependencies.

## Rollback Strategy

This is a refactoring task. If issues arise, individual commits can be reverted as the behavior should remain identical.
