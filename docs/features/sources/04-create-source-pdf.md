> Extracted from: sources.md — Section: 2.3 Source Management — Create Source (PDF)

# 2.3 Source Management — Create Source (PDF)

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=pdf` and a `file` multipart upload creates a new PDF source
- [ ] Content-Type validation rejects non-PDF files (checks `application/pdf` MIME type and `.pdf` extension)
- [ ] File size limit check returns 413 Request Entity Too Large
- [ ] File count per-bot limit check returns 403 Forbidden with `ERR_PDF_LIMIT_REACHED`
- [ ] Total storage quota check returns 402 Payment Required
- [ ] SHA-256 content hash duplicate detection returns 409 Conflict
- [ ] File is uploaded to R2/S3 storage before source record is created
- [ ] Source is created with `status: "pending"` and immediately enqueued for processing
- [ ] Response returns 201 Created with the new source ID
