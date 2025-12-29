# Backend Task 006: URL Processor Transaction Safety

## Background
In `internal/processing/url_processor.go`, the increment of successful ingestion count and the addition of embedding tokens happen sequentially but independently. If the second operation fails, the data becomes inconsistent.

**File:** `internal/processing/url_processor.go`
**Location:** Lines 261-262

## Implementation Summary

### 1. DB Methods (internal/db/usage_ingestions.go)
Added two new transactional functions:
- `IncrementSuccessfulIngestionTx(ctx, tx, userID, at, delta)` - Increments sources_count within a transaction
- `AddEmbeddingTokensTx(ctx, tx, userID, at, tokens)` - Adds embedding_tokens within a transaction

### 2. Transaction Safety (internal/processing/url_processor.go)
Modified `ProcessWithSteps` (lines 262-294) to wrap statistics updates in a transaction:
- Begins a transaction before the operations
- Uses `defer` to ensure rollback on error
- Calls the new transactional DB functions
- Commits only if both operations succeed
- Logs warnings for any failures

### 3. Tests

**Unit Tests (internal/processing/url_processor_test.go):**
- `TestURLProcessor_TransactionSafety` - Tests transaction atomicity
- `TestURLProcessor_TransactionWithMockedDB` - Tests successful transaction flow
- `TestURLProcessor_TransactionRollback` - Tests rollback on failure
- `TestURLProcessor_ConcurrentTransactions` - Tests concurrent transaction safety
- `TestURLProcessor_NewURLProcessor` - Tests processor creation

**DB Tests (internal/db/usage_ingestions_test.go):**
- `TestUsageIngestions_Tx_Success` - Tests successful transaction
- `TestUsageIngestions_Tx_RollbackOnFailure` - Tests rollback on failure
- `TestUsageIngestions_Tx_Atomicity` - Tests atomicity guarantee
- `TestUsageIngestions_Tx_Consecutive` - Tests multiple consecutive transactions

## Checklist
- [x] Create/Update DB methods to support transactions
- [x] Wrap statistics updates in `url_processor.go` within a transaction
- [x] Ensure proper Commit/Rollback handling
- [x] Verify tests validation
