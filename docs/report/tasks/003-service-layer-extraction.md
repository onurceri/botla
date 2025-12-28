# Task 003: Service Layer Sub-domain Extraction

## Agent Prompt

> **Objective:** Extract sub-domain services from `ChatService` to prevent it from becoming a monolith that touches analytics, limits, RAG, policies, and guardrails.
>
> **Context:** The architecture review identified that `ChatService` is accumulating cross-domain logic. While not critical yet, proactive extraction will improve maintainability and testability.
>
> **Approach:**
> 1. Identify distinct responsibilities within ChatService
> 2. Extract focused services (e.g., `ChatContextBuilder`, `LimitEnforcer`)
> 3. Use dependency injection to compose them
> 4. Maintain the same public API for ChatService

---

## Problem Statement

`ChatService` currently handles:
- Chat context initialization
- RAG search orchestration
- Message building
- Agentic loop execution
- Fallback logic
- Analytics tracking
- Plan/quota enforcement
- Guardrails enforcement

This violates the Single Responsibility Principle and makes the service harder to test in isolation.

## Impact

- **Medium Risk**: Refactoring with interface changes
- **Improved Testability**: Each sub-service can be unit tested independently
- **Better Maintainability**: Changes to analytics don't risk breaking chat logic

---

## Acceptance Criteria

- [ ] `ChatContextBuilder` extracted and handles context initialization
- [ ] `QuotaEnforcer` extracted and handles token reservation/adjustment
- [ ] `ChatAnalyticsTracker` extracted (may already exist as `AnalyticsService`)
- [ ] `ChatService` composes these sub-services via dependency injection
- [ ] `ChatService` public API remains unchanged
- [ ] All existing tests pass
- [ ] New unit tests for each extracted service

---

## Current Architecture Analysis

### ChatService Dependencies (from chat_service.go)

```go
type ChatService struct {
    DB            *sql.DB
    Factory       *rag.ClientFactory
    Embedder      rag.EmbeddingClient
    QC            rag.VectorClient
    Log           *logger.Logger
    Guardrails    *GuardrailService    // Already extracted!
    SyncAnalytics bool
}
```

### Identified Responsibilities

| File | Responsibility | Candidate Service |
|---|---|---|
| `chat_context.go` | Context initialization | `ChatContextBuilder` |
| `chat_pipeline.go` | RAG + message building | Keep in `ChatService` |
| `chat_fallback.go` | Fallback logic | Keep in `ChatService` (uses `GuardrailService`) |
| `chat_helpers.go` | Client init, analytics | `ChatAnalyticsTracker` |
| `chat_prompts.go` | Prompt templates | Keep as helper functions |
| `ProcessChatWithValidation` | Quota enforcement | `QuotaEnforcer` |

---

## Implementation Plan

### Phase 1: Extract QuotaEnforcer

The quota enforcement logic in `ProcessChatWithValidation` can be extracted.

- [ ] **Step 1.1**: Create `internal/services/quota_enforcer.go`
  ```go
  package services
  
  import (
      "context"
      "database/sql"
      "errors"
      "fmt"
  
      "github.com/onurceri/botla-co/internal/db"
      "github.com/onurceri/botla-co/internal/models"
  )
  
  // QuotaEnforcer handles token quota reservation and adjustment.
  type QuotaEnforcer struct {
      DB *sql.DB
  }
  
  // NewQuotaEnforcer creates a new QuotaEnforcer.
  func NewQuotaEnforcer(db *sql.DB) *QuotaEnforcer {
      return &QuotaEnforcer{DB: db}
  }
  
  // ReserveTokens reserves estimated tokens for a chat request.
  // Returns the number of tokens reserved.
  func (q *QuotaEnforcer) ReserveTokens(ctx context.Context, userID string, estimatedTokens, maxMonthlyTokens int) error {
      if maxMonthlyTokens <= 0 {
          return nil // No quota enforcement
      }
      err := db.ReserveChatTokens(ctx, q.DB, userID, estimatedTokens, maxMonthlyTokens)
      if err != nil {
          if errors.Is(err, db.ErrTokenQuotaExceeded) {
              return ErrTokenQuotaExceeded
          }
          return fmt.Errorf("reserve tokens: %w", err)
      }
      return nil
  }
  
  // AdjustTokens adjusts token usage after chat completion.
  func (q *QuotaEnforcer) AdjustTokens(ctx context.Context, userID string, estimatedTokens, actualTokens int) {
      delta := actualTokens - estimatedTokens
      if delta != 0 {
          _ = db.AdjustChatTokens(ctx, q.DB, userID, delta)
      }
  }
  
  // RefundTokens refunds reserved tokens on error.
  func (q *QuotaEnforcer) RefundTokens(ctx context.Context, userID string, tokens int) {
      _ = db.AdjustChatTokens(ctx, q.DB, userID, -tokens)
  }
  ```

