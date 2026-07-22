> Extracted from: background-processing.md — Section: 2.4 Background Processing — Text Processor

# 2.4 Background Processing — Text Processor

- [ ] Downloads text content from R2/S3 storage
- [ ] Empty content returns success with 0 chunks (no error)
- [ ] Content is normalized (Turkish character normalization)
- [ ] Ingestion metadata is extracted (capability summary, suggested questions)
- [ ] Text is chunked at 512 target tokens with ~15% overlap
- [ ] Chunks are embedded via OpenAI `text-embedding-3-small`
- [ ] Vectors are upserted to Qdrant
- [ ] Usage statistics are incremented
