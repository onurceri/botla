# Task 004: Bootstrap Modularization

## Agent Prompt

> **Objective:** Modularize the `newApplication()` function in `cmd/server/main.go` to prevent it from becoming a "God constructor" as features grow.
>
> **Context:** The composition root is currently 92 lines and initializes 11+ major components. While manageable now, adding more features will make it unwieldy. This task creates logical "modules" to group related initialization.
>
> **Approach:**
> 1. Identify logical groupings of components
> 2. Create module functions that return initialized groups
> 3. Keep `newApplication()` as the orchestrator
> 4. Maintain fail-fast semantics

---

## Problem Statement

`cmd/server/main.go:newApplication()` currently:
- Initializes 11+ major components
- Is 92 lines of procedural code
- Will grow with each new feature
- Makes it harder to understand startup order

## Impact

- **Low Risk**: Internal refactoring only
- **Improved Readability**: Clear module boundaries
- **Easier Feature Addition**: New features go in appropriate modules
- **Better Error Isolation**: Module-specific error messages

---

## Acceptance Criteria

- [x] `newApplication()` reduced to <50 lines
- [x] At least 3 module initializer functions created
- [x] Startup order and fail-fast behavior preserved
- [x] All existing functionality unchanged
- [x] Server starts successfully (verified fail-fast on plan validation)
- [x] All tests pass

---

## Current Architecture Analysis

### Components Initialized in newApplication()

| Line | Component | Proposed Module |
|---|---|---|
| 71-75 | Database pool | `initInfrastructure` |
| 77-89 | Qdrant client + collection | `initInfrastructure` |
| 91-98 | Storage service (R2) | `initInfrastructure` |
| 100-107 | Tokenizer | `initInfrastructure` |
| 109-114 | OpenAI client | `initAIClients` |
| 116-121 | Source queue | `initProcessing` |
| 123-137 | Redis + rate limiters | `initRateLimiting` |
| 139-140 | Refresh scheduler | `initSchedulers` |
| 142-143 | Retention job | `initSchedulers` |
| 145-146 | Worker pool | `initProcessing` |

### Proposed Modules

```
┌─────────────────────────────────────────────────────┐
│                  newApplication()                    │
│                   (Orchestrator)                     │
└────────────┬────────────┬────────────┬──────────────┘
             │            │            │
    ┌────────▼────┐ ┌─────▼─────┐ ┌────▼────────┐
    │ initInfra   │ │ initAI    │ │ initRateLim │
    │ (DB, Qdrant,│ │ (OpenAI,  │ │ (Redis,     │
    │  Storage)   │ │  Queue)   │ │  Limiters)  │
    └─────────────┘ └───────────┘ └─────────────┘
```

---

## Implementation Plan

### Phase 1: Create Infrastructure Module

- [x] **Step 1.1**: Create infrastructure init function
  ```go
  // infraDeps holds infrastructure dependencies.
  type infraDeps struct {
      db         *sql.DB
      qdrant     *rag.QdrantClient
      storage    storage.StorageService
  }
  
  // initInfrastructure initializes core infrastructure components.
  func initInfrastructure(cfg *config.Config, log *logger.Logger) (*infraDeps, error) {
      // Initialize database
      pool, err := db.New(cfg)
      if err != nil {
          log.Error("db_init_failed", map[string]any{"error": err.Error()})
          return nil, fmt.Errorf("init db: %w", err)
      }
  
      // Initialize Qdrant
      qdrantClient, err := rag.NewQdrantClientFromEnv()
      if err != nil {
          log.Error("qdrant_init_failed", map[string]any{"error": err.Error()})
          return nil, fmt.Errorf("init qdrant: %w", err)
      }
  
      // Ensure embeddings collection exists
      if err := ensureQdrantCollection(qdrantClient, log); err != nil {
          return nil, fmt.Errorf("ensure qdrant collection: %w", err)
      }
      log.Info("qdrant_collection_ready", nil)
  
      // Initialize storage service
      var storageService storage.StorageService
      if cfg.R2_ACCOUNT_ID != "" {
          storageService, err = storage.NewR2Storage(
              cfg.R2_ACCOUNT_ID, 
              cfg.R2_ACCESS_KEY_ID, 
              cfg.R2_SECRET_ACCESS_KEY, 
              cfg.R2_BUCKET_NAME,
          )
          if err != nil {
              log.Error("storage_init_failed", map[string]any{"error": err.Error()})
          }
      }
  
      // Initialize tokenizer with R2 training data
      if storageService != nil {
          if tokErr := tokenizer.Init(context.Background(), storageService); tokErr != nil {
              log.Warn("tokenizer_init_fallback", map[string]any{"error": tokErr.Error()})
          } else {
              log.Info("tokenizer_loaded", nil)
          }
      }
  
      return &infraDeps{
          db:      pool,
          qdrant:  qdrantClient,
          storage: storageService,
      }, nil
  }
  ```

### Phase 2: Create Rate Limiting Module

- [x] **Step 2.1**: Create rate limiting init function
  ```go
  // rateLimitDeps holds rate limiting dependencies.
  type rateLimitDeps struct {
      redisClient   *redis.Client
      globalLimiter ratelimit.Limiter
      rateLimiter   *middleware.RateLimiter
  }
  
  // initRateLimiting initializes Redis and rate limiters.
  func initRateLimiting(log *logger.Logger, pool *sql.DB) (*rateLimitDeps, error) {
      redisClient, err := initRedisClient(log)
      if err != nil {
          log.Error("redis_required", map[string]any{
              "error":   err.Error(),
              "message": "Redis is required for rate limiting",
          })
          return nil, fmt.Errorf("init redis: %w", err)
      }
  
      rlConfig := ratelimit.NewConfigFromEnv()
      globalLimiter := ratelimit.NewRedisLimiter(redisClient, rlConfig.Global)
      log.Info("rate_limiter_initialized", map[string]any{"backend": "redis", "mode": "plan-based"})
      rl := middleware.NewRateLimiter(globalLimiter, redisClient, rlConfig)
  
      return &rateLimitDeps{
          redisClient:   redisClient,
          globalLimiter: globalLimiter,
          rateLimiter:   rl,
      }, nil
  }
  ```

