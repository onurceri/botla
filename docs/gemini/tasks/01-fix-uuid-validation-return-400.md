# Task 01: Fix UUID Validation to Return 400 Instead of 500

**Priority:** 🟡 Medium  
**Effort:** Low (1 hour)  
**Risk Level:** Medium (security noise, poor DX)

---

## Problem Statement

When malformed UUIDs are passed to API endpoints (e.g., `/api/v1/chatbots/invalid-uuid`), the system returns **500 Internal Server Error** instead of **400 Bad Request**. This occurs because UUID parsing errors bubble up as unhandled exceptions.

### Evidence

From `internal/integration/rbac_middleware_test.go` (line 209):
```go
// Note: Currently returns 500 (Internal Server Error) due to unhandled UUID parsing error.
// This should return 400 (Bad Request) for malformed IDs - this should be fixed.
```

### Why This Matters

1. **Security Noise**: 500 errors trigger alerts/monitoring for what is actually a client error
2. **Poor Developer Experience**: API consumers receive unhelpful error messages
3. **Potential Exploit Vector**: Unhandled parsing errors can sometimes leak stack traces or be exploited
4. **Log Pollution**: Server logs fill with 500s that mask real server errors

---

## Acceptance Criteria

- [ ] Malformed UUIDs in path parameters return **400 Bad Request**
- [ ] Error response includes a clear message: `"Invalid ID format"`
- [ ] All endpoints with UUID path params are covered
- [ ] RBAC middleware handles malformed UUIDs gracefully
- [ ] All existing tests pass
- [ ] New tests validate 400 response for malformed UUIDs

---

## Implementation Steps

### Step 1: Create UUID Validation Utility

Create a reusable validation function in `pkg/httputil/`:

**File:** `pkg/httputil/validate.go`

```go
package httputil

import (
    "github.com/google/uuid"
)

// ParseUUID parses a string as UUID and returns an error with clear message if invalid.
func ParseUUID(s string) (uuid.UUID, error) {
    return uuid.Parse(s)
}

// IsValidUUID checks if a string is a valid UUID without returning the parsed value.
func IsValidUUID(s string) bool {
    _, err := uuid.Parse(s)
    return err == nil
}
```

### Step 2: Create Validation Middleware (Optional)

For endpoints that always require a UUID in the path, create middleware:

**File:** `pkg/middleware/validate_uuid.go`

```go
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/google/uuid"
    "github.com/onurceri/botla-co/internal/api"
)

// ValidatePathUUID extracts and validates a UUID from the URL path.
// pathPrefix is the part before the UUID (e.g., "/api/v1/chatbots/")
func ValidatePathUUID(pathPrefix string, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        if !strings.HasPrefix(path, pathPrefix) {
            next.ServeHTTP(w, r)
            return
        }
        
        // Extract potential UUID
        remainder := strings.TrimPrefix(path, pathPrefix)
        parts := strings.SplitN(remainder, "/", 2)
        idPart := parts[0]
        
        if idPart == "" {
            next.ServeHTTP(w, r)
            return
        }
        
        if _, err := uuid.Parse(idPart); err != nil {
            api.WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid ID format")
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Step 3: Update RBAC Middleware

Modify `internal/api/middleware/rbac.go` to validate UUIDs before database lookup:

**Before:**
```go
func (m *RBACMiddleware) CheckWorkspaceAccess(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        workspaceID := extractWorkspaceID(r.URL.Path)
        // ... database lookup that fails on invalid UUID
    })
}
```

**After:**
```go
func (m *RBACMiddleware) CheckWorkspaceAccess(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        workspaceID := extractWorkspaceID(r.URL.Path)
        
        // Validate UUID format before database lookup
        if _, err := uuid.Parse(workspaceID); err != nil {
            api.WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid workspace ID format")
            return
        }
        
        // ... continue with database lookup
    })
}
```

### Step 4: Update Handler-Level Validation

For handlers that extract UUIDs directly, add validation:

**Example in handlers:**
```go
func (h *Handler) GetChatbot(w http.ResponseWriter, r *http.Request) {
    chatbotID := extractChatbotID(r.URL.Path)
    
    // Validate UUID format
    if _, err := uuid.Parse(chatbotID); err != nil {
        api.WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid chatbot ID format")
        return
    }
    
    // Continue with business logic
    chatbot, err := h.db.GetChatbotByID(r.Context(), chatbotID)
    // ...
}
```

### Step 5: Update Integration Tests

Modify the RBAC test to expect 400:

**File:** `internal/integration/rbac_middleware_test.go`

```go
func TestRBACMiddleware_InvalidUUID(t *testing.T) {
    // ... setup ...
    
    req := httptest.NewRequest("GET", "/api/v1/workspaces/not-a-uuid/chatbots", nil)
    w := httptest.NewRecorder()
    
    handler.ServeHTTP(w, req)
    
    // Should return 400, not 500
    assert.Equal(t, http.StatusBadRequest, w.Code)
    
    // Should have clear error message
    var resp map[string]string
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.Contains(t, resp["error"], "Invalid")
}
```

---

## Testing Checklist

- [ ] `go build ./...` succeeds
- [ ] `make test-no-pdf` passes
- [ ] `make lint` passes
- [ ] Test: `curl /api/v1/chatbots/invalid` returns 400
- [ ] Test: `curl /api/v1/chatbots/not-a-uuid/sources` returns 400
- [ ] Test: RBAC middleware correctly returns 400 for malformed IDs

---

## Files to Modify

| File | Change |
|------|--------|
| `pkg/httputil/validate.go` | Create new file with UUID validation utilities |
| `internal/api/middleware/rbac.go` | Add UUID validation before DB lookup |
| `internal/api/handlers/*.go` | Add UUID validation to handlers |
| `internal/integration/rbac_middleware_test.go` | Update test expectations |

---

## Endpoints to Cover

These endpoint patterns need UUID validation:

| Pattern | Parameter |
|---------|-----------|
| `/api/v1/workspaces/:id/*` | workspace_id |
| `/api/v1/chatbots/:id/*` | chatbot_id |
| `/api/v1/sources/:id/*` | source_id |
| `/api/v1/organizations/:id/*` | organization_id |
| `/api/v1/public/chatbots/:id/*` | chatbot_id |

---

## Related Issues

- Code Audit Finding #4: "Unhandled Edge Cases in Security Middleware"
- Integration test TODO in `rbac_middleware_test.go`
