> Extracted from: rag-pipeline.md — Section: 2.3 Embedding (`embedding.go`)

## 2.3 Embedding (`embedding.go`)

- [ ] Uses OpenAI `text-embedding-3-small` model
- [ ] Batches up to 25 chunks per API call
- [ ] Rate-limited to prevent OpenAI 429 errors
- [ ] Handles empty text gracefully (returns zero vector)
- [ ] Caches embeddings for duplicate chunks (within same batch)