- [ ] **Step 1.2**: Create tests `internal/services/quota_enforcer_test.go`

- [ ] **Step 1.3**: Update `ChatService` to use `QuotaEnforcer`
  ```go
  type ChatService struct {
      // ... existing fields
      Quota *QuotaEnforcer
  }
  ```

### Phase 2: Extract ChatContextBuilder

- [ ] **Step 2.1**: Create `internal/services/chat_context_builder.go`
  ```go
  package services
  
  import (
      "context"
      
      "github.com/onurceri/botla-co/internal/models"
  )
  
  // ChatContextBuilder initializes chat context for processing.
  type ChatContextBuilder struct {
      // Configuration dependencies
  }
  
  // NewChatContextBuilder creates a new context builder.
  func NewChatContextBuilder() *ChatContextBuilder {
      return &ChatContextBuilder{}
  }
  
  // Build creates a new chatContext from request parameters.
  func (b *ChatContextBuilder) Build(
      ctx context.Context,
      req models.ChatRequest,
      bot *models.Chatbot,
      ragConfig models.RAGConfig,
      guardrailsCfg *models.GuardrailsConfig,
  ) *chatContext {
      // Move logic from ChatService.initChatContext here
  }
  ```

- [ ] **Step 2.2**: Move `initChatContext` logic to builder

- [ ] **Step 2.3**: Update `ChatService` to use builder

### Phase 3: Verify Analytics Integration

- [ ] **Step 3.1**: Review existing `AnalyticsService`
  - Location: `internal/services/analytics_service.go`
  - Check if it can be used for chat analytics

- [ ] **Step 3.2**: If needed, create `ChatAnalyticsTracker` wrapper
  ```go
  // ChatAnalyticsTracker wraps analytics for chat-specific tracking.
  type ChatAnalyticsTracker struct {
      svc *AnalyticsService
  }
  ```

### Phase 4: Update ChatService Composition

- [ ] **Step 4.1**: Update `NewChatService` to inject sub-services
  ```go
  func NewChatService(
      db *sql.DB,
      factory *rag.ClientFactory,
      embedder rag.EmbeddingClient,
      qc rag.VectorClient,
      log *logger.Logger,
  ) *ChatService {
      return &ChatService{
          DB:         db,
          Factory:    factory,
          Embedder:   embedder,
          QC:         qc,
          Log:        log,
          Guardrails: NewGuardrailService(log),
          Quota:      NewQuotaEnforcer(db),
          Context:    NewChatContextBuilder(),
      }
  }
  ```

- [ ] **Step 4.2**: Update `ProcessChatWithValidation` to use extracted services

- [ ] **Step 4.3**: Update `ProcessChat` to use extracted services

### Phase 5: Verification

- [ ] **Step 5.1**: Run all service tests
  ```bash
  go test ./internal/services/... -v
  ```

- [ ] **Step 5.2**: Run integration tests
  ```bash
  go test ./internal/integration/... -v -run Chat
  ```

- [ ] **Step 5.3**: Run full test suite
  ```bash
  make test-all
  ```

- [ ] **Step 5.4**: Verify no API changes
  - `ProcessChat` signature unchanged
  - `ProcessChatWithValidation` signature unchanged

---

## Files to Create

| File | Purpose |
|---|---|
| `internal/services/quota_enforcer.go` | Token quota management |
| `internal/services/quota_enforcer_test.go` | Unit tests |
| `internal/services/chat_context_builder.go` | Context initialization |
| `internal/services/chat_context_builder_test.go` | Unit tests |

## Files to Modify

| File | Changes |
|---|---|
| `internal/services/chat_service.go` | Add new dependencies, delegate to sub-services |
| `internal/services/chat_context.go` | May be refactored or deprecated |

---

## Design Decisions

### Why Not Extract RAG Search?

The RAG search is tightly coupled to the chat pipeline and doesn't have clear reuse outside of chat. Keep it in `ChatService` for now.

### Why Start with QuotaEnforcer?

It has the clearest boundaries:
- Input: userID, token counts
- Output: success/error
- No dependencies on other chat state

### Backward Compatibility

The public API (`ProcessChat`, `ProcessChatWithValidation`) must remain unchanged. Internal refactoring only.

---

## Success Metrics

- [ ] `ChatService` struct has fewer direct DB calls
- [ ] Each sub-service has focused responsibility
- [ ] Unit test count increases (new granular tests)
- [ ] Integration tests still pass

---

## Rollback Plan

```bash
git checkout main -- internal/services/
```

All changes are additive until the final integration step, so partial rollback is possible.
