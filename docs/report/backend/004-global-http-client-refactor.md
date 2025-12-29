# Backend Task 004: Refactor Global HTTP Client to Dependency Injection

## Background
`internal/rag/openai.go` uses a global variable `GlobalHTTPClient` for overriding the HTTP client in tests. This pattern causes race conditions when running tests in parallel and makes the code harder to reason about and maintain.

**File:** `internal/rag/openai.go`
**Location:** Lines 26-27

## Integration Plan
1.  **Remove Global Variable**
    - Remove `var GlobalHTTPClient *http.Client`.

2.  **Update OpenAIClient Struct**
    - Ensure `OpenAIClient` struct has an exported or accessible way to set the HTTP client (it already has a `http` field, maybe make it public or add a `SetHTTPClient` method).
    - Alternatively, update `NewOpenAIClient` to accept an optional `HTTPClient` interface or options pattern.

3.  **Update Tests**
    - Find tests using `GlobalHTTPClient`.
    - Refactor them to pass the mock HTTP client directly to the `OpenAIClient` instance being tested.

4.  **Verify**
    - Run `go test -race ./internal/rag/...` to ensure no race conditions.

## Checklist
- [ ] Remove `GlobalHTTPClient` variable
- [ ] Allow identifying/injecting HTTP client in `OpenAIClient` instance
- [ ] Refactor unit tests to inject client per instance
- [ ] Run race detector on tests
