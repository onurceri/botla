# Implementation Plan: Complete Repository Migration (Tasks 007-009)

**Created:** January 4, 2026
**Status:** ✅ Phase 1 Complete - ✅ Phase 2 Complete - ✅ Phase 3 Complete - ✅ Phase 4 Complete - ✅ Phase 5 Complete - ✅ Phase 6 Complete

**Phase 5 (Handler Layer) - COMPLETE:**

All handlers have been migrated to use repositories:

**Migrated handlers (27 of 27):**
- ✅ me.go - Migrated to use repositories (UserRepo, OrgService)
- ✅ plan.go - Migrated to use repositories (UserRepo, PlanRepo)
- ✅ chatbot_list.go - Migrated to use repositories (ChatbotRepo, PlanRepo)
- ✅ chatbot_item.go - Migrated to use repositories (ChatbotRepo)
- ✅ source_create.go - Migrated to use repositories (PlanRepo, SourceRepo)
- ✅ source_single.go - Migrated to use repositories (SourceRepo)
- ✅ source_chatbot.go - Migrated to use repositories (SourceRepo)
- ✅ source_refresh.go - Migrated to use repositories (PlanRepo, UsageRepo, SourceRepo)
- ✅ source_utils.go - Migrated to use repositories (UsageRepo, SourceRepo)
- ✅ source_bulk.go - Migrated to use repositories (PlanRepo, SourceRepo)
- ✅ pending_urls.go - Migrated to use repositories (PendingURLRepo, SourceRepo, ChatbotRepo)
- ✅ chat.go - Migrated to use repositories (ChatService, AnalyticsRepo)
- ✅ training_job.go - Migrated to use repositories (TrainingJobRepo, SourceRepo, ChatbotRepo)
- ✅ chatbot_suggestions.go - Migrated to use repositories (SuggestionJobRepo, ChatbotRepo)
- ✅ handoff.go - Migrated to use repositories (ChatbotRepo, ConversationRepo, HandoffRepo)
- ✅ analytics.go - Migrated to use repositories (AnalyticsRepo, ChatbotRepo)
- ✅ public.go - Migrated to use repositories (ChatbotRepo, PlanRepo, UsageRepo, AnalyticsRepo)
- ✅ admin.go - Migrated to use repositories (UserRepo, OrganizationRepo)
- ✅ admin_audit.go - Migrated to use repositories (AdminRepository)
- ✅ admin_errors.go - Migrated to use repositories (AdminRepository)
- ✅ admin_sources.go - Migrated to use repositories (AdminRepository)
- ✅ admin_queues.go - Migrated to use repositories (QueueRepo, SourceRepo)
- ✅ privacy.go - Migrated to use repositories (PrivacyRepo)
- ✅ user_privacy.go - Migrated to use repositories (UserRepo, PrivacyRepo)
- ✅ onboarding.go - Migrated to use repositories (UserRepo)
- ✅ organization.go - Migrated to use repositories (UserRepo)
- ✅ chatbot_context.go - Migrated to use repositories (ChatbotRepo, SourceRepo)

This document provides a complete, step-by-step plan to:
1. Complete the missing repository methods (Phase 1)
2. Migrate all services to use repositories (Phase 2)
3. Migrate all handlers to use repositories (Phase 3)
4. Clean up the deprecated `internal/db` package (Phase 4)

---

## Executive Summary

The codebase has a hybrid architecture problem:
- ✅ Repository layer is fully implemented with Squirrel
- ✅ Services - FULLY migrated to use repositories
- ✅ Processing layer - FULLY migrated to use repositories
- ✅ RAG layer - FULLY migrated to use repositories
- ✅ Handler layer - FULLY migrated to use repositories (27 of 27 handlers)
- ✅ testdb package - FULLY migrated to use repositories
- ⚠️ Deprecated `internal/db/*` files still exist (Phase 7 pending)

**Phase 1 Status:** ✅ COMPLETE - All repository methods added
**Phase 2 Status:** ✅ COMPLETE - All services migrated to use repositories
**Phase 3 Status:** ✅ COMPLETE - All processing files migrated to use repositories
**Phase 4 Status:** ✅ COMPLETE - RAG layer migrated to use repositories
**Phase 5 Status:** ✅ COMPLETE - Handler layer 100% migrated (27 of 27 handlers)
**Phase 6 Status:** ✅ COMPLETE - testdb package migrated to use repositories

---

## Phase 1: Add Missing Repository Methods

Before migrating consumers, all repository interfaces must include every method currently used in production code.

### 1.1 TrainingJobRepository - Add Missing Methods

**File:** `internal/repository/training_job_repo.go`

Add methods to interface and implementation:

```go
// TrainingJobRepository (existing interface) - ADD these methods:

// GetPendingJobs retrieves jobs in pending status for recovery.
GetPendingJobs(ctx context.Context, limit int) ([]*models.TrainingJob, error)

// MarkStepCompleted marks a step as completed in job metadata.
MarkStepCompleted(ctx context.Context, jobID string, step models.TrainingStep, outputHash string) error

// GetLastCompletedStep returns the last completed step for resuming.
GetLastCompletedStep(ctx context.Context, jobID string) (*models.TrainingStep, error)

// Fail marks a job as failed with error details.
Fail(ctx context.Context, id string, step models.TrainingStep, errCode, errMsg string) error

// Complete marks a job as completed.
Complete(ctx context.Context, id string) error

// Cancel marks a job as cancelled.
Cancel(ctx context.Context, id string) error

// GetRetryableJobs retrieves failed jobs that can be retried.
GetRetryableJobs(ctx context.Context, maxRetries, limit int) ([]*models.TrainingJob, error)

// GetRunningJobs retrieves jobs currently running.
GetRunningJobs(ctx context.Context, limit int) ([]*models.TrainingJob, error)
```

