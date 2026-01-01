# Tech Debt Task 007: Migrate Services to Use Repositories

> **Agent Prompt**: You are migrating service layer code to use repository interfaces instead of direct `db.*` calls. This task focuses on `internal/services/` files. For each service, inject the appropriate repository interface via constructor, then replace direct db calls with repository method calls. Apply strict TDD: ensure all existing tests pass after each change. Do not change behavior, only the data access pattern.

## Background

After creating repository implementations (Tasks 001-006), services still call `db.*` functions directly:

| Service | Direct db Calls |
|---------|-----------------|
| `chat_service.go` | `db.GetPlanByUserID` |
| `chat_helpers.go` | `db.GetEnabledActions`, `db.GetPlanByUserID` |
| `chat_pipeline.go` | `db.GetOrCreateConversationBySessionID`, `db.CreateMessage` |
| `chatbot_service.go` | `db.UpdateChatbot`, `db.GetChatbotByID` |
| `refresh_scheduler.go` | Multiple db calls |
| `privacy_service.go` | Many privacy-related db calls |
| `handoff_service.go` | `db.CreateHandoffRequest`, etc. |
| `analytics_service.go` | `db.GetAnalyticsOverview`, etc. |

## Goal

Refactor services to:
1. Accept repository interfaces via constructor injection
2. Replace `db.*` calls with repository method calls
3. Keep behavior identical

## Integration Plan

### Step 1: Update Service Struct

Before:
```go
type ChatService struct {
    DB     *sql.DB
    Vector rag.VectorStore
    // ...
}
```

After:
```go
type ChatService struct {
    planRepo         repository.PlanRepository
    conversationRepo repository.ConversationRepository
    actionRepo       repository.ActionRepository
    Vector           rag.VectorStore
    // ...
}
```

### Step 2: Update Constructor

Before:
```go
func NewChatService(db *sql.DB, vector rag.VectorStore) *ChatService {
    return &ChatService{DB: db, Vector: vector}
}
```

After:
```go
func NewChatService(
    planRepo repository.PlanRepository,
    conversationRepo repository.ConversationRepository,
    actionRepo repository.ActionRepository,
    vector rag.VectorStore,
) *ChatService {
    return &ChatService{
        planRepo:         planRepo,
        conversationRepo: conversationRepo,
        actionRepo:       actionRepo,
        Vector:           vector,
    }
}
```

### Step 3: Replace db Calls

Before:
```go
plan, err := db.GetPlanByUserID(ctx, s.DB, req.UserID)
```

After:
```go
plan, err := s.planRepo.GetByUserID(ctx, req.UserID)
```

### Step 4: Update Wire/DI Setup

Find where services are instantiated (likely `cmd/server/main.go` or wire setup) and update to inject repositories.

### Step 5: Update Tests

Service tests that mock `*sql.DB` should now mock repository interfaces:

```go
func TestChatService_GetPlan(t *testing.T) {
    mockPlanRepo := &repository.MockPlanRepo{
        GetByUserIDFunc: func(ctx context.Context, userID string) (*models.Plan, error) {
            return &models.Plan{Code: "pro"}, nil
        },
    }
    
    svc := NewChatService(mockPlanRepo, nil, nil, nil)
    // test...
}
```

## Services to Migrate (in order)

1. **`chatbot_service.go`** - Uses `ChatbotRepository` (already exists)
2. **`chat_helpers.go`** - Uses `ActionRepository`, `PlanRepository`
3. **`chat_service.go`** - Uses `PlanRepository`
4. **`chat_pipeline.go`** - Uses `ConversationRepository`
5. **`analytics_service.go`** - May need new `AnalyticsRepository`
6. **`refresh_scheduler.go`** - Uses multiple repositories
7. **`privacy_service.go`** - May need new `PrivacyRepository`
8. **`handoff_service.go`** - May need new `HandoffRepository`

## Edge Cases

- **Nil repository handling**: Services should panic early if nil repo injected
- **Transaction support**: Some operations span multiple repos
- **Backward compatibility**: Main.go must be updated atomically

## Checklist

### Phase 1: Simple Services
- [ ] Migrate `chatbot_service.go` to use `ChatbotRepository`
- [ ] Update `chatbot_service_test.go` with mock repos
- [ ] Migrate `chat_helpers.go` to use `ActionRepository`, `PlanRepository`
- [ ] Migrate `chat_service.go` to use `PlanRepository`

### Phase 2: Chat Pipeline
- [ ] Migrate `chat_pipeline.go` to use `ConversationRepository`
- [ ] Update chat pipeline tests

### Phase 3: Complex Services
- [ ] Migrate `analytics_service.go` (may need new interface)
- [ ] Migrate `refresh_scheduler.go`
- [ ] Migrate `privacy_service.go` (may need new interface)
- [ ] Migrate `handoff_service.go` (may need new interface)

### Phase 4: DI/Wire Update
- [ ] Update `cmd/server/main.go` to instantiate repositories
- [ ] Update any wire/DI configuration

### Verification
- [ ] Run `go build ./...`
- [ ] Run `make test-all`
- [ ] Run `make lint`

## Verification

```bash
# After each service migration
go test ./internal/services/..._test.go -v
make test-all
```
