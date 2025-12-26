# Issue 003: Background Task Reliability and Panic Safety

## Priority: Medium
## Confidence: High

## Summary

The `FeedbackHandler` spawns background goroutines that lack proper lifecycle management, use `fmt.Printf` instead of structured logging, and may be terminated abruptly during server shutdown.

## Evidence

**File:** [chat.go](file:///Users/onur/Documents/workspace/botla-co/internal/api/handlers/chat.go#L157-L168)

```go
// Update Analytics
go func() {
    // CR-002: Recover from panics to prevent server crash
    defer func() {
        if r := recover(); r != nil {
            // Log panic for debugging - use fmt since we don't have logger access
            fmt.Printf("feedback_analytics_panic: %v\n", r)
        }
    }()
    bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = db.IncrementFeedback(bgCtx, h.DB, chatbotID, time.Now(), oldThumbsUp, req.ThumbsUp)
}()
```

## Problems Identified

### 1. Unstructured Logging
- Uses `fmt.Printf` which is not captured by structured log aggregators (CloudWatch, ELK, Datadog)
- Panics in production will be invisible in centralized logging
- No correlation IDs or context for debugging

### 2. No Graceful Shutdown
- Goroutines are not tracked by a `WaitGroup` or worker pool
- During server shutdown, these goroutines may be forcefully terminated
- **Result**: Lost analytics data, inconsistent feedback counts

### 3. Silent Error Swallowing
- The error from `db.IncrementFeedback` is discarded with `_ =`
- Database failures go unnoticed
- No retry mechanism for transient failures

### 4. Resource Exhaustion Risk
- No limit on concurrent background goroutines
- Under high load, thousands of goroutines could be spawned
- Potential memory exhaustion and scheduler overhead

## Failure Scenarios

1. **Log Aggregator Blindness**: Production panic goes unnoticed because `fmt.Printf` output isn't captured
2. **Data Loss on Shutdown**: 50 in-flight feedback updates lost when server receives SIGTERM
3. **Silent DB Failure**: Database connection drops, all feedback updates fail silently
4. **Goroutine Explosion**: Feedback endpoint hit 1000 times/second creates 1000 concurrent goroutines

## Recommended Fix

### Create a Background Worker Pool

```go
// internal/workers/pool.go
type WorkerPool struct {
    wg      sync.WaitGroup
    jobs    chan func(context.Context)
    ctx     context.Context
    cancel  context.CancelFunc
    logger  *logger.Logger
    maxSize int
}

func NewWorkerPool(logger *logger.Logger, size int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    pool := &WorkerPool{
        jobs:    make(chan func(context.Context), size),
        ctx:     ctx,
        cancel:  cancel,
        logger:  logger,
        maxSize: size,
    }
    
    for i := 0; i < size; i++ {
        pool.wg.Add(1)
        go pool.worker()
    }
    return pool
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()
    for {
        select {
        case <-p.ctx.Done():
            return
        case job := <-p.jobs:
            p.executeJob(job)
        }
    }
}

func (p *WorkerPool) executeJob(job func(context.Context)) {
    defer func() {
        if r := recover(); r != nil {
            p.logger.Error("worker_panic", map[string]any{
                "panic": fmt.Sprintf("%v", r),
                "stack": string(debug.Stack()),
            })
        }
    }()
    
    ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
    defer cancel()
    job(ctx)
}

func (p *WorkerPool) Submit(job func(context.Context)) bool {
    select {
    case p.jobs <- job:
        return true
    default:
        p.logger.Warn("worker_pool_full", nil)
        return false
    }
}

func (p *WorkerPool) Shutdown(timeout time.Duration) {
    p.cancel()
    
    done := make(chan struct{})
    go func() {
        p.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
    case <-time.After(timeout):
        p.logger.Warn("worker_pool_shutdown_timeout", nil)
    }
}
```

### Update Handler to Use Worker Pool

```go
type ChatHandlers struct {
    DB               *sql.DB
    ChatService      *services.ChatService
    WorkspaceService *services.WorkspaceService
    OrgService       *services.OrganizationService
    WorkerPool       *workers.WorkerPool  // Add this
    Logger           *logger.Logger       // Add this
}

func (h *ChatHandlers) FeedbackHandler(w http.ResponseWriter, r *http.Request) {
    // ... existing validation code ...
    
    h.WorkerPool.Submit(func(ctx context.Context) {
        if err := db.IncrementFeedback(ctx, h.DB, chatbotID, time.Now(), oldThumbsUp, req.ThumbsUp); err != nil {
            h.Logger.Error("feedback_increment_failed", map[string]any{
                "chatbot_id": chatbotID,
                "error":      err.Error(),
            })
        }
    })
    
    w.WriteHeader(http.StatusOK)
}
```

### Wire Up in main.go

```go
// Start worker pool
workerPool := workers.NewWorkerPool(log, 10) // 10 workers

// During shutdown
workerPool.Shutdown(10 * time.Second)
```

## Verification

1. Unit test: Worker pool executes jobs and handles panics
2. Unit test: Worker pool drains jobs during graceful shutdown
3. Integration test: Verify feedback analytics are logged with structured logger
4. Load test: Verify bounded goroutine count under high load

## Related Files

- `cmd/server/main.go` - Application startup/shutdown
- `internal/api/handlers/chat.go` - Handler using background work