**Reference Implementation:** `internal/db/training_job.go`

**Checklist:**
- [x] Add methods to `TrainingJobRepository` interface in `interfaces.go`
- [x] Implement `GetPendingJobs` (uses raw SQL)
- [x] Implement `MarkStepCompleted` (uses raw SQL)
- [x] Implement `GetLastCompletedStep` (uses raw SQL)
- [x] Implement `Fail` (uses raw SQL)
- [x] Implement `Complete` (uses raw SQL)
- [x] Implement `Cancel` (uses raw SQL)
- [x] Implement `GetRetryableJobs` (uses raw SQL)
- [x] Implement `GetRunningJobs` (uses raw SQL)
- [x] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify tests pass

---

### 1.2 SourceRepository - Add Missing Methods

**File:** `internal/repository/source_repo.go`

Add methods to interface and implementation:

```go
// SourceRepository (existing interface) - ADD these methods:

// UpdateSourceHash updates the content hash for a source.
UpdateSourceHash(ctx context.Context, id string, hash string) error

// UpdateSourceProcessing updates processing status, error, chunk count, and processed_at.
UpdateSourceProcessing(ctx context.Context, id string, status string, errorMessage *string, chunkCount int, processedAt *time.Time) error

// UpdateSourceCapability updates the capability summary for a source.
UpdateSourceCapability(ctx context.Context, id string, summary string) error

// UpdateSourceSuggestions updates the suggested questions for a source.
UpdateSourceSuggestions(ctx context.Context, id string, suggestions []string) error
```

**Reference Implementation:** `internal/db/source.go`

**Checklist:**
- [x] Add methods to `SourceRepository` interface in `interfaces.go`
- [x] Implement `UpdateSourceHash` using Squirrel
- [x] Implement `UpdateSourceProcessing` using Squirrel
- [x] Implement `UpdateSourceCapability` using Squirrel
- [x] Implement `UpdateSourceSuggestions` using Squirrel
- [x] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify tests pass

---

### 1.3 UsageRepository - Add Missing Methods

**File:** `internal/repository/usage_repo.go`

Add methods to interface and implementation:

```go
// UsageRepository (existing interface) - ADD these methods:

// ReserveChatTokens atomically reserves tokens for a chat request.
// Returns ErrTokenQuotaExceeded if the reservation would exceed the limit.
ReserveChatTokens(ctx context.Context, userID string, estimatedTokens int, maxMonthlyTokens int) error

// AdjustChatTokens adjusts the token count after a chat request completes.
AdjustChatTokens(ctx context.Context, userID string, deltaTokens int) error

// GetMonthlyChatTokens returns the current monthly chat token usage.
GetMonthlyChatTokens(ctx context.Context, userID string) (int, error)

// IncrementChatTokens adds to the chat_tokens counter (no limit check).
IncrementChatTokens(ctx context.Context, userID string, tokens int) error
```

**Reference Implementation:** `internal/db/usage_chat_tokens.go`, `internal/db/usage_ingestions.go`

**Checklist:**
- [x] Add methods to `UsageRepository` interface in `interfaces.go`
- [x] Implement `ReserveChatTokens` using raw SQL (with ErrTokenQuotaExceeded)
- [x] Implement `AdjustChatTokens` using raw SQL
- [x] Implement `GetMonthlyChatTokens` using raw SQL
- [x] Implement `IncrementChatTokens` using raw SQL
- [x] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify tests pass

---

### 1.4 PlanRepository - Add Missing Methods

**File:** `internal/repository/plan_repo.go`

Add methods to interface and implementation:

```go
// PlanRepository (existing interface) - ADD these methods:

// GetPlanWithLimits retrieves a plan by user ID with all limits populated.
GetPlanWithLimits(ctx context.Context, userID string) (*models.Plan, error)

// GetAllPlansWithLimits retrieves all active plans with their limits.
GetAllPlansWithLimits(ctx context.Context) ([]models.Plan, error)
```

**Reference Implementation:** `internal/db/plan.go`, `internal/db/plan_limits.go`

**Checklist:**
- [x] Add methods to `PlanRepository` interface in `interfaces.go`
- [x] Implement `GetPlanWithLimits` using Squirrel
- [x] Implement `GetAllPlansWithLimits` using Squirrel
- [x] Run `go build ./...` to verify compilation
- [x] Run repository tests to verify implementations pass

---

### 1.5 AnalyticsRepository - Add Missing Methods

**File:** `internal/repository/analytics_repo.go`

Add methods to interface and implementation:

```go
// AnalyticsRepository (existing interface) - ADD these methods:

// UpdateMessageFeedback updates feedback for a message and returns affected chatbot ID.
UpdateMessageFeedback(ctx context.Context, messageID string, thumbsUp bool) (chatbotID string, oldThumbsUp bool, error)

// IncrementFeedback increments positive or negative feedback counters.
IncrementFeedback(ctx context.Context, chatbotID string, oldThumbsUp bool, newThumbsUp bool) error
```

**Reference Implementation:** `internal/db/analytics.go`, `internal/db/source_analytics.go`

**Checklist:**
- [x] Add methods to `AnalyticsRepository` interface in `interfaces.go`
- [x] Implement `UpdateMessageFeedback` using Squirrel
- [x] Implement `IncrementFeedback` using Squirrel
- [x] Run `go build ./...` to verify compilation
- [x] Run repository tests to verify implementations pass

---

### 1.6 AdminRepository - Create New Repository

**File:** `internal/repository/admin_repo.go` (NEW)

Create a new repository for admin operations:

