# Backend Task 006: URL Processor Transaction Safety

## Background
In `internal/processing/url_processor.go`, the increment of successful ingestion count and the addition of embedding tokens happen sequentially but independently. If the second operation fails, the data becomes inconsistent.

**File:** `internal/processing/url_processor.go`
**Location:** Lines 261-262

## Integration Plan
1.  **Refactor DB Methods**
    - Ensure `IncrementSuccessfulIngestion` and `AddEmbeddingTokens` can accept a `*sql.Tx` (or share a common transactional interface/context).
    - If using `sqlc`, use the `WithTx` pattern.

2.  **Implement Transaction**
    - In `ProcessWithSteps`, wrap these two calls in a `tx, err := p.DB.BeginTx(...)`.
    - Commit if both succeed, Rollback if either fails.

3.  **Verify**
    - Run integration tests for source processing.

## Checklist
- [ ] Create/Update DB methods to support transactions
- [ ] Wrap statistics updates in `url_processor.go` within a transaction
- [ ] Ensure proper Commit/Rollback handling
- [ ] Verify tests validation
