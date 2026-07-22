> Extracted from: background-processing.md — Section: 2.7 Background Processing — Worker Pool

# 2.7 Background Processing — Worker Pool

- [ ] Max 16 concurrent workers (capped)
- [ ] Queue buffer of 64 jobs
- [ ] 500ms delay between jobs in non-test environments (rate limiting)
- [ ] Graceful shutdown via `Stop()` — waits for all workers to finish
- [ ] Worker count and queue length are observable via `WorkerCount()` and `QueueLength()`
