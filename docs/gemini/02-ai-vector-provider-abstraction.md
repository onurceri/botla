# Task 02: AI/Vector Provider Interface Abstraction

## Priority
**Medium** - Strategic improvement for future flexibility

## Problem Statement

The RAG pipeline has direct dependencies on specific implementations of OpenAI and Qdrant. While a `VectorClient` interface exists in `internal/rag/qdrant.go`, there is no corresponding abstraction for the LLM/Embedder components, and the interface definitions are co-located with their implementations.

This coupling means switching from OpenAI to Anthropic, or from Qdrant to PGVector, would require changes throughout the ingestion and retrieval logic rather than just swapping a driver.

## Current State Analysis

### What's Already Done (Positive)

The codebase already has some abstractions:

```go
// internal/rag/qdrant.go - VectorClient interface exists
type VectorClient interface {
    EnsureEmbeddingsCollection(ctx context.Context) error
    UpsertEmbedding(ctx context.Context, id interface{}, vector []float32, payload EmbeddingPayload) error
    SearchSimilar(ctx context.Context, embedding []float32, chatbotID string, topK int) ([]SearchResult, error)
    DeleteBySourceID(ctx context.Context, sourceID string) error
    ScrollChunks(ctx context.Context, sourceID string, limit int, offset interface{}) ([]SearchResult, *string, error)
}

// internal/rag/llm_client.go - LLMClient interface exists
type LLMClient interface {
    // ...
}
```

### What Needs Improvement

1. **Interface Location**: Interfaces are defined alongside implementations (e.g., `VectorClient` in `qdrant.go`)
2. **Missing Embedder Interface**: No explicit `Embedder` interface for embedding providers
3. **Direct Instantiation**: Services directly instantiate specific clients via `NewQdrantClientFromEnv()`, `NewOpenAIClientFromEnv()`
4. **No Provider Registry**: No factory pattern to switch providers based on configuration

## Refactoring Goals

1. **Centralized Interfaces**: Move all provider interfaces to a dedicated package
2. **Embedder Abstraction**: Create explicit interface for embedding generation
3. **Factory Pattern**: Implement provider factories for runtime selection
4. **Configuration-Driven**: Enable provider switching via environment variables

## Implementation Plan

### Phase 1: Create Interface Package

**New Directory**: `internal/ai/`

**File**: `internal/ai/interfaces.go`

```go
package ai

import (
    "context"
    "time"
)

// VectorStore abstracts vector database operations
type VectorStore interface {
    EnsureCollection(ctx context.Context) error
    Upsert(ctx context.Context, id interface{}, vector []float32, payload VectorPayload) error
    Search(ctx context.Context, vector []float32, filter SearchFilter, limit int) ([]SearchResult, error)
    Delete(ctx context.Context, filter DeleteFilter) error
    Scroll(ctx context.Context, filter SearchFilter, limit int, offset interface{}) ([]SearchResult, interface{}, error)
}

// Embedder abstracts text embedding generation
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
    Dimension() int
}

// ChatCompleter abstracts chat completion
type ChatCompleter interface {
    Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    CompleteStream(ctx context.Context, req ChatRequest, handler StreamHandler) error
}

// Domain types
type VectorPayload struct {
    ChatbotID    string
    SourceID     string
    ChunkIndex   int
    OriginalText string
    SourceType   string
    CreatedAt    time.Time
}

type SearchFilter struct {
    ChatbotID string
    SourceID  string
}

type DeleteFilter struct {
    SourceID string
}

type SearchResult struct {
    ID      interface{}
    Score   float64
    Payload VectorPayload
}
```

### Phase 2: Create Provider Factories

**File**: `internal/ai/factory.go`

```go
package ai

import (
    "fmt"
    "os"
)

type ProviderType string

const (
    ProviderOpenAI     ProviderType = "openai"
    ProviderOpenRouter ProviderType = "openrouter"
    ProviderQdrant     ProviderType = "qdrant"
    ProviderPGVector   ProviderType = "pgvector"
)

type Config struct {
    VectorProvider   ProviderType
    EmbedderProvider ProviderType
    LLMProvider      ProviderType
}

func ConfigFromEnv() Config {
    return Config{
        VectorProvider:   ProviderType(getEnvOrDefault("VECTOR_PROVIDER", "qdrant")),
        EmbedderProvider: ProviderType(getEnvOrDefault("EMBEDDER_PROVIDER", "openai")),
        LLMProvider:      ProviderType(getEnvOrDefault("LLM_PROVIDER", "openrouter")),
    }
}

func NewVectorStore(cfg Config) (VectorStore, error) {
    switch cfg.VectorProvider {
    case ProviderQdrant:
        return NewQdrantStore()
    case ProviderPGVector:
        return NewPGVectorStore()
    default:
        return nil, fmt.Errorf("unknown vector provider: %s", cfg.VectorProvider)
    }
}

func NewEmbedder(cfg Config) (Embedder, error) {
    switch cfg.EmbedderProvider {
    case ProviderOpenAI:
        return NewOpenAIEmbedder()
    default:
        return nil, fmt.Errorf("unknown embedder provider: %s", cfg.EmbedderProvider)
    }
}
```

