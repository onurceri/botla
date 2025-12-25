# Test Setup Analysis and Issues Report

This document provides a comprehensive analysis of the project's test infrastructure and documents issues, anti-patterns, and areas for improvement.

---

## Executive Summary

The project has a well-structured testing foundation with:
- **Go backend**: ~100+ test files using testify, pgx, and custom test utilities
- **Frontend**: 60+ component tests using Vitest and React Testing Library
- **E2E**: Playwright-based tests
- **Test DB infrastructure**: Sophisticated schema isolation with parallel execution support

However, several issues and anti-patterns were identified that could lead to flakiness, maintenance burden, and reduced test reliability.

---

## ✅ Fixed Issues (2025-12-25)

The following critical test infrastructure issues have been resolved:

### Database Configuration Issues

1. **`DOCKER_TEST_DATABASE_URL` pointed to wrong database** (`Makefile:10`)
   - Was: `$(DOCKER_DATABASE_URL)&options=-c%20search_path%3Dtest` (pointed to `botla_dev`)
   - Fixed: Now explicitly points to `botla_test` database

2. **Schema creation missing before migrations** (`internal/testdb/testdb.go`)
   - `OpenTestDBWithSchema` now creates the schema before running migrations
   - Prevents "schema does not exist" errors

### Race Condition Fixes

3. **Parallel test schema cleanup race condition** (`internal/testdb/testdb.go`)
   - Added `activeSchemas` tracking to prevent cleanup of in-use schemas
   - Added time-based threshold (5 min) for stale schema detection
   - Added migration caching to avoid redundant migrations

4. **Integration test cleanup failures** (`internal/integration/testutils.go`)
   - `TeardownTestEnv` now uses a fresh connection for cleanup
   - Added retry logic (3 attempts) for schema drops
   - Terminates idle connections before dropping schemas

### New Cleanup Functions

5. **Added `CleanupAllTestSchemas()`** - Explicit cleanup for `botla_test_*` schemas
6. **Added `CleanupAllIntegrationSchemas()`** - Explicit cleanup for `botla_it_*` schemas
7. **Automatic cleanup at suite start** - Integration tests now clean stale schemas at start

---

## Critical Issues

### 1. Unused Mock Objects in TestEnv (`internal/integration/testutils.go:231`)

**Issue**: The `SetupTestEnv()` function creates `mockVC` and `mockLLM` but sets them to `nil` in the returned `TestEnv` struct:

```go
mockVC := &rag.MockVectorClient{}
mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)
// ... more expectations ...

mockLLM := &rag.MockFullClient{}
// ... more expectations ...

// Line 231 - BUG: Sets to nil even though they were just created!
return &TestEnv{Cfg: cfg, DB: db, Schema: schema, Server: srv, VectorStore: vs, MockVC: nil, MockLLM: nil, Queue: q}, nil
```

**Impact**: Tests that rely on `te.MockVC` or `te.MockLLM` will get nil pointer panics. The mocks are configured but never used.

**Recommendation**: Assign the created mocks:
```go
return &TestEnv{Cfg: cfg, DB: db, Schema: schema, Server: srv, VectorStore: vs, MockVC: mockVC, MockLLM: mockLLM, Queue: q}, nil
```

---

### 2. Ignored Error in SourceQueue Startup (`internal/integration/testserver.go:118`)

**Issue**: The error from `StartSourceQueue` is silently ignored:

```go
q, _ := processing.StartSourceQueue(pool, memStore, actualLLM, actualVC)
```

**Impact**: If queue startup fails, tests continue with a nil queue, causing panics or silent failures.

**Recommendation**:
```go
q, err := processing.StartSourceQueue(pool, memStore, actualLLM, actualVC)
if err != nil {
    return nil, err
}
```

---

### 3. React QueryClient Created on Every Render (`frontend/src/test-utils.tsx:45-54`)

**Issue**: `QueryWrapper` creates a new `QueryClient` inside the component render:

```tsx
export function QueryWrapper({ children }: { children: ReactNode }) {
    const queryClient = createTestQueryClient()  // Creates new client on every render!
    return (
        <QueryClientProvider client={queryClient}>
            {/* ... */}
        </QueryClientProvider>
    )
}
```

**Impact**:
- Breaks React's referential equality checks
- Causes unnecessary re-renders
- Can lead to stale cache issues
- Query state is lost on every parent re-render

**Recommendation**: Use `React.useMemo` or create the client outside the component:

```tsx
const testQueryClient = createTestQueryClient()

export function QueryWrapper({ children }: { children: ReactNode }) {
    return (
        <QueryClientProvider client={testQueryClient}>
            {/* ... */}
        </QueryClientProvider>
    )
}
```

---

### 4. Mock Widget Makes Real API Calls (`frontend/src/setupTests.ts:97-151`)

**Issue**: The mock WidgetApp component calls the real `api.post` function:

