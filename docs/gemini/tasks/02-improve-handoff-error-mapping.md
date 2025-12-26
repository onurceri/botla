# Task 02: Improve Handoff Error Mapping

**Priority:** 🟡 Medium  
**Effort:** Low (1-2 hours)  
**Risk Level:** Low (debugging improvement)

---

## Problem Statement

The handoff API handlers mask specific service errors, returning generic error responses instead of meaningful error codes. This makes debugging production issues difficult because client-side errors don't match server-side root causes.

### Evidence

From `internal/integration/handoff_test.go` (line 667):
```go
// but currently handler masks errors unless we update it too.
```

### Why This Matters

1. **Debugging Difficulty**: Generic errors hide root causes
2. **Poor API Semantics**: Clients can't differentiate between error types (duplicate, not found, rate limited)
3. **Support Overhead**: More time spent diagnosing issues that could be self-evident

---

## Acceptance Criteria

- [ ] Handoff service errors are mapped to appropriate HTTP status codes
- [ ] Error responses include specific error codes (e.g., `handoff_exists`, `handoff_not_found`)
- [ ] Error messages are localized where applicable
- [ ] All existing tests pass
- [ ] New tests validate specific error responses

---

## Implementation Steps

### Step 1: Define Service Error Types

Ensure `internal/services/errors.go` has specific error types for handoff:

```go
package services

import "errors"

var (
    // Handoff errors
    ErrHandoffExists     = errors.New("handoff request already exists for this session")
    ErrHandoffNotFound   = errors.New("handoff request not found")
    ErrHandoffExpired    = errors.New("handoff request has expired")
    ErrHandoffClosed     = errors.New("handoff request is already closed")
    ErrHandoffRateLimited = errors.New("too many handoff requests")
)
```

### Step 2: Create Error Mapping Utility

Create a centralized error-to-HTTP mapping in `internal/api/errors.go`:

```go
package api

import (
    "errors"
    "net/http"
    
    "github.com/onurceri/botla-co/internal/services"
)

// HandoffErrorResponse maps handoff service errors to HTTP responses
type ErrorMapping struct {
    StatusCode int
    ErrorCode  string
    MessageKey string // For localization
}

var handoffErrorMappings = map[error]ErrorMapping{
    services.ErrHandoffExists: {
        StatusCode: http.StatusConflict,
        ErrorCode:  "handoff_exists",
        MessageKey: "error.handoff.exists",
    },
    services.ErrHandoffNotFound: {
        StatusCode: http.StatusNotFound,
        ErrorCode:  "handoff_not_found",
        MessageKey: "error.handoff.not_found",
    },
    services.ErrHandoffExpired: {
        StatusCode: http.StatusGone,
        ErrorCode:  "handoff_expired",
        MessageKey: "error.handoff.expired",
    },
    services.ErrHandoffClosed: {
        StatusCode: http.StatusConflict,
        ErrorCode:  "handoff_closed",
        MessageKey: "error.handoff.closed",
    },
    services.ErrHandoffRateLimited: {
        StatusCode: http.StatusTooManyRequests,
        ErrorCode:  "handoff_rate_limited",
        MessageKey: "error.handoff.rate_limited",
    },
}

// MapHandoffError maps a service error to an HTTP response
func MapHandoffError(err error) (int, string, bool) {
    for target, mapping := range handoffErrorMappings {
        if errors.Is(err, target) {
            return mapping.StatusCode, mapping.ErrorCode, true
        }
    }
    return http.StatusInternalServerError, "internal_error", false
}
```

### Step 3: Update Handoff Handlers

Modify `internal/api/handlers/handoff.go` to use the error mapping:

**Before:**
```go
func (h *HandoffHandler) CreateHandoff(w http.ResponseWriter, r *http.Request) {
    // ...
    result, err := h.service.CreateHandoff(ctx, req)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError) // Generic!
        return
    }
    // ...
}
```

**After:**
```go
func (h *HandoffHandler) CreateHandoff(w http.ResponseWriter, r *http.Request) {
    // ...
    result, err := h.service.CreateHandoff(ctx, req)
    if err != nil {
        statusCode, errorCode, mapped := api.MapHandoffError(err)
        if mapped {
            api.WriteLocalizedError(w, statusCode, errorCode, langConfig)
        } else {
            // Log unexpected error
            h.log.Error("handoff_create_failed", map[string]any{"error": err.Error()})
            api.WriteError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
        }
        return
    }
    // ...
}
```

### Step 4: Update Service to Return Typed Errors

Ensure `internal/services/handoff_service.go` returns the defined errors:

```go
func (s *HandoffService) CreateHandoff(ctx context.Context, req CreateHandoffRequest) (*HandoffResult, error) {
    // Check for existing handoff
    existing, err := s.db.GetActiveHandoffBySession(ctx, req.SessionID)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("failed to check existing handoff: %w", err)
    }
    if existing != nil {
        return nil, ErrHandoffExists // Typed error!
    }
    
    // ... rest of logic
}
```

### Step 5: Add Localized Messages

Add error messages to the language files:

**File:** `data/sentences/english.json` (or equivalent)
```json
{
  "error.handoff.exists": "A support request already exists for this conversation",
  "error.handoff.not_found": "Support request not found",
  "error.handoff.expired": "This support request has expired",
  "error.handoff.closed": "This support request has already been closed",
  "error.handoff.rate_limited": "Too many support requests. Please try again later"
}
```

### Step 6: Update Tests

Add tests for specific error responses:

```go
func TestCreateHandoff_DuplicateReturns409(t *testing.T) {
    // Setup: Create an existing handoff for the session
    // ...
    
    // Act: Try to create another handoff for same session
    req := httptest.NewRequest("POST", "/api/v1/handoffs", body)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    // Assert: Should return 409 Conflict with specific error code
    assert.Equal(t, http.StatusConflict, w.Code)
    
    var resp map[string]any
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.Equal(t, "handoff_exists", resp["code"])
}
```

---

## Testing Checklist

- [ ] `go build ./...` succeeds
- [ ] `make test-no-pdf` passes
- [ ] `make lint` passes
- [ ] Test: Duplicate handoff returns 409 with `handoff_exists`
- [ ] Test: Non-existent handoff returns 404 with `handoff_not_found`
- [ ] Test: Unexpected errors still return 500 with logging

---

## Files to Modify

| File | Change |
|------|--------|
| `internal/services/errors.go` | Define typed errors for handoff service |
| `internal/api/errors.go` | Create error mapping utility |
| `internal/api/handlers/handoff.go` | Use error mapping in handlers |
| `internal/services/handoff_service.go` | Return typed errors |
| `internal/integration/handoff_test.go` | Add tests for specific error codes |

---

## Error Code Reference

| Service Error | HTTP Status | Error Code | When |
|---------------|-------------|------------|------|
| `ErrHandoffExists` | 409 Conflict | `handoff_exists` | Duplicate handoff for session |
| `ErrHandoffNotFound` | 404 Not Found | `handoff_not_found` | Handoff ID doesn't exist |
| `ErrHandoffExpired` | 410 Gone | `handoff_expired` | Handoff TTL exceeded |
| `ErrHandoffClosed` | 409 Conflict | `handoff_closed` | Handoff already resolved |
| `ErrHandoffRateLimited` | 429 Too Many | `handoff_rate_limited` | Rate limit exceeded |

---

## Related Issues

- Code Audit Finding #6: "Masked Errors in API Handlers"
- Integration test TODO in `handoff_test.go`
