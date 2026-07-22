> Extracted from: sources.md — Section: 2.4 Source Management — Create Source (Text)

# 2.4 Source Management — Create Source (Text)

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=text` and `text` form field creates a new text source
- [ ] Text length limit check returns 413 Request Entity Too Large with `ERR_TEXT_TOO_LONG`
- [ ] Storage quota check returns 402 Payment Required
- [ ] SHA-256 content hash duplicate detection returns 409 Conflict
- [ ] Text is uploaded to R2/S3 storage before source record is created
- [ ] Source is created with `status: "pending"` and immediately enqueued
