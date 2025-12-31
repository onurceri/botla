# Removed Packages

## internal/ai (Removed 2025-12-31)

### Summary
Removed as dead code - zero production imports.

### Files Removed
- client.go (BaseClient with retry logic)
- client_test.go
- interfaces.go (VectorStore, Embedder)
- types.go
- mocks.go
- mocks_test.go
- README.md
- openai/embedder.go
- openai/embedder_test.go
- openrouter/embedder.go
- openrouter/embedder_test.go
- qdrant/client.go
- qdrant/client_test.go

### Total
1,253 lines across 12 files removed.

### Replacement
Use `internal/rag` package instead:
- `rag.NewQdrantClient()` for vector operations
- `rag.NewOpenAIClient()` for OpenAI operations
- `rag.NewOpenRouterClient()` for OpenRouter operations
