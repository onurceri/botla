> Extracted from: chat-operations.md — Section: 1.7 Chat Pipeline — RAG Integration

## 1.7 Chat Pipeline — RAG Integration

- [ ] User message is embedded via OpenAI `text-embedding-3-small`
- [ ] Embedding is searched against Qdrant with tiered similarity thresholds:
  - [ ] High relevance: score ≥ 0.50
  - [ ] Medium relevance: score 0.30–0.49
  - [ ] Low relevance: score < 0.30
- [ ] Retrieved chunks are assembled into the LLM context window
- [ ] Context window respects `RAG_MAX_CONTEXT_TOKENS` env var
- [ ] Top-K results respect `RAG_TOPK` env var
- [ ] LLM generation uses OpenAI (GPT-4o / GPT-4o-mini) or OpenRouter provider
- [ ] Circuit breaker protects LLM calls: 50% failure threshold, 30s timeout
- [ ] Response is streamed or returned as JSON depending on client preference