### Phase 3: Refactor Existing Implementations

**File**: `internal/ai/qdrant/client.go`

```go
package qdrant

import (
    "github.com/onurceri/botla-co/internal/ai"
)

// Client implements ai.VectorStore
type Client struct {
    baseURL string
    apiKey  string
    http    *http.Client
}

// Verify interface compliance
var _ ai.VectorStore = (*Client)(nil)

func New(baseURL, apiKey string) *Client {
    return &Client{
        baseURL: baseURL,
        apiKey:  apiKey,
        http:    &http.Client{Timeout: defaultTimeout()},
    }
}
```

### Phase 4: Update Service Dependencies

**File**: `internal/services/chat_service.go`

```go
type ChatService struct {
    db          *sql.DB
    vectorStore ai.VectorStore   // Interface instead of concrete type
    embedder    ai.Embedder       // Interface instead of concrete type
    llm         ai.ChatCompleter  // Interface instead of concrete type
}

func NewChatService(db *sql.DB, vs ai.VectorStore, emb ai.Embedder, llm ai.ChatCompleter) *ChatService {
    return &ChatService{
        db:          db,
        vectorStore: vs,
        embedder:    emb,
        llm:         llm,
    }
}
```

## Directory Structure After Refactoring

```
internal/
├── ai/
│   ├── interfaces.go      # All provider interfaces
│   ├── factory.go         # Provider factory functions
│   ├── types.go           # Shared domain types
│   ├── qdrant/
│   │   ├── client.go      # VectorStore implementation
│   │   └── client_test.go
│   ├── openai/
│   │   ├── embedder.go    # Embedder implementation
│   │   ├── completer.go   # ChatCompleter implementation
│   │   └── *_test.go
│   └── openrouter/
│       ├── completer.go   # ChatCompleter implementation
│       └── completer_test.go
├── rag/                   # Keep for RAG-specific logic (chunking, search)
│   ├── chunker.go
│   ├── search.go
│   └── ...
```

## Migration Strategy

1. **Create `internal/ai/` package** with interfaces
2. **Add adapter layer** to existing implementations to satisfy new interfaces
3. **Gradually migrate** services to use interfaces
4. **Remove old direct dependencies** once all consumers use interfaces
5. **Reorganize** implementation files into sub-packages

## Affected Files

| File | Action | Description |
|------|--------|-------------|
| `internal/ai/interfaces.go` | NEW | Central interface definitions |
| `internal/ai/factory.go` | NEW | Provider factory functions |
| `internal/ai/qdrant/client.go` | NEW | Refactored Qdrant implementation |
| `internal/ai/openai/embedder.go` | NEW | OpenAI embedder implementation |
| `internal/rag/qdrant.go` | DEPRECATE | Mark as deprecated, keep for compatibility |
| `internal/services/chat_service.go` | MODIFY | Use interfaces for dependencies |
| `cmd/server/main.go` | MODIFY | Use factory to create providers |

## Testing Strategy

### Unit Tests with Mocks

```go
type MockVectorStore struct {
    SearchFunc func(ctx context.Context, vector []float32, filter ai.SearchFilter, limit int) ([]ai.SearchResult, error)
}

func (m *MockVectorStore) Search(ctx context.Context, vector []float32, filter ai.SearchFilter, limit int) ([]ai.SearchResult, error) {
    return m.SearchFunc(ctx, vector, filter, limit)
}

func TestRAGSearch_WithMockVectorStore(t *testing.T) {
    mock := &MockVectorStore{
        SearchFunc: func(...) ([]ai.SearchResult, error) {
            return []ai.SearchResult{{ID: "1", Score: 0.95}}, nil
        },
    }
    
    svc := NewSearchService(mock)
    results, err := svc.Search(ctx, "test query")
    
    assert.NoError(t, err)
    assert.Len(t, results, 1)
}
```

### Integration Tests

Existing integration tests should pass with provider env vars set.

## Acceptance Criteria

- [ ] All provider interfaces defined in `internal/ai/`
- [ ] Factory functions create providers based on configuration
- [ ] Services depend on interfaces, not implementations
- [ ] Mock implementations exist for all interfaces
- [ ] Existing tests pass without modification
- [ ] New unit tests achieve 90%+ coverage on new code

## Estimated Effort

**Size**: Large (4-5 days)
- Phase 1: 0.5 day
- Phase 2: 1 day
- Phase 3: 2 days
- Phase 4: 1 day
- Testing: 1 day

## Dependencies

- None for initial implementation
- PGVector implementation would require additional PostgreSQL extension setup

## Notes

This refactoring prepares the codebase for:
- Easy provider switching (OpenAI → Anthropic)
- Local development with mock providers
- Future PGVector implementation for cost optimization
- Better unit testing with dependency injection
