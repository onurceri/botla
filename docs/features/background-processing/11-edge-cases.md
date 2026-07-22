> Extracted from: background-processing.md — Section: 5. Edge Cases

# 5. Edge Cases

| # | Edge Case | Expected Behavior |
|---|-----------|-------------------|
| 18 | Qdrant unavailable during processing | Job fails at `embed_chunks` step |
| 19 | OpenAI API error during embedding | Job fails at `embed_chunks` step |
| 20 | OpenAI API error during metadata extraction | Processing continues (non-fatal), logged as warning |
| 21 | Invalid PDF file (corrupted) | Job fails at `parse_content` with `ERR_PDF_PARSE_FAILED` |
| 22 | Password-protected PDF | Job fails at `parse_content` with `ERR_PDF_PARSE_FAILED` |
| 23 | Scanned PDF without OCR (Free plan) | Job fails at `parse_content` with `ERR_EMPTY_CONTENT` |
| 24 | Scanned PDF with OCR (Pro+ plan) | Text extracted successfully |
| 25 | URL with JS-rendered content (dynamic disabled) | Job fails with `ERR_DYNAMIC_REQUIRED` |
| 26 | URL with JS-rendered content (dynamic enabled) | Content extracted via go-rod |
| 27 | URL timeout (slow server) | Job fails with `ERR_SCRAPE_TIMEOUT` |
| 28 | URL returns 403 | Job fails with `ERR_SCRAPE_FORBIDDEN` |
| 29 | URL network error | Job fails with `ERR_SCRAPE_NETWORK` |
| 30 | URL empty content (blank page) | Job fails with `ERR_EMPTY_CONTENT` |
| 31 | URL with no internal links | Discovery finds 0 links, processing completes |
| 32 | URL discovery HTML fetch fails | Processing continues, no sub-pages discovered |
| 33 | Queue full | Job marked as failed with `QUEUE_FULL` |
| 34 | Max retries exceeded (3) | Job marked as failed, source marked as failed |
| 35 | Retryable error (rate limit, network) | Job retried with exponential backoff (2s, 4s, 8s) |
| 36 | Non-retryable error (not found, cancelled) | Job fails immediately, no retry |
| 37 | Stuck job recovery at startup | Jobs in `pending`/`running` status are re-enqueued |
| 44 | Suggestion re-aggregation panic | Panic recovered, logged, server continues |
| 55 | Discovery with invalid path filter regex | Discovery continues without filter |
| 56 | Discovery with single-URL plan | Discovery skipped (logged) |
| 57 | Discovery with zero crawl limit | Discovery skipped (logged) |
| 67 | Suggestion aggregation with no sources | Empty suggestions list |
| 68 | Suggestion aggregation with all sources deleted | Empty suggestions list |
| 69 | Concurrent suggestion re-aggregation | No duplicate or conflicting suggestions |
| 70 | Embedding service not available (nil) | Job fails with `ERR_LLM_NOT_SUPPORTED` |
