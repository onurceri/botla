> Extracted from: rag-pipeline.md — Section: 2.1 Ingestion Pipeline

## 2.1 Ingestion Pipeline

- [ ] Source input (URL, sitemap, PDF, raw text) is validated
- [ ] Content is fetched (HTTP for URLs, fitz for PDF, direct for text)
- [ ] Content is cleaned and normalized (whitespace, encoding)
- [ ] Content is chunked at 512 tokens with 15% overlap
- [ ] Chunking is sentence-aware (does not split mid-sentence)
- [ ] Each chunk is assigned a deterministic point ID via `MakePointID(sourceID, index)`
- [ ] Chunks are embedded in batches of 25 using OpenAI `text-embedding-3-small`
- [ ] Embedding batches respect API rate limits
- [ ] Embeddings are upserted to Qdrant with metadata (sourceID, chunk index, text)
- [ ] Ingestion is retryable — duplicate chunks are idempotent
- [ ] Ingestion failures are reported to the source queue
