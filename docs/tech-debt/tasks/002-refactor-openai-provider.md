# TASK-002 — Refactor OpenAI Provider to Use Common Client

## Goal
Update the OpenAI implementation to delegate all HTTP communication to the new shared client, removing duplicated logic.

## Scope
- `internal/ai/openai` package.

## Checklist
- [x] Modify `openai.Embedder` to hold an instance of `ai.BaseClient`.
- [x] Update `Embed` method to use `BaseClient.Post`.
- [x] Update `EmbedBatch` method to use `BaseClient.Post`.
- [x] Ensure specific OpenAI headers are maintained (if any differing from base).
- [x] Verify `NewFromEnv` still works (temporarily, until TASK-005).
- [x] Run existing tests in `internal/ai/openai` to ensure no regression.
- [x] Remove deleted code (old retry loop, old Structs if now shared).

## Edge Cases
- API-specific error message formats.

## Files Likely to Change
- `internal/ai/openai/embedder.go`