```go
// AdminRepository defines the interface for admin-only data access operations.
type AdminRepository interface {
    // Audit Log Operations
    InsertAuditLog(ctx context.Context, entry AuditLogEntry) error
    ListAuditLogs(ctx context.Context, filter AuditFilter, limit, offset int) ([]AuditLogEntry, int, error)

    // Source Management
    AdminListSources(ctx context.Context, filter AdminSourceFilter, limit, offset int) ([]AdminSource, int, error)
    AdminGetSourceByID(ctx context.Context, id string) (*AdminSource, error)
    AdminGetSourceStats(ctx context.Context) (*SourceStats, error)
    AdminReprocessSource(ctx context.Context, id string) error

    // Error Logs
    ListErrorLogs(ctx context.Context, severity string, limit, offset int) ([]ErrorLog, int, error)
    GetErrorLogByID(ctx context.Context, id string) (*ErrorLog, error)
    GetErrorStats(ctx context.Context) (*ErrorStats, error)

    // Queue Operations
    GetQueueStats(ctx context.Context) ([]QueueStats, error)
    GetStuckJobs(ctx context.Context, threshold time.Duration) ([]StuckJob, error)
}
```

**Supporting Types (in `interfaces.go`):**
```go
type AuditLogEntry struct { ... }
type AuditFilter struct { ... }
type AdminSource struct { ... }
type AdminSourceFilter struct { ... }
type SourceStats struct { ... }
type ErrorLog struct { ... }
type ErrorStats struct { ... }
type QueueStats struct { ... }
type StuckJob struct { ... }
```

**Reference Implementation:** `internal/db/admin_audit.go`, `internal/db/admin_sources.go`, `internal/db/admin_errors.go`, `internal/db/admin_queue.go`

**Checklist:**
- [x] Define `AdminRepository` interface in `interfaces.go`
- [x] Define all supporting types in `interfaces.go`
- [x] Create `internal/repository/admin_repo.go` with `PostgresAdminRepo`
- [x] Implement `InsertAuditLog` using Squirrel
- [x] Implement `ListAuditLogs` using Squirrel
- [x] Implement `AdminListSources` using Squirrel
- [x] Implement `AdminGetSourceByID` using Squirrel
- [x] Implement `AdminGetSourceStats` using Squirrel
- [x] Implement `AdminReprocessSource` using Squirrel
- [x] Implement `ListErrorLogs` using Squirrel
- [x] Implement `GetErrorLogByID` using Squirrel
- [x] Implement `GetErrorStats` using Squirrel
- [N/A] Implement `GetQueueStats` using Squirrel (already in QueueRepository)
- [N/A] Implement `GetStuckJobs` using Squirrel (already in QueueRepository)
- [x] Create mock implementation in `mock_admin_repo.go`
- [x] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify tests pass

---

### 1.7 PendingURLRepository - Create New Repository

**File:** `internal/repository/pending_url_repo.go` (NEW)

Create a new repository for pending URL operations:

```go
// PendingURLRepository defines the interface for pending discovered URL operations.
type PendingURLRepository interface {
    // InsertPendingURL adds a URL to the pending list for approval.
    InsertPendingURL(ctx context.Context, chatbotID string, sourceID *string, url string) error

    // ListPendingURLs returns pending URLs for a chatbot with pagination.
    ListPendingURLs(ctx context.Context, chatbotID string, limit, offset int) ([]models.PendingURL, error)

    // CountPendingURLs returns the total count of pending URLs for a chatbot.
    CountPendingURLs(ctx context.Context, chatbotID string) (int, error)

    // UpdatePendingURLStatus updates the status of multiple pending URLs.
    UpdatePendingURLStatus(ctx context.Context, chatbotID string, urlIDs []string, status string) (int, error)

    // GetPendingURLsByIDs returns pending URLs by their IDs.
    GetPendingURLsByIDs(ctx context.Context, chatbotID string, urlIDs []string) ([]models.PendingURL, error)

    // DeletePendingURLsByChatbot clears all pending URLs for a chatbot.
    DeletePendingURLsByChatbot(ctx context.Context, chatbotID string) (int, error)
}
```

**Reference Implementation:** `internal/db/pending_url.go`

**Checklist:**
- [x] Define `PendingURLRepository` interface in `interfaces.go`
- [x] Create `internal/repository/pending_url_repo.go` with `PostgresPendingURLRepo`
- [x] Implement all methods using Squirrel/raw SQL
- [x] Create mock implementation in `mock_pending_url_repo.go`
- [x] Run `go build ./...` to verify compilation
- [ ] Run `make test-all` to verify tests pass

---

### 1.8 PrivacyRepository - Add Missing Methods

**File:** `internal/repository/privacy_repo.go`

Add methods to interface and implementation:

```go
// PrivacyRepository (existing interface) - ADD these methods:

// GetUserConsents retrieves all consent records for a user.
GetUserConsents(ctx context.Context, userID string) ([]UserConsent, error)

// UpsertConsent creates or updates a consent record.
UpsertConsent(ctx context.Context, userID string, consentType string, granted bool, ipAddress, userAgent string) error

// AnonymizeUserData anonymizes a user's personal data and deletes their content.
AnonymizeUserData(ctx context.Context, userID string) error

// GetUserFilesForDeletion returns file paths that should be deleted from storage.
GetUserFilesForDeletion(ctx context.Context, userID string) ([]string, error)
```

**Reference Implementation:** `internal/db/privacy.go`, `internal/db/consent.go`

**Checklist:**
- [x] Add methods to `PrivacyRepository` interface in `interfaces.go`
- [x] Implement `GetUserConsents` using Squirrel
- [x] Implement `UpsertConsent` using Squirrel
- [x] Implement `AnonymizeUserData` using Squirrel
- [x] Implement `GetUserFilesForDeletion` using Squirrel
- [x] Run `go build ./...` to verify compilation
- [x] Run repository tests to verify implementations pass

---

### 1.9 ConversationRepository - Add Missing Methods

**File:** `internal/repository/conversation_repo.go`

Add methods to interface and implementation:

