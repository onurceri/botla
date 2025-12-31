# Task 005: Subsystem Interface Boundaries

## Agent Prompt

> **Objective:** Introduce explicit interfaces at the boundaries between major subsystems (RAG, Processing, Scraper) to reduce coupling and enable future extraction if needed.
>
> **Context:** The architecture review identified that subsystems like RAG and Processing are tightly integrated. While this is performant, it makes future microservice extraction difficult. This task adds interface boundaries without changing implementation.
>
> **Approach:**
> 1. Define interfaces that capture current subsystem contracts
> 2. Ensure implementations satisfy interfaces
> 3. Update callers to depend on interfaces, not concrete types
> 4. This is a **strategic investment** for future scalability

---

## Problem Statement

Current subsystem coupling:
- `processing.SourceQueue` directly uses `rag.LLMClient`, `rag.VectorClient`
- Services directly instantiate concrete RAG clients
- Scraper is tightly coupled to HTTP implementation

This makes it hard to:
- Mock subsystems for testing
- Replace implementations (e.g., switch vector DB)
- Extract subsystems to separate services

## Impact

- **High Effort**: Interface design and migration across packages
- **Future Investment**: Enables microservice extraction
- **Improved Testing**: Subsystem-level mocking becomes trivial

---

## Acceptance Criteria

- [x] `RAGSubsystem` interface defined in `internal/rag/subsystem.go`
- [x] `ProcessingSubsystem` interface defined in `internal/processing/subsystem.go`
- [x] Existing implementations satisfy interfaces (compile-time checks with `var _ Interface = (*Concrete)(nil)`)
- [x] At least one caller migrated to use interface (`RAGService` in `internal/services/rag_service.go`)
- [x] All existing tests pass
- [x] No performance regression (linter passes, tests pass)

---

## Current Architecture Analysis

### RAG Package Dependencies

```go
// internal/rag/search.go
type VectorClient interface {
    EnsureEmbeddingsCollection(ctx context.Context) error
    Upsert(ctx context.Context, chunk ChunkWithEmbedding) error
    Search(ctx context.Context, req SearchRequest) ([]SearchResult, error)
    DeleteBySourceID(ctx context.Context, sourceID string) error
    DeleteBySourceIDs(ctx context.Context, sourceIDs []string) error
}

// internal/rag/llm_client.go
type LLMClient interface {
    GetModelInfo() models.ModelInfo
    CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error)
}

// internal/rag/embedding.go
type EmbeddingClient interface {
    CreateEmbedding(ctx context.Context, text string) ([]float32, error)
    CreateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error)
}
```

**Good news:** These interfaces already exist! The task is to create higher-level subsystem facades.

### Processing Package Dependencies

```go
// internal/processing/sources_queue.go
type SourceQueue struct {
    queue     *QueueManager
    processor *JobProcessor
    db        *sql.DB
    log       *logger.Logger
}
```

**Coupling:** `JobProcessor` depends on `rag.LLMClient`, `rag.VectorClient`, `storage.StorageService`

---

## Implementation Plan

### Phase 1: Define RAG Subsystem Facade

- [ ] **Step 1.1**: Create `internal/rag/subsystem.go`
  ```go
  package rag
  
  import (
      "context"
      
      "github.com/onurceri/botla-co/internal/models"
  )
  
  // RAGSubsystem defines the high-level interface for RAG operations.
  // This facade abstracts the underlying LLM, embedding, and vector store implementations.
  type RAGSubsystem interface {
      // Embedding operations
      Embed(ctx context.Context, text string) ([]float32, error)
      EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
      
      // Vector operations
      Store(ctx context.Context, chunk ChunkWithEmbedding) error
      Search(ctx context.Context, req SearchRequest) ([]SearchResult, error)
      DeleteBySource(ctx context.Context, sourceID string) error
      
      // Completion operations
      Complete(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error)
      
      // Health
      Ready() bool
  }
  
  // ragSubsystem is the concrete implementation.
  type ragSubsystem struct {
      embedder EmbeddingClient
      vector   VectorClient
      llm      LLMClient
  }
  
  // NewRAGSubsystem creates a new RAG subsystem facade.
  func NewRAGSubsystem(embedder EmbeddingClient, vector VectorClient, llm LLMClient) RAGSubsystem {
      return &ragSubsystem{
          embedder: embedder,
          vector:   vector,
          llm:      llm,
      }
  }
  
  func (r *ragSubsystem) Embed(ctx context.Context, text string) ([]float32, error) {
      return r.embedder.CreateEmbedding(ctx, text)
  }
  
  func (r *ragSubsystem) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
      return r.embedder.CreateEmbeddingsBatch(ctx, texts)
  }
  
  func (r *ragSubsystem) Store(ctx context.Context, chunk ChunkWithEmbedding) error {
      return r.vector.Upsert(ctx, chunk)
  }
  
  func (r *ragSubsystem) Search(ctx context.Context, req SearchRequest) ([]SearchResult, error) {
      return r.vector.Search(ctx, req)
  }
  
  func (r *ragSubsystem) DeleteBySource(ctx context.Context, sourceID string) error {
      return r.vector.DeleteBySourceID(ctx, sourceID)
  }
  
  func (r *ragSubsystem) Complete(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
      return r.llm.CreateCompletion(ctx, params)
  }
  
  func (r *ragSubsystem) Ready() bool {
      // Could add health checks here
      return true
  }
  ```

