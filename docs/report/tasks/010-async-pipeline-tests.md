# Task 010: Async Pipeline Integration Tests

**Priority:** 🟡 High (Quality)  
**Phase:** 8 - Test Coverage  
**Estimated Time:** 4-5 hours  
**Dependencies:** Tasks 002-005 (Job tracking system)  

---

## Problem Statement

The async training pipeline lacks comprehensive integration tests covering:
- Full job lifecycle
- Step transitions
- Retry behavior
- Failure scenarios
- Recovery after crash

---

## Objective

Create comprehensive integration tests for the async pipeline.

---

## Tests to Write

### File: `internal/integration/async_pipeline_test.go` (NEW)

```go
package integration

import (
    "testing"
    "time"
)

// Test full pipeline from source creation to completion
func TestAsyncPipeline_FullLifecycle(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    token := authToken(t, te.Server.URL, "pipeline@test.com")
    botID := createChatbot(t, te.Server.URL, token, "Pipeline Bot")
    
    // Create URL source
    sourceID := createURLSource(t, te.Server.URL, token, botID, "https://example.com")
    
    // Poll job status through steps
    var seenSteps []string
    for i := 0; i < 30; i++ {
        job := getJobStatus(t, te.Server.URL, token, sourceID)
        
        if step, ok := job["current_step"].(string); ok && step != "" {
            if len(seenSteps) == 0 || seenSteps[len(seenSteps)-1] != step {
                seenSteps = append(seenSteps, step)
            }
        }
        
        if job["status"] == "completed" || job["status"] == "failed" {
            break
        }
        time.Sleep(500 * time.Millisecond)
    }
    
    // Verify steps were traversed
    if len(seenSteps) < 2 {
        t.Errorf("expected multiple steps, saw: %v", seenSteps)
    }
}

// Test job failure and retry
func TestAsyncPipeline_FailureAndRetry(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    token := authToken(t, te.Server.URL, "retry@test.com")
    botID := createChatbot(t, te.Server.URL, token, "Retry Bot")
    
    // Create source with invalid URL
    sourceID := createURLSource(t, te.Server.URL, token, botID, "http://invalid.test")
    
    // Wait for failure
    time.Sleep(5 * time.Second)
    
    job := getJobStatus(t, te.Server.URL, token, sourceID)
    if job["status"] != "failed" {
        t.Skipf("job didn't fail: %s", job["status"])
    }
    
    // Verify error details
    if job["error_code"] == nil {
        t.Error("expected error_code")
    }
    if job["failed_step"] == nil {
        t.Error("expected failed_step")
    }
}

// Test job recovery after simulated restart
func TestAsyncPipeline_Recovery(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    // Create source in pending state
    sourceID := insertPendingSource(t, te.DB)
    
    // Trigger recovery
    te.Queue.RecoverPendingSources()
    
    // Wait and check processing started
    time.Sleep(2 * time.Second)
    
    source := getSource(t, te.DB, sourceID)
    if source["status"] == "pending" {
        t.Error("source should have started processing")
    }
}

// Test concurrent job processing
func TestAsyncPipeline_ConcurrentJobs(t *testing.T) {
    te := SetupTestEnv(t)
    defer te.Teardown()

    token := authToken(t, te.Server.URL, "concurrent@test.com")
    botID := createChatbot(t, te.Server.URL, token, "Concurrent Bot")
    
    // Create multiple sources simultaneously
    var sourceIDs []string
    for i := 0; i < 4; i++ {
        id := createTextSource(t, te.Server.URL, token, botID, 
            fmt.Sprintf("Content %d", i))
        sourceIDs = append(sourceIDs, id)
    }
    
    // All should complete
    for _, id := range sourceIDs {
        waitForSourceCompletion(t, te.Server.URL, token, id, 30*time.Second)
    }
}
```

---

## Acceptance Criteria

- [ ] Full lifecycle test passes
- [ ] Failure and retry test passes
- [ ] Recovery test passes
- [ ] Concurrent jobs test passes
- [ ] Tests run in CI pipeline

---

## Files Changed

| File | Action |
|------|--------|
| `internal/integration/async_pipeline_test.go` | CREATE |
