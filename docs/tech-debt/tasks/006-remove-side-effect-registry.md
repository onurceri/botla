# TASK-006 — Remove Side-Effect Registry Pattern

## Goal
Remove `init()` functions that rely on global state for registration. Use explicit dependency injection wiring in the application root.

## Scope
- `internal/ai` packages.

## Checklist
- [ ] Remove `init()` function in `internal/ai/openai/embedder.go`.
- [ ] Remove `init()` function in `internal/ai/openrouter/embedder.go`.
- [ ] Remove `init()` function in `internal/ai/qdrant/client.go`.
- [ ] Remove global registry map in `internal/ai/factory.go`.
- [ ] Create a `Wiring` helper or update `cmd/server/main.go` to explicitly choose and instantiate the correct Embedder/VectorStore based on application config.
- [ ] Verify application startup works correctly.

## Edge Cases
- Switching providers based on config (previously handled by factory string lookup).

## Files Likely to Change
- `internal/ai/openai/embedder.go`
- `internal/ai/openrouter/embedder.go`
- `internal/ai/factory.go`
- `cmd/server/main.go`
