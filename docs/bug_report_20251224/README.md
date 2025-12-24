# Backend Bug Report - December 2024

This document catalogs potential bugs and code quality issues found in the Botla backend codebase (`internal/`).

## Report Contents

- [Critical Issues](#critical-issues)
- [High Severity Issues](#high-severity-issues)
- [Medium Severity Issues](#medium-severity-issues)
- [Low Severity Issues](#low-severity-issues)
- [Database Layer Issues](#database-layer-issues)
- [Recommended Fixes](#recommended-fixes)

---

## Critical Issues

### CR-001: Potential Panic - Empty Choices Array in Agentic Loop

**File:** `internal/services/chat_pipeline.go`
**Lines:** 211-221

```go
choice := response.Choices[0]
```

**Problem:** If the LLM returns an empty `Choices` array, this code will panic with `index out of range`.

**Impact:** Process crash, loss of user session data.

**Trigger Condition:** LLM API returns a valid response but with no choices (rare but possible with certain model errors or malformed responses).

**Recommendation:**
```go
if len(response.Choices) == 0 {
    return fmt.Errorf("LLM returned empty choices")
}
choice := response.Choices[0]
```

---

### CR-002: Goroutine Without Panic Recovery in Feedback Handler

**File:** `internal/api/handlers/chat.go`
**Lines:** 160-164

```go
go func() {
    bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = db.IncrementFeedback(bgCtx, h.DB, chatbotID, time.Now(), oldThumbsUp, req.ThumbsUp)
}()
```

**Problem:** Goroutine spawned without panic recovery. If `db.IncrementFeedback` panics, the entire process crashes.

**Impact:** Single feedback request can crash the entire server.

**Recommendation:** Add recover() in the goroutine or make the call synchronous.

---

### CR-003: Insecure Origin Check - Domain Bypass

**File:** `internal/api/handlers/public.go`
**Lines:** 200-216

```go
if strings.Contains(origin, d) {
    allowed = true
    break
}
```

**Problem:** Using `strings.Contains` for origin validation allows bypass. Origin `example.com.evil.com` would match allowed domain `example.com`.

**Impact:** Unauthorized access to chatbot widgets, potential data leakage.

**Recommendation:** Parse the origin URL and compare against registered domains properly:
```go
parsed, err := url.Parse(origin)
if err == nil {
    hostname := parsed.Hostname()
    if hostname == d || strings.HasSuffix(hostname, "."+d) {
        allowed = true
    }
}
```

---

### CR-004: SQL Injection Risk in Analytics Query

**File:** `internal/db/analytics.go`
**Lines:** 221-236

```go
whereClause := "WHERE chatbot_id = $1"
if includeAll {
    whereClause = ""
}
query := "SELECT ... " + whereClause + " ORDER BY created_at DESC"
```

**Problem:** While current inputs appear controlled, dynamic query construction via string concatenation is dangerous.

**Impact:** Potential SQL injection if logic changes or inputs are not properly validated.

**Recommendation:** Use parameterized queries with proper placeholder handling for PostgreSQL.

---

## High Severity Issues

### HI-001: Missing Timeout on LLM Tool Name Generation

**File:** `internal/api/handlers/action.go`
**Lines:** 80-84, 182-187

```go
toolName, err := h.ToolNameGenerator.Generate()
```

**Problem:** Tool name generation is called without context timeout. Can hang indefinitely.

**Impact:** Handler thread exhaustion, denial of service.

**Recommendation:** Add timeout context:
```go
ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
defer cancel()
toolName, err := h.ToolNameGenerator.Generate(ctx)
```

---

### HI-002: Goroutine Without Recovery in Source Deletion

**File:** `internal/api/handlers/source_single.go`
**Line:** 74

```go
go processing.ReAggregateSuggestionsForChatbot(context.Background(), h.DB, s.ChatbotID, h.Log)
```

**Problem:** Goroutine without panic recovery after critical operation.

**Impact:** Process crash during source deletion cleanup.

**Recommendation:** Add recover() or run synchronously.

---

### HI-003: Silent Failure on Vector Cleanup

**File:** `internal/api/handlers/chatbot.go`
**Lines:** 189-191

```go
for _, sid := range sourceIDs {
    _ = h.VectorStore.DeleteBySourceID(r.Context(), sid)
}
```

**Problem:** Errors from vector store deletion are silently ignored.

**Impact:** Orphaned vectors in Qdrant, wasted storage, potential search quality issues.

**Recommendation:** Log errors and implement retry logic.

---

### HI-004: Nil Return Instead of Error

**File:** `internal/services/chat_pipeline.go`
**Lines:** 284-287

```go
messageID, err := db.CreateMessage(ctx, s.DB, msg)
if err != nil {
    return ""
}
```

**Problem:** Returns empty string instead of propagating the error.

**Impact:** Caller cannot distinguish between success with empty ID and failure.

---

## Medium Severity Issues

### MI-001: Race Condition in Refresh Scheduler

**File:** `internal/services/refresh_scheduler.go`
**Lines:** 70-82

```go
func (s *Scheduler) Start() {
    s.stopChan = make(chan struct{})  // NEW channel created
    // ...
}

func (s *Scheduler) Stop() {
    close(s.stopChan)  // Closes OLD channel if Start() was called twice
    s.wg.Wait()
}
```

**Problem:** If `Stop()` is called and then `Start()` is called again, the new `stopChan` is disconnected from any running goroutine.

**Impact:** `Stop()` may hang waiting for goroutines that are listening on a different channel.

---

### MI-002: No Panic Recovery in Scheduler Run Loop

**File:** `internal/services/refresh_scheduler.go`
**Lines:** 82-98

**Problem:** The scheduler's main run loop has no `recover()`. If `processDueChatbots` panics, the scheduler dies silently.

**Impact:** Chatbot refreshes stop working until service restart.

---

### MI-003: Transaction Missing in DeleteWorkspace

**File:** `internal/services/chatbot_service.go`
**Lines:** 68-89

Multiple sequential DB operations without transaction:
```go
// Delete sources
// Delete source analytics
// Delete messages
// Delete conversations
// Delete workspace
```

**Problem:** If any step fails mid-operation, data becomes inconsistent.

**Impact:** Orphaned records, broken chatbot state.

**Recommendation:** Wrap in transaction with proper rollback on error.

---

### MI-004: String-Based Error Checking (Fragile)

**File:** `internal/api/handlers/organization.go`
**Lines:** 68-71, 118-121, 136-140, 238-243, 276-280

```go
if strings.Contains(err.Error(), "exists") {
    // handle duplicate
}
```

**Problem:** Depends on exact error message text from database.

**Impact:** Breaks if error messages change or are localized.

**Recommendation:** Use specific error codes or structured error types.

---

### MI-005: JSON Marshal Errors Silently Ignored

**Files:**
- `internal/db/chatbot.go` (lines 31-46, 166-215)
- `internal/db/chatbot_refresh.go` (lines 61-100)

```go
_ = json.Marshal(data)
```

**Problem:** JSON marshaling errors are discarded.

**Impact:** Data corruption, fields saved as `null` unexpectedly.

---

### MI-006: Rows.Err() Not Checked

**Files:**
- `internal/services/workspace_service.go` (line 111)
- `internal/api/handlers/me.go` (lines 84-91)

```go
for rows.Next() {
    // scan
}
return orgs, nil  // No rows.Err() check
```

**Problem:** Iteration errors are not checked.

**Impact:** Partial results returned without error indication.

---

### MI-007: Error Swallowing in RAG Search

**File:** `internal/services/chat_service.go`
**Line:** 90

```go
s.performRAGSearch(ctx, cc)
```

**Problem:** Error is not returned, function doesn't even return an error type.

**Impact:** RAG failures silently fall back to basic mode.

---

### MI-008: Fallback to "Free" Plan on Any Database Error

**File:** `internal/api/handlers/plan.go`
**Lines:** 159-161

```go
if err != nil {
    planCode = "free"
}
```

**Problem:** Any database error causes fallback to free plan.

**Impact:** Users might unexpectedly lose paid features due to transient DB issues.

---

## Low Severity Issues

### LI-001: Goroutine Leak in Handoff Service

**File:** `internal/services/handoff_service.go`
**Lines:** 168-177

```go
go func() {
    bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := db.IncrementAnalytics(bgCtx, s.DB, bot.ID, time.Now(), false, 0, true, 0); err != nil {
        // only logged
    }
}()
```

**Problem:** Analytics increment is fire-and-forget.

---

### LI-002: Error Swallowing in Email Sending

**File:** `internal/services/handoff_service.go`
**Line:** 183

```go
return nil
```

**Problem:** Email sending returns `nil` even though feature is not implemented.

---

### LI-003: RowsAffected Error Ignored

**Files:**
- `internal/db/message_sources.go` (line 75)
- `internal/db/pending_url.go` (lines 75, 92, 106, 120)

```go
affected, _ := result.RowsAffected()
```

---

### LI-004: Unreachable ErrNoRows Check

**File:** `internal/db/source.go`
**Lines:** 306-308

```go
err := row.Scan(&count)
if err == sql.ErrNoRows {  // Never reached - Scan returns error directly
```

---

### LI-005: Duplicate Logic for Refresh Time Calculation

**File:** `internal/services/chatbot_service.go`
**Lines:** 489-513

`calculateNextRefreshTime` duplicates logic from `refresh_scheduler.go`.

---

### LI-006: New Service Created Per Request

**File:** `internal/api/handlers/handoff.go`
**Lines:** 87-88, 112, 147

```go
services.NewHandoffService(h.DB, h.Log)
```

**Problem:** Service created for every request instead of reusing.

---

### LI-007: Hardcoded Multipart Form Size Limit

**File:** `internal/api/handlers/source_create.go`
**Line:** 35

```go
r.ParseMultipartForm(52 << 20)  // 52MB hardcoded
```

---

### LI-008: Queue.Enqueue Errors Ignored

**Files:**
- `internal/api/handlers/source_create.go` (lines 248-250)
- `internal/api/handlers/source_bulk.go` (lines 125-127)

---

## Database Layer Issues

### DB-001: Silent JSON Unmarshal Errors

**File:** `internal/db/chatbot_refresh.go`
**Lines:** 61-100

```go
_ = json.Unmarshal(sj, &arr)
```

---

### DB-002: Legacy Fallback Masks Errors

**File:** `internal/db/source.go`
**Lines:** 23-38

The catch-all fallback catches ALL errors, not just `undefined_column`.

---

### DB-003: Transaction Begin Error Not Checked

**File:** `internal/db/conversation.go`
**Lines:** 102-104

If `BeginTx` fails, `tx` is nil and `tx.Rollback()` panics.

---

### DB-004: Silent JSON Scan Errors

**File:** `internal/db/user.go`
**Lines:** 28, 53

```go
if err := data.Scan(onboardingDataJSON); err == nil {
```

---

## Recommended Fixes Summary

### Priority 1 (Critical - This Sprint)

1. Add bounds check before `response.Choices[0]` access
2. Add panic recovery in all goroutines
3. Fix origin validation to use proper URL parsing
4. Review and fix SQL injection risk in analytics.go

### Priority 2 (High - This Month)

5. Add timeouts to all external calls
6. Implement transactions for multi-step deletes
7. Fix nil return vs error issues
8. Log all ignored errors properly

### Priority 3 (Medium - Next Sprint)

9. Replace string-based error checking with error codes
10. Add rows.Err() checks after all iterations
11. Implement proper context propagation
12. Create reusable service instances

---

## Testing Recommendations

1. **Unit Tests:** Add tests for empty choices, nil database results
2. **Integration Tests:** Test concurrent scheduler start/stop
3. **Fuzz Tests:** Test handler input validation
4. **Load Tests:** Test goroutine behavior under load

---

*Report generated: December 2024*
*Coverage of: internal/api/handlers/, internal/services/, internal/db/*