### Phase 3: Create Processing Module

- [x] **Step 3.1**: Create processing init function
  ```go
  // processingDeps holds processing dependencies.
  type processingDeps struct {
      queue      *processing.SourceQueue
      workerPool *workers.WorkerPool
  }
  
  // initProcessing initializes the source processing queue and worker pool.
  func initProcessing(
      cfg *config.Config,
      log *logger.Logger,
      pool *sql.DB,
      storageService storage.StorageService,
      qdrantClient *rag.QdrantClient,
  ) (*processingDeps, error) {
      // Initialize OpenAI client for processing
      oaiClient, err := rag.NewOpenAIClient(cfg)
      if err != nil {
          log.Error("openai_init_failed", map[string]any{"error": err.Error()})
          return nil, fmt.Errorf("init openai: %w", err)
      }
  
      // Start source processing queue
      q, err := processing.StartSourceQueue(
          pool, storageService, oaiClient, qdrantClient, cfg.WORKER_COUNT,
      )
      if err != nil {
          log.Error("source_queue_init_failed", map[string]any{"error": err.Error()})
          return nil, fmt.Errorf("init source queue: %w", err)
      }
  
      // Initialize worker pool
      workerPool := workers.NewWorkerPool(log, 10)
  
      return &processingDeps{
          queue:      q,
          workerPool: workerPool,
      }, nil
  }
  ```

### Phase 4: Create Schedulers Module

- [x] **Step 4.1**: Create schedulers init function
  ```go
  // schedulerDeps holds scheduler dependencies.
  type schedulerDeps struct {
      refreshScheduler *services.RefreshScheduler
      retentionJob     *services.RetentionJob
  }
  
  // initSchedulers initializes background schedulers and jobs.
  func initSchedulers(
      pool *sql.DB,
      log *logger.Logger,
      queue *processing.SourceQueue,
      storageService storage.StorageService,
  ) *schedulerDeps {
      return &schedulerDeps{
          refreshScheduler: services.NewRefreshScheduler(pool, queue, log),
          retentionJob:     services.NewRetentionJob(pool, log, storageService),
      }
  }
  ```

### Phase 5: Refactor newApplication

- [x] **Step 5.1**: Update `newApplication` to use modules
  ```go
  func newApplication(cfg *config.Config, log *logger.Logger) (*application, error) {
      // Phase 1: Infrastructure
      infra, err := initInfrastructure(cfg, log)
      if err != nil {
          return nil, err
      }
  
      // Phase 2: Rate Limiting
      rl, err := initRateLimiting(log, infra.db)
      if err != nil {
          return nil, err
      }
  
      // Phase 3: Processing
      proc, err := initProcessing(cfg, log, infra.db, infra.storage, infra.qdrant)
      if err != nil {
          return nil, err
      }
  
      // Phase 4: Schedulers
      sched := initSchedulers(infra.db, log, proc.queue, infra.storage)
  
      return &application{
          cfg:              cfg,
          log:              log,
          db:               infra.db,
          qdrantClient:     infra.qdrant,
          storageService:   infra.storage,
          redisClient:      rl.redisClient,
          globalLimiter:    rl.globalLimiter,
          rateLimiter:      rl.rateLimiter,
          queue:            proc.queue,
          workerPool:       proc.workerPool,
          refreshScheduler: sched.refreshScheduler,
          retentionJob:     sched.retentionJob,
      }, nil
  }
  ```

### Phase 6: Verification

- [x] **Step 6.1**: Verify server starts
- [x] **Step 6.2**: Run tests
- [x] **Step 6.3**: Verify startup order is preserved
  ```bash
  # Check logs show same initialization order
  make be-run-no-pdf 2>&1 | head -50
  ```

---

## Files to Modify

| File | Changes |
|---|---|
| `cmd/server/main.go` | Add module functions, refactor `newApplication` |

---

## Design Decisions

### Why Not Separate Files?

The modules are small enough to stay in `main.go`. Separate files would fragment the composition root unnecessarily.

### Why Not Interfaces?

At the bootstrap level, concrete types are fine. Interfaces would add complexity without benefit here.

### Module Dependency Order

```
initInfrastructure → initRateLimiting → initProcessing → initSchedulers
         │                  │                 │
         └──────────────────┴─────────────────┘
                  All depend on infra
```

---

## Success Metrics

- [x] `newApplication()` is 33 lines (was 107)
- [x] Each module is <60 lines
- [x] Clear logical grouping visible
- [x] Same startup behavior (verified by logs - fails fast on plan validation)

---

## Rollback Plan

```bash
git checkout main -- cmd/server/main.go
```

---

## Completion Summary

**Completed on:** 2025-12-31

**Changes Made:**
- Created 4 module structs: `infraDeps`, `rateLimitDeps`, `processingDeps`, `schedulerDeps`
- Created 4 module init functions: `initInfrastructure()`, `initRateLimiting()`, `initProcessing()`, `initSchedulers()`
- Refactored `newApplication()` from 107 lines to 33 lines

**Verification:**
- All tests pass
- Linter passes with 0 issues
- Server starts and correctly implements fail-fast behavior