```go
// ConversationRepository (existing interface) - ADD these methods:

// IncrementMessageCount atomically increments the message count.
IncrementMessageCount(ctx context.Context, conversationID string) error

// SaveMessageSources persists source usage for a message.
SaveMessageSources(ctx context.Context, messageID string, sources []models.ChunkMetadata) error
```

**Reference Implementation:** `internal/db/conversation.go`

**Checklist:**
- [x] Add methods to `ConversationRepository` interface in `interfaces.go`
- [x] Implement `IncrementMessageCount` using Squirrel
- [x] Implement `SaveMessageSources` using Squirrel
- [x] Run `go build ./...` to verify compilation
- [x] Run repository tests to verify implementations pass

---

### Phase 1 Summary

After completing Phase 1:
- ✅ All repository interfaces are complete
- ✅ All production `db.*` calls can be mapped to repository methods
- ✅ All repository implementations use Squirrel
- ✅ All repository methods have mock implementations
- ✅ All repository methods compile successfully
- ✅ Repository tests pass (integration tests require running database)

---

## Phase 2: Migrate Services to Use Repositories

This phase updates `internal/services/` files to use repository interfaces instead of direct `db.*` calls.

### 2.1 Identify Repository Dependencies for Each Service

**Service Files and Required Repositories:**

| Service File | Required Repositories |
|--------------|----------------------|
| `admin_service.go` | `AdminRepository`, `UserRepository`, `OrganizationRepository` |
| `quota_enforcer.go` | `UsageRepository` |
| `plan_service.go` | `PlanRepository` |
| `chat_service.go` | `PlanRepository` |
| `chat_helpers.go` | `ActionRepository`, `PlanRepository` |
| `chat_pipeline.go` | `ConversationRepository` |
| `chatbot_service.go` | `ChatbotRepository`, `SourceRepository` |
| `analytics_service.go` | `AnalyticsRepository` |
| `handoff_service.go` | `HandoffRepository` |
| `privacy_service.go` | `PrivacyRepository`, `UserRepository` |
| `refresh_scheduler.go` | `ChatbotRepository`, `SourceRepository`, `PlanRepository` |

### 2.2 Migration Pattern for Each Service

**Step 1: Update Service Struct**

Before:
```go
type AdminService struct {
    DB          *sql.DB
    VectorStore rag.VectorStore
}
```

After:
```go
type AdminService struct {
    adminRepo    repository.AdminRepository
    userRepo     repository.UserRepository
    vectorStore  rag.VectorStore
}
```

**Step 2: Update Constructor**

Before:
```go
func NewAdminService(db *sql.DB, vector rag.VectorStore) *AdminService {
    return &AdminService{DB: db, VectorStore: vector}
}
```

After:
```go
func NewAdminService(
    adminRepo repository.AdminRepository,
    userRepo repository.UserRepository,
    vectorStore rag.VectorStore,
) *AdminService {
    return &AdminService{
        adminRepo:   adminRepo,
        userRepo:    userRepo,
        vectorStore: vectorStore,
    }
}
```

**Step 3: Replace db Calls**

Before:
```go
err := db.InsertAuditLog(ctx, h.DB, entry)
```

After:
```go
err := s.adminRepo.InsertAuditLog(ctx, entry)
```

### 2.3 Detailed Service Migrations

#### 2.3.1 Migrate `admin_service.go`

**File:** `internal/services/admin_service.go`

**Changes:**
1. Add `AdminRepository` to struct
2. Update constructor to accept `AdminRepository`
3. Replace `db.InsertAuditLog` → `adminRepo.InsertAuditLog`
4. Update tests to use mock repository

**Checklist:**
- [x] Add `AdminRepository` to struct
- [x] Update constructor signature
- [x] Replace `db.InsertAuditLog` calls
- [x] Update service tests with mock
- [x] Run `go build ./...`
- [ ] Run `go test ./internal/services/admin_service_test.go -v`

---

#### 2.3.2 Migrate `quota_enforcer.go`

**File:** `internal/services/quota_enforcer.go`

**Changes:**
1. Add `UsageRepository` to struct
2. Update constructor to accept `UsageRepository`
3. Replace `db.ReserveChatTokens` → `usageRepo.ReserveChatTokens`
4. Replace `db.AdjustChatTokens` → `usageRepo.AdjustChatTokens`
5. Update tests to use mock repository

**Checklist:**
- [x] Add `UsageRepository` to struct
- [x] Update constructor signature
- [x] Replace `db.ReserveChatTokens` calls
- [x] Replace `db.AdjustChatTokens` calls
- [x] Update service tests with mock
- [x] Run `go build ./...`
- [ ] Run `go test ./internal/services/quota_enforcer_test.go -v`

---

#### 2.3.3 Migrate `plan_service.go`

**File:** `internal/services/plan_service.go`

**Changes:**
1. Add `PlanRepository` to struct
2. Update constructor to accept `PlanRepository`
3. Replace `db.GetPlanWithLimits` → `planRepo.GetPlanWithLimits`
4. Replace `db.GetAllPlansWithLimits` → `planRepo.GetAllPlansWithLimits`
5. Update tests to use mock repository

**Checklist:**
- [x] Add `PlanRepository` to struct
- [x] Update constructor signature
- [x] Replace `db.GetPlanWithLimits` calls
- [x] Replace `db.GetAllPlansWithLimits` calls
- [x] Update service tests with mock
- [x] Run `go build ./...`
- [ ] Run `go test ./internal/services/plan_service_test.go -v`

---

#### 2.3.4 Migrate `chat_helpers.go`

**File:** `internal/services/chat_helpers.go`

**Changes:**
1. Add `ActionRepository`, `PlanRepository` to struct
2. Update constructor to accept repositories
3. Replace `db.GetEnabledActions` → `actionRepo.ListEnabled`
4. Replace `db.GetPlanByUserID` → `planRepo.GetByUserID`
5. Update tests to use mock repositories

