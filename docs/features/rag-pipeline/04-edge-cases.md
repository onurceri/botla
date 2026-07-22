> Extracted from: rag-pipeline.md — Section: 4. Edge Cases

# 4. Edge Cases

## 4.1 RAG Pipeline Edge Cases

| # | Edge Case | Expected Behavior |
|---|---|---|
| 4.1.1 | Empty content ingested | Ingestion succeeds but produces no chunks; chatbot responds "No content available" |
| 4.1.2 | Content is all gibberish/meaningless text | Chunking proceeds; search results will be low quality; no crash |
| 4.1.3 | Very large source (1000+ pages) | Chunked and embedded in batches; ingestion may take time but completes |
| 4.1.4 | Source URL returns 403 Forbidden | Ingestion fails; error is reported; retry logic may apply |
| 4.1.5 | Source URL redirects multiple times | Follows redirects up to configured limit; SSRF check on final URL |
| 4.1.6 | Source URL is very slow (> 30s to respond) | Request times out; ingestion fails; error reported |
| 4.1.7 | PDF has embedded images with no text (scanned document) | No text extracted; error reported (OCR not supported without fitz/full CGO) |
| 4.1.8 | Qdrant is down during query | Error is returned; chatbot responds with "Service unavailable" |
| 4.1.9 | Qdrant is down during ingestion | Ingestion fails; queued for retry |
| 4.1.10 | OpenAI API key is invalid | All embedding/LLM calls fail; clear error message logged |
| 4.1.11 | OpenAI API is down | Circuit breaker opens; fast-fail for all LLM-dependent requests |
| 4.1.12 | Search query is empty string | Validation rejects; returns 400 Bad Request |
| 4.1.13 | Search query is very long (> 1000 tokens) | Truncated or returns error; no crash |
| 4.1.14 | Search returns 0 results (empty Qdrant collection) | Treated as low tier; fallback mode activated |
| 4.1.15 | All search results are from a single source | Deduplication works; results are diverse |
| 4.1.16 | Token count exceeds LLM context window | Context is truncated (oldest chunks removed first); response quality warning |
| 4.1.17 | Chunk contains only whitespace | Filtered out; not embedded |
| 4.1.18 | Chunk overlap contains the same sentence | Deduplication in context assembly prevents duplicate context |
| 4.1.19 | Circuit breaker transitions multiple times | Metrics track transitions; no infinite loops |
| 4.1.20 | Multiple concurrent queries to the same chatbot | Handled concurrently; no race conditions in session state |
