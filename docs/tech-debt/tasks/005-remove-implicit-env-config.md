# TASK-005 — Remove Implicit Environment Configuration in AI Packages

## Goal
Make dependencies explicit by removing `NewFromEnv` and `os.Getenv` usages inside low-level AI packages. Pass configuration down from the top level.

## Scope
- `internal/ai/openai`
- `internal/ai/openrouter`
- `internal/ai/qdrant`

## Checklist
- [x] Update `openai.NewEmbedder` to accept a Config struct (Key, BaseURL, Model).
- [x] Remove `openai.NewFromEnv`.
- [x] Update `openrouter.NewEmbedder` to accept a Config struct.
- [x] Remove `openrouter.NewFromEnv`.
- [x] Update `qdrant.NewStore` to accept Config.
- [x] Remove `qdrant.NewFromEnv`.
- [x] Update `cmd/server/main.go` and `factory.go` to load config from `config.Config` and pass it down.
- [x] Update integration tests to manually construct services instead of relying on Env.

## Completion Status
Completed on: 2025-12-31
All NewFromEnv functions removed
Factory pattern removed (unused code)
Config structs added to each package
Fail-fast validation implemented
All tests passing
All linters passing

## Edge Cases
- Missing configuration values (should fail at call site, not deep in library).

## Files Likely to Change
- `internal/ai/openai/embedder.go`
- `internal/ai/openrouter/embedder.go`
- `internal/ai/qdrant/client.go`
- `internal/ai/factory.go`
- `cmd/server/main.go`