**Checklist:**
- [ ] Add `ActionRepository`, `PlanRepository` to struct
- [ ] Update constructor signature
- [ ] Replace `db.GetEnabledActions` calls
- [ ] Replace `db.GetPlanByUserID` calls
- [ ] Update service tests with mock
- [ ] Run `go build ./...`
- [ ] Run `go test ./internal/services/chat_helpers_test.go -v`

---

#### 2.3.5 Migrate `chat_pipeline.go`

**File:** `internal/services/chat_pipeline.go`

**Changes:**
1. Add `ConversationRepository` to struct
2. Update constructor to accept repository
3. Replace `db.GetOrCreateConversationBySessionID` → `conversationRepo.GetOrCreateBySessionID`
4. Replace `db.CreateMessage` → `conversationRepo.CreateMessage`
5. Update tests to use mock repository

**Checklist:**
- [ ] Add `ConversationRepository` to struct
- [ ] Update constructor signature
- [ ] Replace `db.GetOrCreateConversationBySessionID` calls
- [ ] Replace `db.CreateMessage` calls
- [ ] Update service tests with mock
- [ ] Run `go build ./...`
- [ ] Run `go test ./internal/services/chat_pipeline_test.go -v`

---

#### 2.3.6 Migrate `chatbot_service.go`

**File:** `internal/services/chatbot_service.go`

**Changes:**
1. Add `ChatbotRepository`, `SourceRepository` to struct
2. Update constructor to accept repositories
3. Replace `db.GetChatbotByID` → `chatbotRepo.GetByID`
4. Replace `db.UpdateChatbot` → `chatbotRepo.Update`
5. Update tests to use mock repositories

**Checklist:**
- [ ] Add `ChatbotRepository`, `SourceRepository` to struct
- [ ] Update constructor signature
- [ ] Replace `db.GetChatbotByID` calls
- [ ] Replace `db.UpdateChatbot` calls
- [ ] Update service tests with mock
- [ ] Run `go build ./...`
- [ ] Run `go test ./internal/services/chatbot_service_test.go -v`

---

#### 2.3.7 Migrate `refresh_scheduler.go`

**File:** `internal/services/refresh_scheduler.go`

**Changes:**
1. Add `ChatbotRepository`, `SourceRepository`, `PlanRepository` to struct
2. Update constructor to accept repositories
3. Replace `db.GetPlanByUserID` → `planRepo.GetByUserID`
4. Replace `db.GetDueForRefresh` → `chatbotRepo.GetDueForRefresh`
5. Update tests to use mock repositories

**Checklist:**
- [ ] Add repositories to struct
- [ ] Update constructor signature
- [ ] Replace `db.GetPlanByUserID` calls
- [ ] Replace `db.GetDueForRefresh` calls
- [ ] Update service tests with mock
- [ ] Run `go build ./...`
- [ ] Run `go test ./internal/services/refresh_scheduler_test.go -v`

---

#### 2.3.8 Migrate Other Services

Remaining services to migrate:
- `handoff_service.go` → `HandoffRepository`
- `analytics_service.go` → `AnalyticsRepository`
- `privacy_service.go` → `PrivacyRepository`

**Pattern:** Same as above for each service.

---

### 2.4 Update Dependency Injection (main.go)

**File:** `cmd/server/main.go`

After all services are migrated, update the main.go to:

1. Create repository instances:
```go
adminRepo := repository.NewPostgresAdminRepo(db)
planRepo := repository.NewPostgresPlanRepo(db)
// ... etc
```

2. Pass repositories to service constructors:
```go
adminService := services.NewAdminService(adminRepo, userRepo, vectorStore)
quotaService := services.NewQuotaEnforcer(usageRepo)
// ... etc
```

**Checklist:**
- [x] Update all repository instantiations
- [x] Update all service instantiations
- [x] Run `go build ./cmd/server/...`
- [x] Run `go test ./internal/repository/...`
- [x] Run `go test ./internal/services/...`

---

### Phase 2 Summary

After completing Phase 2:
- ✅ AdminService - Migrated to use AdminRepository
- ✅ PlanService - Migrated to use PlanRepository
- ✅ QuotaEnforcer - Migrated to use UsageRepository
- ✅ cmd/server/main.go - Updated to create repositories
- ✅ internal/api/router/router.go - Updated to create repositories
- ✅ internal/integration/fixtures/server.go - Updated to create repositories
- ✅ ChatService - Migrated (PlanRepo, ConversationRepo, AnalyticsRepo, ActionRepo)
- ✅ ChatbotService - Migrated (ChatbotRepo, PlanRepo)
- ✅ ChatHelpers - Migrated (uses repositories through ChatService)
- ✅ ChatPipeline - Migrated (ConversationRepo)
- ✅ RefreshScheduler - Migrated (ChatbotRepo, SourceRepo, PlanRepo, AnalyticsRepo)
- ✅ HandoffService - Migrated (HandoffRepo, ConversationRepo, AnalyticsRepo)
- ✅ AnalyticsService - Migrated (AnalyticsRepo)
- ✅ PrivacyService - Migrated (PrivacyRepo)

- ✅ Processing layer - FULLY migrated (6 of 6 files complete)
- ✅ RAG layer - FULLY migrated (tool_executor.go complete)

---

## Phase 3: Migrate Processing Layer to Use Repositories

This phase updates `internal/processing/` files to use repository interfaces.

### 3.1 Processing Files and Required Repositories

| Processing File | Required Repositories |
|-----------------|----------------------|
| `job_processor.go` | `TrainingJobRepository`, `SourceRepository`, `ChatbotRepository`, `PlanRepository` |
| `sources_queue.go` | `TrainingJobRepository` |
| `url_processor.go` | `TrainingJobRepository`, `SourceRepository`, `UsageRepository`, `PendingURLRepository` |
| `text_processor.go` | `TrainingJobRepository`, `SourceRepository`, `UsageRepository` |
| `pdf_processor.go` | `TrainingJobRepository`, `SourceRepository`, `UsageRepository` |
| `suggestions.go` | `ChatbotRepository`, `SuggestionJobRepository` |

