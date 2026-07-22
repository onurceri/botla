> Extracted from: rag-pipeline.md — Section: 2.6 Search (`search.go`) — Tiered Similarity Search

## 2.6 Search (`search.go`) — Tiered Similarity Search

- [ ] Query is embedded using the same embedding model
- [ ] Search is performed against the relevant Qdrant collection
- [ ] Results are classified into tiers by score:
  - [ ] **High tier:** score ≥ 0.50 — strong match, returned as-is
  - [ ] **Medium tier:** score 0.30–0.49 — weak match, returned with confidence warning
  - [ ] **Low tier:** score < 0.30 — no good match, triggers fallback
- [ ] `RAG_TOPK` environment variable controls number of results (default: 5)
- [ ] `RAG_MAX_CONTEXT_TOKENS` limits total context window
- [ ] Results are deduplicated by source
- [ ] Metadata (source URL, title) is preserved for citation
