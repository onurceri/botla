# Tech Debt Task 008: Migrate Handlers to Use Repositories

> **Agent Prompt**: You are migrating HTTP handler code to use repository interfaces instead of direct `db.*` calls. This task focuses on `internal/api/handlers/` files. For each handler struct, inject the appropriate repository interface, then replace direct db calls with repository method calls. Apply strict TDD: ensure all existing tests pass after each change. Do not change behavior, only the data access pattern.

## Background

After creating repository implementations and migrating services, handlers still call `db.*` functions directly. This bypasses the repository abstraction and creates inconsistency.

**High-usage handlers with direct db calls:**

| Handler File | Direct db Calls |
|--------------|-----------------|
| `usage.go` | `db.GetUserByID`, `db.CountChatbots*`, `db.GetMonthlyTokenUsage`, etc. |
| `chatbot_context.go` | `db.GetChatbotByID`, `db.GetSourceByID` |
| `source_create.go` | `db.GetPlanByUserID`, `db.CountSourcesByType` |
| `source_refresh.go` | `db.GetPlanByUserID`, `db.UpdateSourceForRefresh` |
| `analytics.go` | `db.GetGlobalAnalytics`, `db.GetSourceUsageStats` |
| `admin_queues.go` | `db.GetQueueStats`, `db.GetStuckJobs` |
| `training_job.go` | `db.GetJobBySourceID` |
| `chatbot_suggestions.go` | `db.CreateSuggestionJob`, `db.GetLatestSuggestionJobForChatbot` |
| `privacy.go` | Multiple privacy db calls |
| `admin.go` | Admin-specific db calls |

## Goal

Refactor handlers to:
1. Accept repository interfaces via constructor injection
2. Replace `db.*` calls with repository method calls
3. Keep behavior identical

## Integration Plan

### Step 1: Update Handler Struct

Before:
```go
type UsageHandler struct {
    DB *sql.DB
}
```

After:
```go
type UsageHandler struct {
    userRepo    repository.UserRepository
    chatbotRepo repository.ChatbotRepository
    usageRepo   repository.UsageRepository
}
```

### Step 2: Update Constructor

```go
func NewUsageHandler(
    userRepo repository.UserRepository,
    chatbotRepo repository.ChatbotRepository,
    usageRepo repository.UsageRepository,
) *UsageHandler {
    return &UsageHandler{
        userRepo:    userRepo,
        chatbotRepo: chatbotRepo,
        usageRepo:   usageRepo,
    }
}
```

### Step 3: Replace db Calls

Before:
```go
u, err := db.GetUserByID(r.Context(), h.DB, uid)
chatbotsCount, err := db.CountChatbotsByUserID(ctx, h.DB, userID)
```

After:
```go
u, err := h.userRepo.GetByID(r.Context(), uid)
chatbotsCount, err := h.chatbotRepo.CountByUserID(ctx, userID)
```

### Step 4: Update Router Configuration

Find where handlers are registered (likely `internal/api/router.go`) and update to inject repositories.

### Step 5: Update Tests

Handler tests should use mock repositories:

```go
func TestUsageHandler_GetUsage(t *testing.T) {
    mockUserRepo := &repository.MockUserRepo{
        GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
            return &models.User{ID: id, PlanCode: "pro"}, nil
        },
    }
    
    handler := NewUsageHandler(mockUserRepo, mockChatbotRepo, mockUsageRepo)
    // test...
}
```

## Handlers to Migrate (in priority order)

### High Priority (used in every request)
1. **`chatbot_context.go`** - Uses `ChatbotRepository`, `SourceRepository`

### Medium Priority (common operations)
2. **`usage.go`** - Needs `UserRepository`, `UsageRepository`
3. **`source_create.go`** - Uses `PlanRepository`, `SourceRepository`
4. **`source_refresh.go`** - Uses `PlanRepository`, `SourceRepository`
5. **`analytics.go`** - May need `AnalyticsRepository`
6. **`training_job.go`** - May need `TrainingJobRepository`

### Lower Priority (admin/specialized)
7. **`admin_queues.go`** - May need `QueueRepository`
8. **`chatbot_suggestions.go`** - May need `SuggestionRepository`
9. **`privacy.go`** - Uses `PrivacyRepository`
10. **`admin.go`** - Uses `AdminChatbotRepository`

## New Repositories Potentially Needed

Some handlers use db functions that aren't covered by existing repositories:

| Function | Potential Repository |
|----------|---------------------|
| `db.GetUserByID` | `UserRepository` |
| `db.GetMonthlyTokenUsage` | `UsageRepository` |
| `db.GetQueueStats` | `QueueRepository` |
| `db.GetJobBySourceID` | `TrainingJobRepository` |
| `db.CreateSuggestionJob` | `SuggestionRepository` |

These can be:
- Created as separate tasks, OR
- Added to existing repositories if logically related

## Edge Cases

- **Nil repository handling**: Handlers should validate repos in constructor
- **Multiple repository coordination**: Some handlers touch many tables
- **Test database sharing**: Handler tests may need test DB setup

## Checklist

### Phase 1: Core Handlers
- [ ] Migrate `chatbot_context.go` to use repositories
- [ ] Update `chatbot_context` tests
- [ ] Migrate `usage.go` to use repositories
- [ ] Update usage handler tests

### Phase 2: Source Handlers
- [ ] Migrate `source_create.go`
- [ ] Migrate `source_refresh.go`
- [ ] Migrate `source_single.go`
- [ ] Migrate `source_chatbot.go`

### Phase 3: Analytics/Jobs
- [ ] Migrate `analytics.go`
- [ ] Migrate `training_job.go`
- [ ] Migrate `chatbot_suggestions.go`

### Phase 4: Admin/Privacy
- [ ] Migrate `admin_queues.go`
- [ ] Migrate `admin.go`
- [ ] Migrate `privacy.go`

### Phase 5: Router Update
- [ ] Update `internal/api/router.go` to inject repositories
- [ ] Update integration test server setup

### Verification
- [ ] Run `go build ./...`
- [ ] Run `make test-all`
- [ ] Run `make lint`

## Verification

```bash
# After each handler migration
go test ./internal/api/handlers/..._test.go -v

# Full test suite
make test-all
```
