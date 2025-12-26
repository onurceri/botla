# AI/Vector Provider Abstraction

This package provides a clean abstraction layer for AI and vector database providers, enabling easy swapping of implementations and better testing.

## Architecture

The package follows a registry-based factory pattern where providers register themselves during initialization:

```
internal/ai/
├── interfaces.go      # Core interfaces (VectorStore, Embedder)
├── types.go           # Shared domain types
├── factory.go         # Provider registry and factory functions
├── mocks.go           # Mock implementations for testing
├── qdrant/
│   └── client.go      # Qdrant VectorStore implementation
├── openai/
│   └── embedder.go    # OpenAI Embedder implementation
└── openrouter/
    └── embedder.go    # OpenRouter Embedder implementation
```

## Core Interfaces

### VectorStore

Abstracts vector database operations:

```go
type VectorStore interface {
    EnsureCollection(ctx context.Context) error
    Upsert(ctx context.Context, id interface{}, vector []float32, payload VectorPayload) error
    Search(ctx context.Context, vector []float32, filter SearchFilter, limit int) ([]SearchResult, error)
    Delete(ctx context.Context, filter DeleteFilter) error
    Scroll(ctx context.Context, filter SearchFilter, limit int, offset interface{}) ([]SearchResult, interface{}, error)
}
```

### Embedder

Abstracts text embedding generation:

```go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
    Dimension() int
}
```

## Usage

### Creating Providers

Use the factory functions with configuration:

```go
import (
    "github.com/onurceri/botla-co/internal/ai"
    _ "github.com/onurceri/botla-co/internal/ai/qdrant"  // Register Qdrant
    _ "github.com/onurceri/botla-co/internal/ai/openai"  // Register OpenAI
)

func main() {
    cfg := ai.ConfigFromEnv()
    
    // Create vector store (default: Qdrant)
    vectorStore, err := ai.NewVectorStore(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create embedder (default: OpenAI)
    embedder, err := ai.NewEmbedder(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the providers
    ctx := context.Background()
    embedding, err := embedder.Embed(ctx, "Hello, world!")
    // ...
}
```

### Environment Configuration

Configure providers via environment variables:

- `VECTOR_PROVIDER` - Vector database provider (default: "qdrant")
  - `qdrant` - Qdrant vector database
  - `pgvector` - PostgreSQL pgvector (not yet implemented)

- `EMBEDDER_PROVIDER` - Embedding provider (default: "openai")
  - `openai` - OpenAI embeddings
  - `openrouter` - OpenRouter embeddings

- `LLM_PROVIDER` - Chat completion provider (default: "openrouter")
  - `openai` - OpenAI chat completions
  - `openrouter` - OpenRouter chat completions

### Provider-Specific Configuration

**Qdrant:**
- `QDRANT_URL` - Qdrant server URL (required)
- `QDRANT_API_KEY` - Qdrant API key (optional)
- `QDRANT_TIMEOUT_MS` - Request timeout in milliseconds (default: 15000)

**OpenAI:**
- `OPENAI_API_KEY` - OpenAI API key (required)
- `OPENAI_API_BASE` - API base URL (default: https://api.openai.com)
- `OPENAI_EMBEDDING_MODEL` - Embedding model (default: text-embedding-3-small)
- `OPENAI_TIMEOUT_MS` - Request timeout in milliseconds (default: 30000)

**OpenRouter:**
- `OPENROUTER_API_KEY` - OpenRouter API key (required, falls back to OPENAI_API_KEY)
- `OPENROUTER_API_BASE` - API base URL (default: https://openrouter.ai/api/v1)
- `OPENROUTER_EMBEDDING_MODEL` - Embedding model (default: text-embedding-3-small)
- `OPENROUTER_TIMEOUT_MS` - Request timeout in milliseconds (default: 30000)

## Testing

### Using Mock Implementations

The package provides mock implementations for testing:

```go
import "github.com/onurceri/botla-co/internal/ai"

func TestMyService(t *testing.T) {
    mockEmbedder := &ai.MockEmbedder{
        EmbedFunc: func(ctx context.Context, text string) ([]float32, error) {
            return []float32{0.1, 0.2, 0.3}, nil
        },
    }
    
    mockVectorStore := &ai.MockVectorStore{
        SearchFunc: func(ctx context.Context, vector []float32, filter ai.SearchFilter, limit int) ([]ai.SearchResult, error) {
            return []ai.SearchResult{
                {
                    ID:    "result-1",
                    Score: 0.95,
                    Payload: ai.VectorPayload{
                        ChatbotID: "bot-123",
                        SourceID:  "src-456",
                        // ...
                    },
                },
            }, nil
        },
    }
    
    // Use mocks in your service
    service := NewMyService(mockVectorStore, mockEmbedder)
    // ... test your service
}
```

### Registering Custom Factories

For testing with specific provider implementations:

```go
func TestWithCustomProvider(t *testing.T) {
    ai.RegisterEmbedder(ai.ProviderOpenAI, func() (ai.Embedder, error) {
        return &ai.MockEmbedder{}, nil
    })
    defer delete(ai.embedderFactories, ai.ProviderOpenAI)
    
    cfg := ai.Config{EmbedderProvider: ai.ProviderOpenAI}
    embedder, err := ai.NewEmbedder(cfg)
    // ...
}
```

## Adding New Providers

To add a new provider:

1. **Create the implementation** in a new sub-package (e.g., `internal/ai/newprovider/`)
2. **Implement the interface** (e.g., `VectorStore` or `Embedder`)
3. **Register the factory** in an `init()` function:

```go
package newprovider

import "github.com/onurceri/botla-co/internal/ai"

func init() {
    ai.RegisterVectorStore(ai.ProviderType("newprovider"), func() (ai.VectorStore, error) {
        return NewFromEnv()
    })
}

type Client struct {
    // ...
}

var _ ai.VectorStore = (*Client)(nil)

func NewFromEnv() (*Client, error) {
    // Initialize from environment variables
}

// Implement all VectorStore methods...
```

4. **Add the provider constant** to `factory.go`
5. **Import the package** where needed to trigger registration

## Design Rationale

### Registry Pattern

We use a registry-based factory pattern to avoid import cycles:
- Sub-packages import the `ai` package for interfaces
- Sub-packages register themselves via `init()` functions  
- Main package imports sub-packages to trigger registration
- No circular dependencies

### Environment-Based Configuration

Providers are configured via environment variables to:
- Avoid loading the full config package (which may have dependencies)
- Enable easy testing without complex setup
- Follow 12-factor app principles

### Mock-Friendly Design

All interfaces use function injection in mocks:
- Easy to customize behavior per test
- No need for complex mock frameworks
- Tests remain readable and maintainable

## Coverage

All provider implementations have 90%+ test coverage:
- Unit tests with mock HTTP servers
- Interface compliance tests
- Error handling tests
- Edge case tests

Run tests with:
```bash
go test -v ./internal/ai/...
```
