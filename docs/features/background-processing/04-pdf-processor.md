> Extracted from: background-processing.md — Section: 2.3 Background Processing — PDF Processor

# 2.3 Background Processing — PDF Processor

- [ ] Downloads PDF file from R2/S3 storage to a temp file
- [ ] Extracts text via `pdf.ExtractPDFText` (fitz CGO, OCR support for Pro+ plans)
- [ ] Empty content after extraction returns error `ERR_EMPTY_CONTENT`
- [ ] Content is normalized (Turkish character normalization)
- [ ] Ingestion metadata is extracted (capability summary, suggested questions)
- [ ] Text is chunked at 512 target tokens with ~15% overlap
- [ ] Chunks are embedded via OpenAI `text-embedding-3-small` (25-chunk batches, rate limited)
- [ ] Vectors are upserted to Qdrant with deterministic point IDs
- [ ] Usage statistics are incremented
- [ ] Temp file is cleaned up after processing