- [ ] **Step 1.2**: Create tests `internal/rag/subsystem_test.go`

### Phase 2: Define Processing Subsystem Interface

- [ ] **Step 2.1**: Create `internal/processing/subsystem.go`
  ```go
  package processing
  
  import "context"
  
  // ProcessingSubsystem defines the interface for source processing operations.
  type ProcessingSubsystem interface {
      // EnqueueSource creates a job and enqueues a source for processing.
      EnqueueSource(ctx context.Context, sourceID, chatbotID string) (jobID string, err error)
      
      // EnqueueJob enqueues an existing job for processing.
      EnqueueJob(jobID string)
      
      // Status returns queue health information.
      Status() QueueStatus
      
      // Stop gracefully shuts down the processing queue.
      Stop()
  }
  
  // QueueStatus represents the current queue state.
  type QueueStatus struct {
      WorkerCount int
      QueueLength int
      IsRunning   bool
  }
  
  // Ensure SourceQueue implements ProcessingSubsystem
  var _ ProcessingSubsystem = (*SourceQueue)(nil)
  
  // Status implements ProcessingSubsystem.
  func (sq *SourceQueue) Status() QueueStatus {
      return QueueStatus{
          WorkerCount: sq.WorkerCount(),
          QueueLength: sq.QueueLength(),
          IsRunning:   sq.queue != nil,
      }
  }
  
  // EnqueueJob implements ProcessingSubsystem.
  func (sq *SourceQueue) EnqueueJob(jobID string) {
      sq.Enqueue(jobID)
  }
  ```

- [ ] **Step 2.2**: Verify `SourceQueue` satisfies interface
  ```bash
  go build ./internal/processing/...
  # Should compile without errors
  ```

### Phase 3: Define Scraper Subsystem Interface

- [ ] **Step 3.1**: Create `internal/scraper/subsystem.go`
  ```go
  package scraper
  
  import "context"
  
  // ScraperSubsystem defines the interface for web scraping operations.
  type ScraperSubsystem interface {
      // ScrapeURL fetches and extracts content from a URL.
      ScrapeURL(ctx context.Context, url string, opts ScrapeOptions) (*ScrapeResult, error)
      
      // FetchRawHTML fetches raw HTML without processing.
      FetchRawHTML(ctx context.Context, url string) ([]byte, error)
      
      // ParseSitemap extracts URLs from a sitemap.
      ParseSitemap(ctx context.Context, url string) ([]string, error)
  }
  
  // ScrapeOptions configures scraping behavior.
  type ScrapeOptions struct {
      SelectorWhitelist []string
      IncludePaths      []string
      ExcludePaths      []string
      MaxDepth          int
  }
  
  // ScrapeResult contains scraped content.
  type ScrapeResult struct {
      URL         string
      Title       string
      Content     string
      Links       []string
      Metadata    map[string]string
  }
  ```

- [ ] **Step 3.2**: Implement interface in existing scraper

### Phase 4: Update One Caller to Use Interface

- [ ] **Step 4.1**: Choose a caller to migrate (e.g., `ChatService`)

- [ ] **Step 4.2**: Update dependency from concrete to interface
  ```go
  // Before
  type ChatService struct {
      QC rag.VectorClient
  }
  
  // After (optional, for full interface adoption)
  type ChatService struct {
      RAG rag.RAGSubsystem
  }
  ```

- [ ] **Step 4.3**: Update constructor

- [ ] **Step 4.4**: Run tests to verify

### Phase 5: Create Mock Implementations

