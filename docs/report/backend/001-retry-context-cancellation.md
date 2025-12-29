# Backend Task 001: Implement Context Cancellation Check in Retry Logic

## Background
The current retry logic in `internal/rag/openai.go` sleeps between retries without checking if the context has been cancelled. This can cause the application to hang unnecessarily during shutdown or client timeouts.

**File:** `internal/rag/openai.go`
**Location:** Lines 90-121 (approx)

## Integration Plan
1.  **Analyze `CreateEmbedding` and `CreateCompletion` methods**
    - Locate the retry loop `for attempt := 0; attempt < 4; attempt++`.
    - Locate the `time.Sleep` call.

2.  **Implement Context-Aware Sleep**
    - Replace `time.Sleep(...)` with a `select` statement that listens to `ctx.Done()`.
    - If `ctx.Done()` receives, return `nil, ctx.Err()` immediately.

3.  **Verify**
    - Create a unit test where the context is cancelled immediately.
    - Ensure the function returns immediately with context canceled error instead of waiting.

## Checklist
- [x] Locate retry loops in `internal/rag/openai.go`
- [x] Replace `time.Sleep` with `select { case <-ctx.Done(): ... case <-time.After(...): ... }`
- [x] Add unit test for context cancellation during retry
- [x] Run tests to ensure no regressions
