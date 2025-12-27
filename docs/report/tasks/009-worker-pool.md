# Task 009: Worker Pool for Parallel Processing

**Priority:** 🟡 Medium (Performance)  
**Phase:** 2 - Async Training Improvements  
**Estimated Time:** 4-5 hours  
**Dependencies:** Task 004 (Integrate Job Tracking)  

---

## Problem Statement

Current source queue has a single worker processing jobs sequentially. This is slow for batch imports and doesn't scale.

---

## Objective

Implement a configurable worker pool with multiple concurrent workers, graceful shutdown, and health monitoring.

---

## Implementation Details

### Step 1: Add Configuration

**File:** `pkg/config/config.go` (MODIFY)

```go
type Config struct {
    WorkerCount int `env:"WORKER_COUNT" envDefault:"4"`
}
```

### Step 2: Modify Source Queue

**File:** `internal/processing/sources_queue.go` (MODIFY)

Key changes:
- Add `workerCount` field to struct
- Start multiple workers in `StartSourceQueue`
- Add worker ID to logging
- Implement `WorkerCount()` and `QueueLength()` methods

```go
// Start worker pool
for i := 0; i < workerCount; i++ {
    q.wg.Add(1)
    go q.worker(i)
}

func (q *SourceQueue) worker(workerID int) {
    defer q.wg.Done()
    for {
        select {
        case <-q.stopCh:
            return
        case jobID := <-q.ch:
            q.processJob(jobID, workerID)
        }
    }
}
```

### Step 3: Update Server Initialization

**File:** `cmd/server/main.go` (MODIFY)

Pass `WorkerCount` from config to queue initialization.

### Step 4: Add Queue Health to Health Endpoint

**File:** `internal/api/handlers/health.go` (MODIFY)

Add queue worker count and length to health response.

---

## Tests to Write

### Unit Tests: `internal/processing/worker_pool_test.go`

- `TestWorkerPool_MultipleWorkers` - Verify correct worker count
- `TestWorkerPool_ParallelProcessing` - Verify parallel execution
- `TestWorkerPool_GracefulShutdown` - Verify shutdown waits for workers
- `TestWorkerPool_MaxWorkerLimit` - Verify cap at 16 workers

### Integration Test: `internal/integration/worker_pool_test.go`

- Add multiple sources at once, verify parallel processing

---

## Acceptance Criteria

- [x] Configurable worker count via WORKER_COUNT
- [x] Workers process jobs in parallel
- [x] Graceful shutdown waits for in-flight jobs
- [x] Worker count capped at 16
- [x] Queue health visible in health endpoint
- [x] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `pkg/config/config.go` | MODIFY |
| `internal/processing/sources_queue.go` | MODIFY |
| `cmd/server/main.go` | MODIFY |
| `internal/api/handlers/health.go` | MODIFY |
| `internal/processing/worker_pool_test.go` | CREATE |
| `internal/integration/worker_pool_test.go` | CREATE |
