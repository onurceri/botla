# Task 00: Fix Transitional Service Initialization

**Priority:** 🔴 Critical  
**Effort:** Low (1-2 hours)  
**Risk Level:** High (potential nil pointer panics in production)

---

## Problem Statement

The `PublicChat` function in `internal/api/handlers/public.go` uses a deprecated "transitional pattern" that creates a `ChatService` with **4 nil dependencies**. This is explicitly marked as a TODO in the codebase:

```go
// Line 329-334
func PublicChat(dbpool *sql.DB) http.HandlerFunc {
    // Note: This is a transitional pattern; in production, ChatService should be properly initialized
    svc := services.NewChatService(dbpool, nil, nil, nil, nil) // ⚠️ 4 nils!
    h := &PublicHandlers{DB: dbpool, ChatService: svc}
    return h.PublicChat
}
```

### Why This Is Critical

1. **Nil Pointer Risk**: If any code path attempts to use `oaiClient`, `qdrantClient`, `storageService`, or `logger`, the application will panic
2. **Hidden Dependency Graph**: Manual service creation in handlers makes dependencies hard to track
3. **Inconsistent Initialization**: `PublicHandlers` struct exists for proper DI but is bypassed

---

## Acceptance Criteria

- [ ] The deprecated `PublicChat` function is removed from `internal/api/handlers/public.go`
- [ ] `PublicHandlers` is properly initialized in the router with fully constructed dependencies
- [ ] All public chat endpoints use the properly initialized `PublicHandlers`
- [ ] No `nil` values are passed to `services.NewChatService`
- [ ] All existing tests pass
- [ ] The codebase compiles without errors

---

## Implementation Steps

### Step 1: Audit Current Usage

Find all references to the deprecated `PublicChat` function:

```bash
grep -rn "PublicChat" --include="*.go" internal/ cmd/
```

**Expected locations:**
- `internal/api/handlers/public.go` (definition)
- `internal/api/router/routes.go` (likely usage)

### Step 2: Update Router Setup

Modify `internal/api/router/routes.go` (or equivalent) to:

1. Accept a properly constructed `ChatService` as a parameter
2. Create `PublicHandlers` with the injected service
3. Use `publicHandlers.PublicChat` instead of `handlers.PublicChat(db)`

**Before:**
```go
mux.HandleFunc("/api/v1/public/chatbots/", handlers.PublicChat(db))
```

**After:**
```go
publicHandlers := &handlers.PublicHandlers{
    DB:          db,
    ChatService: chatService, // Injected from main.go
}
mux.HandleFunc("/api/v1/public/chatbots/", publicHandlers.PublicChat)
```

### Step 3: Update `cmd/server/main.go`

Ensure `ChatService` is created in `newApplication()` and passed to the router:

```go
func newApplication(cfg *config.Config, log *logger.Logger) (*application, error) {
    // ... existing code ...
    
    // Create ChatService with all dependencies
    chatService := services.NewChatService(pool, oaiClient, qdrantClient, storageService, log)
    
    // Pass to router
    mux := router.New(cfg, pool, log, queue, storageService, qdrantClient, chatService)
    
    // ...
}
```

### Step 4: Remove Deprecated Function

Delete the following from `internal/api/handlers/public.go`:

```go
// PublicChatFunc returns a http.HandlerFunc for backwards compatibility
//
// Deprecated: Use PublicHandlers.PublicChat instead
func PublicChat(dbpool *sql.DB) http.HandlerFunc {
    // Create a ChatService for backwards compatibility
    // Note: This is a transitional pattern; in production, ChatService should be properly initialized
    svc := services.NewChatService(dbpool, nil, nil, nil, nil)
    h := &PublicHandlers{DB: dbpool, ChatService: svc}
    return h.PublicChat
}
```

### Step 5: Update Tests

If any tests use the deprecated function, update them to use `PublicHandlers` directly:

```go
func TestPublicChat(t *testing.T) {
    mockChatService := &mocks.ChatService{...}
    handlers := &handlers.PublicHandlers{
        DB:          testDB,
        ChatService: mockChatService,
    }
    // Test handlers.PublicChat
}
```

---

## Testing Checklist

- [ ] `go build ./...` succeeds
- [ ] `make test-no-pdf` passes
- [ ] `make lint` passes
- [ ] Manual test: Public chat endpoint responds correctly
- [ ] Integration tests for public chat work

---

## Files to Modify

| File | Change |
|------|--------|
| `internal/api/handlers/public.go` | Remove deprecated `PublicChat` function |
| `internal/api/router/routes.go` | Update to use `PublicHandlers` struct |
| `cmd/server/main.go` | Ensure `ChatService` is created and passed to router |
| `internal/api/router/*.go` | Update function signatures if needed |

---

## Rollback Plan

If issues arise, the deprecated function can be restored temporarily. However, the goal is complete removal to prevent future nil pointer bugs.

---

## Related Issues

- Code Audit Finding #1: "Transitional and Brittle Service Initialization"
- Architectural Review: "Suggested Improvement - Extract DI into internal/app package"
