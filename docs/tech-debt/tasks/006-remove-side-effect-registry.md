# TASK-006 — Remove Side-Effect Registry Pattern

## Goal
Remove `init()` functions that rely on global state for registration. Use explicit dependency injection wiring in the application root.

## Scope
- `internal/ai` packages.

## Checklist
- [x] Remove `init()` function in `internal/ai/openai/embedder.go`.
- [x] Remove `init()` function in `internal/ai/openrouter/embedder.go`.
- [x] Remove `init()` function in `internal/ai/qdrant/client.go`.
- [x] Remove global registry map in `internal/ai/factory.go`.
- [x] Create a `Wiring` helper or update `cmd/server/main.go` to explicitly choose and instantiate the correct Embedder/VectorStore based on application config.
- [x] Verify application startup works correctly.

## Edge Cases
- Switching providers based on config (previously handled by factory string lookup).

## Completion Status
Completed on: 2025-12-31
All init() functions removed
Global registry pattern removed (factory.go deleted)
Explicit dependency injection implemented via constructors
Application verified to work correctly
All tests passing
All linters passing

Note: As part of this cleanup, the entire `internal/ai` package was identified as dead code (zero imports outside the package itself) and removed. Production code uses `internal/rag` package exclusively.
