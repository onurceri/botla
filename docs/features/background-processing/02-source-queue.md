> Extracted from: background-processing.md — Section: 2.1 Background Processing — Source Queue

# 2.1 Background Processing — Source Queue

- [ ] Bounded worker pool processes sources asynchronously (max 16 workers, 64-job buffer)
- [ ] Each source type has a dedicated processor: URLProcessor, PDFProcessor, TextProcessor
- [ ] Processing pipeline: fetch_source → parse_content → chunk_text → embed_chunks → store_vectors
- [ ] Step progress is tracked: 10% → 30% → 50% → 80% → 100%
- [ ] Jobs are retried up to 3 times with exponential backoff (2s, 4s, 8s)
- [ ] Retryable errors: rate limits, network errors, timeouts, 429/502/503 responses
- [ ] Non-retryable errors: context cancelled, not found, invalid source type
- [ ] Stuck jobs (pending/running) are recovered at startup
- [ ] Queue full returns error; job is marked as failed with `QUEUE_FULL` error code
- [ ] 500ms delay between jobs in non-test environments (rate limiting)