### 3.2 Migration Pattern

**Pattern:** Same as services - add repository to struct, update constructor, replace db calls.

### 3.3 Detailed Processing Migrations

#### 3.3.1 Migrate `job_processor.go`

**File:** `internal/processing/job_processor.go`

**Changes:**
1. Add repositories to struct
2. Update constructor to accept repositories
3. Replace `db.GetTrainingJob` → `trainingJobRepo.GetByID`
4. Replace `db.GetSourceByID` → `sourceRepo.GetByID`
5. Replace `db.GetChatbotByID` → `chatbotRepo.GetByID`
6. Replace `db.UpdateJobStatus` → `trainingJobRepo.UpdateJobStatus`
7. Replace `db.UpdateSourceProcessing` → `sourceRepo.UpdateSourceProcessing`
8. Replace `db.IncrementRetryCount` → `trainingJobRepo.ResetForRetry`

**Checklist:**
- [ ] Add repositories to struct
- [ ] Update constructor signature
- [ ] Replace all `db.*` calls with repository calls
- [ ] Run `go build ./...`
- [ ] Run tests for processing layer

---

#### 3.3.2 Migrate `sources_queue.go`

**File:** `internal/processing/sources_queue.go`

**Changes:**
1. Add `TrainingJobRepository` to struct
2. Update constructor
3. Replace `db.CreateTrainingJob` → `trainingJobRepo.Create`
4. Replace `db.FailJob` → `trainingJobRepo.Fail`
5. Replace `db.GetPendingJobs` → `trainingJobRepo.GetPendingJobs`

**Checklist:**
- [ ] Add `TrainingJobRepository` to struct
- [ ] Update constructor signature
- [ ] Replace all `db.*` calls
- [ ] Run `go build ./...`
- [ ] Run tests for sources_queue

---

#### 3.3.3 Migrate `url_processor.go`

**File:** `internal/processing/url_processor.go`

**Changes:**
1. Add repositories to struct
2. Update constructor
3. Replace `db.MarkStepCompleted` → `trainingJobRepo.MarkStepCompleted`
4. Replace `db.IncrementSuccessfulIngestion` → `usageRepo.IncrementSuccessfulIngestion`
5. Replace `db.AddEmbeddingTokens` → `usageRepo.AddEmbeddingTokens`
6. Replace `db.UpdateSourceCapability` → `sourceRepo.UpdateSourceCapability`
7. Replace `db.UpdateSourceSuggestions` → `sourceRepo.UpdateSourceSuggestions`
8. Replace `db.CountSourcesByType` → `sourceRepo.CountByType`
9. Replace `db.SourceExists` → `sourceRepo.Exists`
10. Replace `db.CreateDiscoveredSource` → `sourceRepo.Create`
11. Replace `db.InsertPendingURL` → `pendingURLRepo.InsertPendingURL`
12. Replace `db.UpdateSourceHash` → `sourceRepo.UpdateSourceHash`

**Checklist:**
- [ ] Add repositories to struct
- [ ] Update constructor signature
- [ ] Replace all `db.*` calls
- [ ] Run `go build ./...`
- [ ] Run tests for url_processor

---

#### 3.3.4 Migrate `text_processor.go` and `pdf_processor.go`

**File:** `internal/processing/text_processor.go`, `internal/processing/pdf_processor.go`

**Changes:** Same pattern as url_processor.go

**Checklist:**
- [ ] Add repositories to struct
- [ ] Update constructor signature
- [ ] Replace all `db.*` calls
- [ ] Run `go build ./...`
- [ ] Run tests for text_processor and pdf_processor

---

#### 3.3.5 Migrate `suggestions.go`

**File:** `internal/processing/suggestions.go`

**Changes:**
1. Add repositories to struct
2. Update constructor
3. Replace `db.UpdateChatbotSuggestedQuestions` → `chatbotRepo.UpdateSuggestedQuestions`
4. Replace `db.UpdateSuggestionJobStatus` → `suggestionJobRepo.UpdateStatus`
5. Replace `db.FailSuggestionJob` → `suggestionJobRepo.Fail`
6. Replace `db.CompleteSuggestionJob` → `suggestionJobRepo.Complete`

**Checklist:**
- [ ] Add repositories to struct
- [ ] Update constructor signature
- [ ] Replace all `db.*` calls
- [ ] Run `go build ./...`
- [ ] Run tests for suggestions

---

## Phase 4: Migrate RAG Layer to Use Repositories

This phase updates `internal/rag/` files to use repository interfaces.

### 4.1 RAG Files and Required Repositories

| RAG File | Required Repositories |
|----------|-----------------------|
| `tool_executor.go` | `SourceRepository`, `HandoffRepository`, `ActionRepository` |

### 4.2 Migration

**File:** `internal/rag/tool_executor.go`

**Changes:**
1. Add repositories to struct (SourceRepository, HandoffRepository, ActionRepository)
2. Create NewToolExecutor constructor
3. Replace `db.ListSourcesByChatbotID` → `sourceRepo.GetByChatbot`
4. Replace `db.HasActiveHandoffRequest` → `handoffRepo.HasActiveHandoffRequest`
5. Replace `db.CreateHandoffRequest` → `handoffRepo.CreateHandoffRequest`

**Checklist:**
- [x] Add repositories to struct
- [x] Update constructor signature
- [x] Replace all `db.*` calls
- [x] Run `go build ./...`
- [x] Run tests for rag layer

---

## Phase 5: Migrate Handlers to Use Repositories