```tsx
const send = async () => {
    const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/chat`, { message: text })  // Real API call!
    // ...
}
```

**Impact**:
- Violates the test isolation principle
- Creates unexpected network dependencies
- Can fail in CI environments without API access
- Makes tests slower

**Recommendation**: Mock the API call:
```tsx
const send = async () => {
    // Mock the response instead
    setMessages((m) => [...m, 'Mock response'])
}
```

---

### 5. Conflicting LocalStorage Mocks

**Issue**: Tests re-define localStorage mocks that conflict with the global mock in `setupTests.ts`:

`frontend/src/__tests__/App.auth-redirect.test.tsx:10-17`:
```tsx
Object.defineProperty(window, 'localStorage', {
    value: {
        getItem: vi.fn().mockReturnValue(null),
        // ...
    },
    writable: true,
})
```

`frontend/src/pages/__tests__/LoginPage.test.tsx:11-18`:
```tsx
Object.defineProperty(window, 'localStorage', {
    value: {
        getItem: vi.fn(),
        // ...
    },
    writable: true,
})
```

**Impact**:
- Overwrites the global mock, breaking its state
- Inconsistent behavior between tests
- Tests may interfere with each other

**Recommendation**: Remove redundant localStorage mocks and use the global one from `setupTests.ts`.

---

## High Priority Issues

### 6. Unrealistic 100% Coverage Thresholds (`frontend/vite.config.js:27-35`)

**Issue**:
```javascript
coverage: {
    lines: 100,
    functions: 100,
    branches: 100,
    statements: 100,
}
```

**Impact**:
- 100% coverage is extremely difficult to maintain
- Encourages writing tests for edge cases that don't add value
- Can block legitimate PRs
- Tests may become brittle and hard to refactor

**Recommendation**: Set realistic thresholds (e.g., 80-85% for statements, 75-80% for branches).

---

### 7. Pointer to Bool in JSON Body (`internal/integration/privacy_test.go:41-46`)

**Issue**:
```go
marketing := true
analytics := false
body := map[string]*bool{
    "marketing": &marketing,
    "analytics": &analytics,
}
b, _ := json.Marshal(body)
```

**Impact**:
- Unusual pattern that's hard to read
- Pointer semantics for boolean values is non-idiomatic Go
- The `&marketing` address-of pattern can cause issues if the variable is reused

**Recommendation**: Use proper struct types or `json.RawMessage` for dynamic JSON.

---

### 8. Ignored JSON Marshal Errors

Multiple instances of ignoring `json.Marshal` errors:

- `internal/integration/privacy_test.go:47`: `b, _ := json.Marshal(body)`
- `internal/integration/privacy_test.go:103`: `b, _ := json.Marshal(body)`

**Impact**: Silent failures if the body can't be marshaled, leading to confusing test failures.

**Recommendation**:
```go
b, err := json.Marshal(body)
require.NoError(t, err)
```

---

### 9. Hardcoded UUIDs in Tests (`internal/api/handlers/admin_queues_test.go`)

**Issue**:
```go
langID := "00000000-0000-0000-0000-000000000001"
planID := "00000000-0000-0000-0000-000000000001"
userID := "00000000-0000-0000-0000-000000000001"
chatbotID := "00000000-0000-0000-0000-000000000001"
stuckID := "00000000-0000-0000-0000-000000000002"
```

**Impact**:
- All these tests use the same UUIDs
- If tests run in parallel, they will conflict on foreign key constraints
- Violates test isolation principle

**Recommendation**: Use `gen_random_uuid()` or generate unique UUIDs per test.

---

### 10. Environment Variable Side Effects (`internal/integration/testutils.go:49-77`)

**Issue**: `SetupTestEnv()` modifies global environment variables without cleanup:

```go
if os.Getenv("DB_HOST") == "" {
    _ = os.Setenv("DB_HOST", "localhost")
}
// ... more Setenv calls ...
```

**Impact**:
- Tests can leak environment state to other tests
- Can cause unexpected behavior in CI
- Makes debugging harder

**Recommendation**: Save and restore environment state:
```go
originalHost := os.Getenv("DB_HOST")
defer os.Setenv("DB_HOST", originalHost)
os.Setenv("DB_HOST", "localhost")
```

---

## Medium Priority Issues

### 11. Inconsistent Parallel Test Usage

Only 3 test files use `t.Parallel()`:
- `internal/testdb/testdb.go:155` (commented out)
- `internal/testdb/withtx_test.go:113, 141`

Most tests don't use parallel execution, which:
- Slows down test suite
- Underutilizes CI resources

**Recommendation**: Enable parallel execution for independent tests using `OpenParallelTestDB`.

---

### 12. Integration Tests Share Schema (`internal/integration/testutils.go:153`)

**Issue**:
```go
_, _ = db.Exec(`TRUNCATE TABLE chatbots, users, organizations, workspaces, data_sources, analytics, handoff_requests, messages, conversations CASCADE`)
```

**Impact**:
- Tests share the same schema
- TRUNCATE is not transactional within a test
- Can cause race conditions in parallel test runs

**Recommendation**: Each integration test should use a unique schema like unit tests do.

---

### 13. Silent Migration Failures (`internal/testdb/testdb.go:289-292`)

**Issue**:
```go
if output, err := cmd.CombinedOutput(); err != nil {
    t.Logf("migration info: %s (error: %v)", string(output), err)
}
```

**Impact**: Migration failures only log a message instead of failing the test. Tests may run against an incorrect schema.

**Recommendation**:
```go
if output, err := cmd.CombinedOutput(); err != nil {
    t.Fatalf("migration failed: %s (error: %v)", string(output), err)
}
```

---

### 14. Test Flakiness from Time-Based Assertions

`internal/integration/privacy_test.go:156-168` uses a polling loop with timeout:

```go
deadline := time.Now().Add(5 * time.Second)
for {
    // ...
    if time.Now().After(deadline) {
        t.Fatalf("timed out waiting for export request to complete")
    }
    time.Sleep(100 * time.Millisecond)
}
```

**Impact**:
- Can be flaky under heavy system load
- 5 seconds may not be enough in CI
- Polling is less reliable than channels or callbacks

**Recommendation**: Use a context with timeout and `select` with default case for non-blocking checks.

---

### 15. Inconsistent Error Handling Patterns

Some tests use `require.NoError` where `assert.NoError` would be better for cleanup scenarios:

```go
require.NoError(t, err)  // Stops test immediately
vs
assert.NoError(t, err)   // Continues, allowing cleanup
```

**Recommendation**: Use `assert` when you need to perform cleanup before failing.

---

## Low Priority Issues

### 16. Missing Test Documentation

Many test files lack doc comments explaining:
- What the test validates
- What setup is required
- What edge cases are covered

### 17. No Test Data Factories

Tests create test data with inline SQL, making them:
- Verbose and repetitive
- Hard to maintain when schema changes
- Inconsistent between tests

**Recommendation**: Create factory functions in a `testfixtures` package.

### 18. E2E Tests Skip Without Environment

`frontend/e2e/*.spec.ts` files check `!process.env.E2E_API_BASE` and skip:
- Makes CI configuration complex
- Silent skips can mask test problems

### 19. No Test Naming Convention

Tests use inconsistent naming:
- `TestMe_Success` (underscore)
- `TestPrivacyFlow` (CamelCase)
- `TestGetQueues` (CamelCase with descriptive subtest)

**Recommendation**: Adopt `TestXxx_Yyy` convention consistently.

---

## Architecture Issues

### 20. Duplicate Test Utilities

Multiple files provide similar functionality:
- `internal/integration/testutils.go`
- `internal/integration/test_utils.go`
- `internal/testdb/testdb.go`
- `internal/integration/mock_llm.go`

**Impact**: Confusing which utility to use, code duplication.

**Recommendation**: Consolidate into a single `internal/testutils` package.

---

### 21. Testserver.go is Duplicated in Complexity

`testserver.go` manually registers routes instead of using the production router:
- Duplication risk
- Routes may diverge from production
- Harder to maintain

**Recommendation**: Export a test router builder that reuses production route registration.

---

## Recommendations Summary

### Immediate Actions (Critical)
1. Fix the `MockVC`/`MockLLM` nil assignment bug
2. Handle the `StartSourceQueue` error properly
3. Fix the QueryClient creation in `QueryWrapper`
4. Remove real API calls from mock WidgetApp

### Short-Term Actions (High Priority)
1. Reduce coverage thresholds to realistic levels
2. Remove conflicting localStorage mocks
3. Use proper error handling for JSON marshal
4. Generate unique UUIDs for test data
5. Add environment variable cleanup

### Medium-Term Actions
1. Consolidate test utilities
2. Enable parallel testing more broadly
3. Create test data factories
4. Improve test documentation
5. Standardize test naming

### Long-Term Improvements
1. Add integration test isolation (per-test schemas)
2. Implement test retry/failure analysis
3. Create test coverage reports per package
4. Add mutation testing for critical packages
5. Establish test performance benchmarks

---

## Appendix: File Reference

| File | Issues |
|------|--------|
| `internal/integration/testutils.go` | #1, #10, #12 |
| `internal/integration/testserver.go` | #2 |
| `internal/integration/privacy_test.go` | #6, #7, #14 |
| `internal/api/handlers/admin_queues_test.go` | #9 |
| `internal/testdb/testdb.go` | #13 |
| `frontend/vite.config.js` | #6 |
| `frontend/src/setupTests.ts` | #4 |
| `frontend/src/test-utils.tsx` | #3 |
| `frontend/src/__tests__/App.auth-redirect.test.tsx` | #5 |
| `frontend/src/pages/__tests__/LoginPage.test.tsx` | #5 |
