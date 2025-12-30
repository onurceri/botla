# TASK-001 — Extract Common AI HTTP Client

## Goal
Create a shared, robust HTTP client abstraction in `internal/ai` that handles the common logic found in both OpenAI and OpenRouter implementations: request marshaling, retries with exponential backoff, and error parsing.

## Scope
- **Included**: Creating `internal/ai/http_client.go` (or similar), defining generic request/response structs if applicable, implementing the retry loop.
- **Excluded**: modifying existing providers (that is mostly for the next tasks, though simple wiring is fine to test).

## Checklist
- [x] Create `internal/ai/base_embedder.go` or `client.go`.
- [x] Define a `BaseClient` struct that accepts a generic configuration (BaseURL, APIKey, Header map).
- [x] Implement `Post(ctx, path, body, responseTarget)` method with:
    - [x] 4 retries with exponential backoff.
    - [x] `context` support.
    - [x] JSON marshaling/unmarshaling.
    - [x] HTTP status code error handling.
- [x] Write unit tests for `BaseClient` covering:
    - [x] Successful request.
    - [x] 429 Rate limit retry behavior.
    - [x] 5xx Server error retry behavior.
    - [x] Non-retryable errors (401, 400).
    - [x] Context cancellation.

## Edge Cases
- Network timeouts during retries.
- Malformed JSON responses.
- Empty bodies.

## Files Likely to Change
- `internal/ai/client.go` (NEW)
- `internal/ai/client_test.go` (NEW)
