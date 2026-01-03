# AGENTS.md - internal/rag

RAG pipeline implementation with OpenAI embeddings, Qdrant vector storage, tiered search, and circuit breaker protection.

## WHERE TO LOOK

| File | Purpose |
|------|---------|
| `subsystem.go` | RAGSubsystem interface, core embedding/search/store/complete operations |
| `embedding.go` | EmbeddingService for batch embedding creation and Qdrant upsert (25-chunk batches, rate limited ~58/sec) |
| `chunker.go` | ChunkText function: token-aware text segmentation with 15% tail overlap, sentence/paragraph boundaries |
| `circuit_breaker.go` | CircuitBreakerClient/Manager wrapping LLMClient with gobreaker, failure ratio 0.5, 30s timeout |
| `topic_extractor.go` | ExtractIngestionMetadata: capability summaries, suggested questions, ALWAYS ENGLISH prompts |
| `search.go` | SearchContextTiered: tiered similarity search (high/medium/low), threshold-based context filtering |
| `qdrant.go` | QdrantClient: HTTP-based vector operations, collection ensure, upsert, search, scroll, delete |
| `openai.go` | OpenAIClient: embeddings (text-embedding-3-small), chat completions with 4x retry, exponential backoff |

## CONVENTIONS

- **Interface-driven**: EmbeddingClient, VectorClient, LLMClient define contracts; implementations swappable
- **Rate limiting**: EmbeddingService uses ~58 req/sec ticker, 25-chunk batches
- **Retry logic**: OpenAI client retries 4x with exponential backoff (200ms * 2^attempt)
- **Circuit breaker**: Default settings: 3 max requests, 60s interval, 30s timeout, 0.5 failure ratio, 5 min requests
- **Chunk IDs**: Deterministic SHA256-based UUID generation via MakePointID(sourceID, index)
- **Context tiering**: High (≥HighThreshold), Medium (≥MediumThreshold), Low (below)
- **Prompt language**: topic_extractor prompts ALWAYS in English; output language controlled via directive

## ANTI-PATTERNS

- **DON'T bypass circuit breaker**: Always route LLM calls through CircuitBreakerClient
- **DON'T ignore tiered search**: Use SearchContextTiered for confidence-aware context selection
- **DON'T use large prompts for extraction**: topic_extractor truncates to 4000 chars max
- **DON'T forget rate limiting**: EmbeddingService batching and tickers prevent API throttling
- **DON'T hardcode thresholds**: SearchContextTiered reads RAG_TOPK/RAG_MAX_CONTEXT_TOKENS from env
- **DON'T skip null checks**: ragSubsystem methods return typed errors (ErrNilEmbedder, etc.)