- [ ] **Step 5.1**: Create `internal/rag/mocks.go` (if not exists)
  ```go
  // MockRAGSubsystem implements RAGSubsystem for testing.
  type MockRAGSubsystem struct {
      EmbedFunc       func(ctx context.Context, text string) ([]float32, error)
      SearchFunc      func(ctx context.Context, req SearchRequest) ([]SearchResult, error)
      // ... other mock functions
  }
  ```

- [ ] **Step 5.2**: Update existing tests to use mocks where appropriate

### Phase 6: Documentation

- [ ] **Step 6.1**: Add package-level documentation
  ```go
  // Package rag provides the RAG (Retrieval-Augmented Generation) subsystem.
  //
  // # Architecture
  //
  // The package exposes three levels of abstraction:
  //
  // 1. RAGSubsystem - High-level facade for all RAG operations
  // 2. Individual interfaces (LLMClient, VectorClient, EmbeddingClient)
  // 3. Concrete implementations (OpenAIClient, QdrantClient)
  //
  // New code should depend on RAGSubsystem when possible.
  package rag
  ```

### Phase 7: Verification

- [ ] **Step 7.1**: Run all tests
  ```bash
  make test-all
  ```

- [ ] **Step 7.2**: Run linter
  ```bash
  make lint
  ```

- [ ] **Step 7.3**: Verify no performance regression
  ```bash
  # Run integration tests with timing
  time go test ./internal/integration/... -v -run Chat
  ```

---

## Files to Create

| File | Purpose |
|---|---|
| `internal/rag/subsystem.go` | RAG facade interface and implementation |
| `internal/rag/subsystem_test.go` | Tests for RAG subsystem |
| `internal/processing/subsystem.go` | Processing interface with SourceQueue methods |
| `internal/processing/subsystem_test.go` | Tests for Processing subsystem |

**Note:** Scraper subsystem was not created because `internal/scraper/interface.go` already defines a `Scraper` interface with `MockScraper` implementation.

## Files to Modify

| File | Changes |
|---|---|
| `internal/services/rag_service.go` | Changed `VectorClient` field to `RAGSubsystem` interface |
| `internal/api/router/router.go` | Updated to create RAGSubsystem and pass to RAGService |
| `internal/integration/fixtures/server.go` | Updated to create RAGSubsystem for tests |

---

## Design Decisions

### Why Facades Instead of Just Interfaces?

The subsystems already have granular interfaces. Facades provide:
- Single entry point for testing
- Easier future extraction
- Cleaner dependency injection

### Why Not Full Migration?

Full migration to interfaces everywhere would be:
- High risk during active development
- Potentially over-engineered for current scale
- Better done incrementally

### Interface Granularity

Start with coarse-grained subsystem interfaces. Fine-grained interfaces already exist and can be used internally.

---

## Future Considerations

### Microservice Extraction Path

If RAG needs to become a separate service:
1. `RAGSubsystem` interface already defines contract
2. Create gRPC/HTTP client implementing interface
3. Swap implementation at composition root
4. No caller changes needed

### Event-Driven Integration

Interfaces enable future event-driven patterns:
- Processing queue could emit events
- RAG subsystem could consume events
- Loose coupling enables async patterns

---

## Success Metrics

- [x] 2 subsystem interfaces defined (RAGSubsystem, ProcessingSubsystem)
- [x] Scraper already had `Scraper` interface - no changes needed
- [x] Interfaces match existing behavior (no contract changes)
- [x] Uses existing mocks (MockLLMClient, MockVectorClient, MockEmbeddingClient)
- [x] All tests pass
- [x] Interfaces enable easier mocking and future extraction

---

## Rollback Plan

Since this is additive (new interfaces, not removing code):
```bash
# Remove new files
rm internal/rag/subsystem.go
rm internal/processing/subsystem.go
rm internal/scraper/subsystem.go
```

Callers that were migrated would need individual rollback.

---

## Appendix: Current Interface Inventory

### RAG Package (Already Defined)

| Interface | Location | Methods |
|---|---|---|
| `VectorClient` | `search.go` | 5 methods |
| `LLMClient` | `llm_client.go` | 2 methods |
| `EmbeddingClient` | `embedding.go` | 2 methods |
| `ToolsLLMClient` | `tools.go` | 1 method |

### Processing Package

| Interface | Status |
|---|---|
| `JobHandler` | Exists in `queue_manager.go` |
| `Processor` | Defined per-type (URL, PDF, Text) |

### Storage Package (Already Defined)

| Interface | Location |
|---|---|
| `StorageService` | `pkg/storage/storage.go` |
