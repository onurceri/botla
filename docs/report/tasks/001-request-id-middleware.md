# Task 001: Implement Request-ID Middleware

**Priority:** 🔴 Critical (Foundation)  
**Phase:** 1 - Observability Foundation  
**Estimated Time:** 2-3 hours  
**Dependencies:** None  

---

## Problem Statement

Currently, there is no request-ID propagation across the system. This makes debugging production issues extremely difficult as there's no way to correlate logs from a single request across different services and layers.

**Evidence from codebase:**
- `pkg/logger/logger.go` has structured JSON logging but no request-ID field
- No middleware generates or propagates request IDs
- grep for "request-id" returns no results

---

## Objective

Implement a middleware that:
1. Generates a unique request ID for each incoming HTTP request
2. Extracts existing request ID from `X-Request-ID` header (for distributed tracing)
3. Adds request ID to response headers
4. Makes request ID available in context for logging

---

## Implementation Details

### Step 1: Create Request-ID Middleware

**File:** `pkg/middleware/request_id.go` (NEW)

```go
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDKey is the context key for request ID
type requestIDKey struct{}

// RequestIDHeader is the header name for request ID
const RequestIDHeader = "X-Request-ID"

// RequestID middleware generates or extracts a request ID for each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get existing request ID from header
		requestID := r.Header.Get(RequestIDHeader)
		
		// Generate new one if not provided
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Add to response header
		w.Header().Set(RequestIDHeader, requestID)
		
		// Add to context
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext extracts request ID from context
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
```

### Step 2: Update Logger to Accept Request ID

**File:** `pkg/logger/logger.go` (MODIFY)

Add a method that accepts context and automatically extracts request ID:

```go
// InfoCtx logs with request ID from context
func (l *Logger) InfoCtx(ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if reqID := middleware.RequestIDFromContext(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	l.write("INFO", msg, fields)
}

// ErrorCtx logs with request ID from context
func (l *Logger) ErrorCtx(ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if reqID := middleware.RequestIDFromContext(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	l.write("ERROR", msg, fields)
}

// WarnCtx logs with request ID from context
func (l *Logger) WarnCtx(ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if reqID := middleware.RequestIDFromContext(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	l.write("WARN", msg, fields)
}

// DebugCtx logs with request ID from context
func (l *Logger) DebugCtx(ctx context.Context, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if reqID := middleware.RequestIDFromContext(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	l.write("DEBUG", msg, fields)
}
```

### Step 3: Register Middleware in Router

**File:** `internal/api/router/router.go` (MODIFY)

Find where middleware is applied and add request ID middleware as the FIRST middleware:

```go
// In the middleware chain, add RequestID first
handler := middleware.RequestID(mux)
handler = middleware.CORS(handler)
handler = middleware.Recovery(handler, log, cfg.Environment)
// ... other middleware
```

### Step 4: Update Key Handlers to Use Context Logging

Start with critical handlers that need tracing:

**Files to update:**
- `internal/api/handlers/chat.go` - Use `log.InfoCtx(r.Context(), ...)`
- `internal/api/handlers/source_create.go` - Use `log.InfoCtx(r.Context(), ...)`
- `internal/processing/sources_queue.go` - Pass context through pipeline

---

## Tests to Write

### Unit Tests

**File:** `pkg/middleware/request_id_test.go` (NEW)

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesNewID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request ID is in context
		reqID := RequestIDFromContext(r.Context())
		if reqID == "" {
			t.Error("expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify request ID is in response header
	respID := rec.Header().Get(RequestIDHeader)
	if respID == "" {
		t.Error("expected request ID in response header")
	}
	
	// Verify it looks like a UUID
	if len(respID) != 36 {
		t.Errorf("expected UUID format, got %s", respID)
	}
}

func TestRequestID_UsesExistingHeader(t *testing.T) {
	existingID := "existing-request-id-12345"
	
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := RequestIDFromContext(r.Context())
		if reqID != existingID {
			t.Errorf("expected %s, got %s", existingID, reqID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify same ID is returned
	respID := rec.Header().Get(RequestIDHeader)
	if respID != existingID {
		t.Errorf("expected %s in response, got %s", existingID, respID)
	}
}

func TestRequestIDFromContext_NoID(t *testing.T) {
	ctx := context.Background()
	id := RequestIDFromContext(ctx)
	if id != "" {
		t.Errorf("expected empty string, got %s", id)
	}
}
```

### Integration Test

**File:** `internal/integration/request_id_test.go` (NEW)

```go
package integration

import (
	"net/http"
	"testing"
)

func TestRequestID_Integration(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	// Test 1: Response includes generated request ID
	resp, err := http.Get(te.Server.URL + "/api/v1/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	reqID := resp.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Error("expected X-Request-ID header in response")
	}

	// Test 2: Provided request ID is returned
	client := &http.Client{}
	req, _ := http.NewRequest("GET", te.Server.URL+"/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "test-id-12345")
	
	resp2, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.Header.Get("X-Request-ID") != "test-id-12345" {
		t.Error("expected provided request ID to be returned")
	}
}
```

---

## Verification Steps

1. **Run unit tests:**
   ```bash
   go test ./pkg/middleware/... -v -run TestRequestID
   ```

2. **Run integration tests:**
   ```bash
   go test ./internal/integration/... -v -run TestRequestID
   ```

3. **Manual verification:**
   ```bash
   # Start the server
   make be-run
   
   # Make a request and check headers
   curl -i http://localhost:8080/api/v1/health
   # Should see X-Request-ID in response headers
   
   # Provide custom request ID
   curl -i -H "X-Request-ID: my-custom-id" http://localhost:8080/api/v1/health
   # Should see X-Request-ID: my-custom-id in response
   ```

4. **Check logs include request ID:**
   ```bash
   # Make a request that triggers logging
   curl http://localhost:8080/api/v1/chatbots
   # Check server logs for request_id field in JSON output
   ```

---

## Acceptance Criteria

- [ ] Every HTTP response includes `X-Request-ID` header
- [ ] If client provides `X-Request-ID`, it's preserved
- [ ] Request ID is accessible via `RequestIDFromContext(ctx)`
- [ ] Logger methods with `Ctx` suffix include request_id in output
- [ ] All unit tests pass
- [ ] Integration test passes
- [ ] Manual verification confirms headers are present

---

## Files Changed

| File | Action |
|------|--------|
| `pkg/middleware/request_id.go` | CREATE |
| `pkg/middleware/request_id_test.go` | CREATE |
| `pkg/logger/logger.go` | MODIFY |
| `internal/api/router/router.go` | MODIFY |
| `internal/integration/request_id_test.go` | CREATE |

---

## Rollback Plan

If issues arise:
1. Remove `middleware.RequestID` from router middleware chain
2. Keep the middleware code but don't use it
3. Logger changes are backward compatible (old methods still work)