This phase updates `internal/api/handlers/` files to use repository interfaces.

### 5.1 Handler Migration Pattern

**Pattern:** Same as services - add repositories to handler struct, update constructor, replace db calls.

### 5.2 Priority Order

1. **High Priority** (frequently used):
   - `chatbot_context.go`
   - `usage.go`
   - `me.go`

2. **Medium Priority** (common operations):
   - `source_create.go`
   - `source_single.go`
   - `source_chatbot.go`
   - `source_refresh.go`
   - `source_utils.go`
   - `source_bulk.go`
   - `pending_urls.go`
   - `chat.go`
   - `chatbot_item.go`
   - `chatbot_list.go`
   - `plan.go`

3. **Lower Priority** (specialized operations):
   - `training_job.go`
   - `chatbot_suggestions.go`
   - `handoff.go`
   - `analytics.go`
   - `public.go`

4. **Admin/Privacy** (admin/specialized):
   - `admin.go`
   - `admin_audit.go`
   - `admin_errors.go`
   - `admin_sources.go`
   - `admin_queues.go`
   - `privacy.go`
   - `user_privacy.go`
   - `onboarding.go`

### 5.3 Example Handler Migration

**File:** `internal/api/handlers/usage.go`

**Before:**
```go
type UsageHandler struct {
    DB *sql.DB
}

func NewUsageHandler(db *sql.DB) *UsageHandler {
    return &UsageHandler{DB: db}
}
```

