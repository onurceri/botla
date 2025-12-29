# Backend Task 002: Handle Ignored Request Parsing Error

## Background
In `internal/rag/openai.go`, the error return from `http.NewRequestWithContext` is currently ignored using `_`. While unlikely to fail with static inputs, robust software should handle all potential errors to prevent silent failures or panics in edge cases (e.g., invalid method or URL).

**File:** `internal/rag/openai.go`
**Location:** Line ~91

## Integration Plan
1.  **Locate Ignored Error**
    - Search for `req, _ := http.NewRequestWithContext` in `internal/rag/openai.go`.

2.  **Add Error Handling**
    - Capture the error: `req, err := http.NewRequestWithContext(...)`
    - If `err != nil`, return an appropriate error immediately (e.g., `fmt.Errorf("failed to create request: %w", err)`).

3.  **Verify**
    - Verify code compiles.
    - Briefly check if `CreateEmbeddingsBatch` or other methods share this pattern and fix them too.

## Checklist
- [x] Locate `http.NewRequestWithContext` call in `CreateEmbedding`
- [x] Check `CreateEmbeddingsBatch` and `CreateCompletion` for similar issues
- [x] Add proper error handling for request creation
- [x] Run tests to ensure no regressions

