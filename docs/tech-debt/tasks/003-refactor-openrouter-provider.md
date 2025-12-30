# TASK-003 — Refactor OpenRouter Provider to Use Common Client

## Goal
Update the OpenRouter implementation to delegate HTTP communication to the new shared client, ensuring consistent behavior with OpenAI.

## Scope
- `internal/ai/openrouter` package.

## Checklist
- [x] Modify `openrouter.Embedder` to hold an instance of `ai.BaseClient`.
- [x] Configure `BaseClient` with OpenRouter-specific headers (`HTTP-Referer`, `X-Title`).
- [x] Update `Embed` method to use `BaseClient.Post`.
- [x] Update `EmbedBatch` method to use `BaseClient.Post`.
- [x] Run existing tests in `internal/ai/openrouter`.
- [x] Delete duplicated retry logic and structs.

## Edge Cases
- Headers specific to OpenRouter must be preserved.

## Files Likely to Change
- `internal/ai/openrouter/embedder.go`
