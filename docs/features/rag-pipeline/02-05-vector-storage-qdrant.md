> Extracted from: rag-pipeline.md — Section: 2.5 Vector Storage — Qdrant (`qdrant.go`)

## 2.5 Vector Storage — Qdrant (`qdrant.go`)

- [ ] HTTP client connects to configured Qdrant instance
- [ ] Collections are created/verified on startup
- [ ] Point IDs are deterministic (`MakePointID(sourceID, index)`)
- [ ] Metadata stored per point: `source_id`, `chunk_index`, `text`, `source_url`
- [ ] Upserts are idempotent (same point ID = update)
- [ ] Batch operations are used for performance
- [ ] Connection errors are retried with backoff
