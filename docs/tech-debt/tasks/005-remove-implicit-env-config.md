# TASK-005 — Remove Implicit Environment Configuration in AI Packages

## Goal
Make dependencies explicit by removing `NewFromEnv` and `os.Getenv` usages inside low-level AI packages. Pass configuration down from the top level.

## Scope
- `internal/ai/openai`
- `internal/ai/openrouter`
- `internal/ai/qdrant`

## Checklist
- [ ] Update `openai.NewEmbedder` to accept a Config struct (Key, BaseURL, Model).
- [ ] Remove `openai.NewFromEnv`.
- [ ] Update `openrouter.NewEmbedder` to accept a Config struct.
- [ ] Remove `openrouter.NewFromEnv`.
- [ ] Update `qdrant.NewStore` to accept Config.
- [ ] Remove `qdrant.NewFromEnv`.
- [ ] Update `cmd/server/main.go` and `factory.go` to load config from `config.Config` and pass it down.
- [ ] Update integration tests to manually construct services instead of relying on Env.

## Edge Cases
- Missing configuration values (should fail at call site, not deep in library).

## Files Likely to Change
- `internal/ai/openai/embedder.go`
- `internal/ai/openrouter/embedder.go`
- `internal/ai/qdrant/client.go`
- `internal/ai/factory.go`
- `cmd/server/main.go`