**After:**
```go
type UsageHandler struct {
    userRepo    repository.UserRepository
    chatbotRepo repository.ChatbotRepository
    usageRepo   repository.UsageRepository
}

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

**db → repo replacements:**
- `db.GetUserByID` → `userRepo.GetByID`
- `db.CountChatbotsByUserID` → `chatbotRepo.CountByUserID`
- `db.GetMonthlyTokenUsage` → `usageRepo.GetMonthlyTokenUsage`

### 5.4 Update Router Configuration

**File:** `internal/api/router.go`

After all handlers are migrated, update the router to inject repositories:

```go
func NewRouter(db *sql.DB) *chi.Mux {
    // Create repositories
    userRepo := repository.NewPostgresUserRepo(db)
    chatbotRepo := repository.NewPostgresChatbotRepo(db)
    usageRepo := repository.NewPostgresUsageRepo(db)
    // ... etc

    // Create handlers with repositories
    usageHandler := handlers.NewUsageHandler(userRepo, chatbotRepo, usageRepo)
    // ... etc

    // Register routes
    r.Get("/api/usage", usageHandler.GetUsage)
    // ... etc
}
```

---

### Phase 6: Update testdb Package

**File:** `internal/testdb/fixtures.go`

**Changes:**
1. ✅ Replace `db.GetUserByID` → `UserRepository.GetByID`
2. ✅ Remove unused `internal/db` import
3. ✅ Update `CreateOrganization` function to use repository pattern

**Checklist:**
- [x] Update `CreateOrganization` to use repositories
- [x] Replace all `db.*` calls in fixtures
- [x] Run `go build ./...`
- [x] Run integration tests

---

### Phase 6 Summary

After completing Phase 6:
- ✅ testdb/fixtures.go - Updated to use repository pattern
- ✅ Removed deprecated `db.GetUserByID` call in CreateOrganization
- ✅ Removed unused `internal/db` import
- ✅ All testdb tests pass

**Phase 6 Status:** ✅ COMPLETE

---

## Phase 7: Clean Up Deprecated Files

**⚠️ WARNING:** Only proceed with this phase after ALL migrations are complete and tests pass.

### 7.1 Delete internal/db/* Files

**Delete all files in `internal/db/` EXCEPT:**
- `db.go` (contains connection setup and New function)

**Files to DELETE:**
```
internal/db/action.go
internal/db/action_logs.go
internal/db/admin_chatbots.go
internal/db/admin_users.go
internal/db/admin_audit.go
internal/db/admin_stats.go
internal/db/admin_sources.go
internal/db/admin_orgs.go
internal/db/admin_queue.go
internal/db/admin_errors.go
internal/db/chatbot.go
internal/db/source.go
internal/db/conversation.go
internal/db/plan.go
internal/db/plan_limits.go
internal/db/user.go
internal/db/handoff.go
internal/db/privacy.go
internal/db/consent.go
internal/db/training_job.go
internal/db/suggestion_job.go
internal/db/pending_url.go
internal/db/message_sources.go
internal/db/chatbot_cache.go
internal/db/chatbot_refresh.go
internal/db/usage_chat_tokens.go
internal/db/usage_ingestions.go
internal/db/source_analytics.go
internal/db/analytics.go
internal/db/privacy_export.go
```

### 7.2 Delete internal/db/*_test.go Files

All test files for deleted functions should also be deleted:
```
internal/db/*_test.go
```

### 7.3 Verification After Cleanup

```bash
# Verify no compilation errors
go build ./...

# Verify no remaining db.* calls
grep -r "internal/db" internal/ --include="*.go" | grep -v "_test.go" | grep -v "internal/db/db.go"
# Should only show: internal/db/db.go

# Run all tests
make test-all

# Verify code coverage is maintained
make cover-gate
```

---

## Complete Checklist

### Phase 1: Repository Methods
- [ ] TrainingJobRepository - Add missing methods
- [ ] SourceRepository - Add missing methods
- [ ] UsageRepository - Add missing methods
- [ ] PlanRepository - Add missing methods
- [ ] AnalyticsRepository - Add missing methods
- [ ] AdminRepository - Create new repository
- [ ] PendingURLRepository - Create new repository
- [ ] PrivacyRepository - Add missing methods
- [ ] ConversationRepository - Add missing methods

### Phase 2: Services
- [x] admin_service.go - Migrated to repositories
- [x] quota_enforcer.go - Migrated to repositories
- [x] plan_service.go - Migrated to repositories
- [x] chat_helpers.go - Migrated to repositories
- [x] chat_pipeline.go - Migrated to repositories
- [x] chatbot_service.go - Migrated to repositories
- [x] refresh_scheduler.go - Migrated to repositories
- [x] handoff_service.go - Migrated to repositories
- [x] analytics_service.go - Migrated to repositories
- [x] privacy_service.go - Migrated to repositories
- [x] Update main.go dependency injection

### Phase 3: Processing Layer
- [x] job_processor.go - Migrated to repositories
- [x] sources_queue.go - Migrated to repositories
- [x] url_processor.go - Migrated to repositories
- [x] text_processor.go - Migrated to repositories
- [x] pdf_processor.go - Migrated to repositories
- [x] suggestions.go - Migrated to repositories ⬅️ **JUST COMPLETED**

### Phase 4: RAG Layer
- [x] tool_executor.go - Migrate to repositories

### Phase 5: Handlers
- [x] chatbot_context.go - Migrated to use repositories (ChatbotRepo, SourceRepo)
- [x] usage.go - Already using repositories
- [x] me.go - Migrated to repositories (UserRepo, OrgService)
- [x] source_create.go - Migrated to repositories (PlanRepo, SourceRepo)
- [x] source_single.go - Migrated to repositories (SourceRepo)
- [x] source_chatbot.go - Migrated to repositories (SourceRepo)
- [x] source_refresh.go - Migrated to repositories (PlanRepo, UsageRepo, SourceRepo)
- [x] source_utils.go - Migrated to repositories (UsageRepo, SourceRepo)
- [x] source_bulk.go - Migrated to repositories (PlanRepo, SourceRepo)
- [x] pending_urls.go - Migrated to repositories (PendingURLRepo, SourceRepo, ChatbotRepo)
- [x] chat.go - Migrated to repositories (ChatService, AnalyticsRepo)
- [x] chatbot_item.go - Migrated to repositories (ChatbotRepo)
- [x] chatbot_list.go - Migrated to repositories (ChatbotRepo, PlanRepo)
- [x] plan.go - Migrated to repositories (UserRepo, PlanRepo)
- [x] training_job.go - Migrated to repositories (TrainingJobRepo, SourceRepo, ChatbotRepo)
- [x] chatbot_suggestions.go - Migrated to repositories (SuggestionJobRepo, ChatbotRepo)
- [x] handoff.go - Migrated to repositories (ChatbotRepo, ConversationRepo, HandoffRepo)
- [x] analytics.go - Migrated to repositories (AnalyticsRepo, ChatbotRepo)
- [x] public.go - Migrated to repositories (ChatbotRepo, PlanRepo, UsageRepo, AnalyticsRepo)
- [x] admin.go - Migrated to repositories (UserRepo, OrganizationRepo)
- [x] admin_audit.go - Migrated to repositories (AdminRepository)
- [x] admin_errors.go - Migrated to repositories (AdminRepository)
- [x] admin_sources.go - Migrated to repositories (AdminRepository)
- [x] admin_queues.go - Migrated to repositories (QueueRepo, SourceRepo)
- [x] privacy.go - Migrated to repositories (PrivacyRepo)
- [x] user_privacy.go - Migrated to repositories (UserRepo, PrivacyRepo)
- [x] onboarding.go - Migrated to repositories (UserRepo)
- [x] organization.go - Migrated to repositories (UserRepo)
- [x] Update router.go dependency injection - Completed

### Phase 5 Summary (January 4, 2026)
**Migrated handlers (27 of 27):** All handlers have been fully migrated to use repositories.

**Summary of changes:**
- admin_audit.go - Uses AdminRepository.ListAuditLogs
- admin_errors.go - Uses AdminRepository.ListErrorLogs, GetErrorLogByID, GetErrorStats
- admin_sources.go - Uses AdminRepository.AdminListSources, AdminGetSourceByID, AdminGetSourceStats, AdminReprocessSource
- chatbot_context.go - Uses ChatbotRepository.GetByID and SourceRepository.GetByID
- organization.go - Uses UserRepository.GetByEmail
- source_create.go - Uses SourceRepository.GetLastDeletedAtForURL (NEW method added)
- All other handlers already using repositories

**Progress:** 100% - Phase 5 COMPLETE

### Phase 6: testdb
- [x] fixtures.go - Update to use repositories

### Phase 7: Cleanup
- [ ] Delete all internal/db/* files (except db.go)
- [ ] Delete all internal/db/*_test.go files
- [ ] Verify build succeeds
- [ ] Verify all tests pass
- [ ] Verify code coverage maintained

---

## Execution Strategy

### Step 1: Start with Phase 1
Complete all repository methods first. This ensures all migration targets exist.

### Step 2: Migrate in Layers
Complete one layer before moving to the next:
- Phase 2: Services (core business logic)
- Phase 3: Processing (background jobs)
- Phase 4: RAG (AI layer)
- Phase 5: Handlers (HTTP layer)

### Step 3: Incremental Verification
After each migration:
```bash
go build ./...
go test ./...  # or specific package
```

### Step 4: Final Cleanup
Only delete db files after:
- All migrations complete
- All tests pass
- Build succeeds

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Missing repository methods | Phase 1 adds all missing methods before migration |
| Test failures | Run tests after each file migration |
| Circular dependencies | Review import structure before migration |
| Build failures | Build after each phase |
| Coverage drop | Run `make cover-gate` before cleanup |

---

## Success Criteria

- [ ] No direct `db.*` calls in production code
- [ ] All data access goes through repository interfaces
- [ ] `go build ./...` succeeds
- [ ] `make test-all` passes
- [ ] `make cover-gate` passes (≥90% coverage)
- [ ] No `internal/db` package imports in production code
- [ ] All `internal/db/*` files deleted (except `db.go`)
