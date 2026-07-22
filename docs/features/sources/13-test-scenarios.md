> Extracted from: sources.md — Section: 4. Test Scenarios

# 4. Test Scenarios

### 4.1 URL Source — Happy Path

- [ ] Create a chatbot with a valid API key
- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=url` and `source_url=https://example.com`
- [ ] Verify response is 201 Created with a source ID
- [ ] `GET /api/v1/chatbots/{id}/sources` shows the new source with `status: "pending"`
- [ ] Wait for processing to complete (poll `GET /api/v1/sources/{id}`)
- [ ] Verify source status transitions: `pending` → `processing` → `completed`
- [ ] Verify `chunk_count` > 0
- [ ] Verify `processed_at` is set
- [ ] `GET /api/v1/sources/{id}/chunks` returns chunks with text content
- [ ] Verify each chunk has: id, chatbot_id, source_id, chunk_index, original_text, source_type, created_at
- [ ] Verify chatbot's `suggested_questions` is populated (via `GET /api/v1/chatbots/{id}`)

### 4.2 PDF Source — Happy Path

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=pdf` and a valid PDF file (multipart upload)
- [ ] Verify response is 201 Created with a source ID
- [ ] Poll `GET /api/v1/sources/{id}/job` until status is `completed`
- [ ] Verify `chunk_count` > 0
- [ ] Verify `GET /api/v1/sources/{id}/chunks` returns extracted text chunks
- [ ] Verify chatbot's `suggested_questions` includes questions derived from the PDF content

### 4.3 Text Source — Happy Path

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=text` and `text` form field containing a paragraph
- [ ] Verify 201 Created with source ID
- [ ] Poll job until `completed`
- [ ] Verify `chunk_count` > 0
- [ ] Verify chunks contain the original text split at sentence/paragraph boundaries

### 4.4 Sitemap Discovery — Happy Path

- [ ] `POST /api/v1/chatbots/{id}/sitemap/discover` with `{"sitemap_url": "https://example.com/sitemap.xml"}`
- [ ] Verify 200 OK with `urls` array and `total_count`
- [ ] Verify each URL has `loc`, `lastmod`, `changefreq` fields
- [ ] Configure chatbot with `include_paths: ["/blog/"]` and repeat
- [ ] Verify only URLs matching `/blog/` are returned
- [ ] Configure chatbot with `exclude_paths: ["/admin/"]` and repeat
- [ ] Verify `/admin/` URLs are excluded

### 4.5 Bulk Create — Happy Path

- [ ] `POST /api/v1/chatbots/{id}/sources/bulk` with `{"urls": ["https://example.com/page1", "https://example.com/page2"]}`
- [ ] Verify 201 Created with `created_count: 2`
- [ ] `GET /api/v1/chatbots/{id}/sources` shows both sources with `status: "pending"`
- [ ] Both sources are processed independently

### 4.6 Sitemap → Bulk Create → Process Flow

- [ ] `POST /api/v1/chatbots/{id}/sitemap/discover` with a valid sitemap URL
- [ ] Note the discovered URLs
- [ ] `POST /api/v1/chatbots/{id}/sources/bulk` with the discovered URLs
- [ ] Verify all sources are created and enqueued
- [ ] Poll until all sources reach `completed` status
- [ ] Verify chatbot answers questions based on the ingested content

### 4.7 URL Discovery — Auto Mode

- [ ] Create a chatbot with `discovery_mode: "auto"`
- [ ] Add a URL source that links to other pages on the same domain
- [ ] After processing, verify sub-pages were auto-created as sources with `is_discovered: true`
- [ ] Verify discovered sources are also processed (reach `completed` status)
- [ ] Verify discovered sources have `is_discovered: true` and won't crawl further

### 4.8 Refresh Source — Content Changed

- [ ] Create a URL source and wait for completion
- [ ] Note the `chunk_count` and `processed_at`
- [ ] `POST /api/v1/sources/{id}/refresh`
- [ ] Verify 202 Accepted
- [ ] Poll job until `completed`
- [ ] Verify `processed_at` is updated
- [ ] Verify chunks are updated (if content changed on the remote page)

### 4.9 Refresh Source — Content Unchanged

- [ ] Create a URL source and wait for completion
- [ ] Note the `chunk_count` and `hash`
- [ ] `POST /api/v1/sources/{id}/refresh`
- [ ] Poll job until `completed`
- [ ] Verify source status is `completed` with same `chunk_count` (content unchanged → skipped)
- [ ] Verify `processed_at` is still updated (refresh was performed)

### 4.10 Refresh Source — Cooldown

- [ ] Create a URL source and wait for completion
- [ ] `POST /api/v1/sources/{id}/refresh` — first refresh succeeds (202)
- [ ] Immediately `POST /api/v1/sources/{id}/refresh` again
- [ ] Verify 429 Too Many Requests with `Retry-After` header
- [ ] Wait for cooldown period, then refresh again — should succeed

### 4.11 Source Deletion

- [ ] Create a source and wait for completion
- [ ] `DELETE /api/v1/sources/{id}`
- [ ] Verify 204 No Content
- [ ] `GET /api/v1/sources/{id}` returns 404
- [ ] `GET /api/v1/chatbots/{id}/sources` does not include the deleted source
- [ ] Verify vectors are removed from Qdrant (chunks endpoint returns empty)
- [ ] Verify file is removed from R2/S3 storage (if PDF or text source)

### 4.12 Source Chunks — Pagination

- [ ] Create a source with enough content to produce 50+ chunks
- [ ] `GET /api/v1/sources/{id}/chunks` — verify 20 chunks returned with `next_cursor`
- [ ] `GET /api/v1/sources/{id}/chunks?offset={next_cursor}` — verify next 20 chunks
- [ ] Continue until all chunks are retrieved
- [ ] Verify total chunks match `chunk_count` on the source

### 4.13 Source Chunks — ETag Caching

- [ ] `GET /api/v1/sources/{id}` — note the `ETag` header
- [ ] `GET /api/v1/sources/{id}` with `If-None-Match: {etag}` — verify 304 Not Modified
- [ ] After source is re-processed (refresh), `ETag` changes
- [ ] Old `ETag` returns full 200 response with updated data

### 4.14 Plan Enforcement — URL Limits

- [ ] Configure a plan with `ScrapingMaxURLsPerBot: 1`
- [ ] Create one URL source (succeeds)
- [ ] Create a second URL source — verify 403 Forbidden with `ERR_URL_LIMIT_REACHED`
- [ ] Bulk create with 2 URLs — verify only 1 is created (limited by available slot)
- [ ] PDF and text sources are not affected by URL limit

### 4.15 Plan Enforcement — PDF Limits

- [ ] Configure a plan with `FilesMaxFilesPerBot: 1` and `FilesMaxSizeMB: 1`
- [ ] Upload a 500KB PDF — succeeds (201)
- [ ] Upload a second PDF — verify 403 Forbidden with `ERR_PDF_LIMIT_REACHED`
- [ ] Upload a 2MB PDF — verify 413 Request Entity Too Large with `ERR_FILE_TOO_LARGE`

### 4.16 Plan Enforcement — Text Length Limit

- [ ] Configure a plan with `FilesMaxTextLength: 100`
- [ ] Submit text with 50 characters — succeeds (201)
- [ ] Submit text with 200 characters — verify 413 Request Entity Too Large with `ERR_TEXT_TOO_LONG`

### 4.17 Plan Enforcement — Monthly Ingestion Quota

- [ ] Configure a plan with monthly ingestion limit
- [ ] Create sources until quota is exhausted
- [ ] Next source creation returns 402 Payment Required
- [ ] Bulk create also checks monthly quota and trims to available count

### 4.18 Plan Enforcement — Refresh Limits

- [ ] Configure a plan with `RefreshEnabled: false`
- [ ] `POST /api/v1/sources/{id}/refresh` — verify 403 Forbidden with `ERR_PLAN_REFRESH_UNAVAILABLE`
- [ ] Configure a plan with `RefreshEnabled: true` and `RefreshMaxMonthly: 1`
- [ ] Refresh once (succeeds)
- [ ] Refresh again — verify 402 Payment Required with `ERR_MONTHLY_REFRESH_EXCEEDED`

### 4.19 SSRF Protection

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_url=http://10.0.0.1/admin`
- [ ] Verify 403 Forbidden with `ERR_BLOCKED_URL`
- [ ] Test with `http://192.168.1.1`, `http://127.0.0.1`, `http://172.16.0.1` — all blocked
- [ ] Test with `http://169.254.169.254` (AWS metadata) — blocked
- [ ] Test with legitimate public URL — succeeds

### 4.20 Duplicate Detection

- [ ] Create a URL source with `https://example.com/page`
- [ ] Create the same URL again — verify 409 Conflict with `ERR_DUPLICATE_URL`
- [ ] Create a URL with trailing slash (`https://example.com/page/`) — verify 409 (normalized)
- [ ] Upload a PDF, then upload the same PDF file again — verify 409 Conflict with `ERR_DUPLICATE_CONTENT`
- [ ] Submit text, then submit the same text again — verify 409 Conflict with `ERR_DUPLICATE_CONTENT`
- [ ] Bulk create with a mix of new and duplicate URLs — duplicates are skipped, new ones created

### 4.21 Delete-and-Readd Cooldown

- [ ] Create a URL source, wait for completion
- [ ] Delete the source
- [ ] Immediately re-create the same URL — verify 429 Too Many Requests with `Retry-After` header and `ERR_READD_COOLDOWN_ACTIVE`
- [ ] Wait for cooldown period, re-create — succeeds

### 4.22 Cross-Tenant Access

- [ ] Create Source A under Chatbot A (User A)
- [ ] Authenticate as User B
- [ ] `GET /api/v1/sources/{sourceA_id}` — verify 404
- [ ] `DELETE /api/v1/sources/{sourceA_id}` — verify 404
- [ ] `POST /api/v1/sources/{sourceA_id}/refresh` — verify 404
- [ ] `GET /api/v1/sources/{sourceA_id}/chunks` — verify 404
- [ ] `GET /api/v1/sources/{sourceA_id}/job` — verify 404
- [ ] `POST /api/v1/sources/{sourceA_id}/job/retry` — verify 404

### 4.23 Unauthorized Access

- [ ] All source endpoints return 401 when no auth cookie is provided
- [ ] All source endpoints return 401 with an invalid/expired JWT
- [ ] All source endpoints return 401 with a malformed token

### 4.24 Method Not Allowed

- [ ] `POST /api/v1/sources/{id}` — verify 405 Method Not Allowed
- [ ] `PUT /api/v1/sources/{id}` — verify 405 Method Not Allowed
- [ ] `PATCH /api/v1/sources/{id}` — verify 405 Method Not Allowed
- [ ] `GET /api/v1/sources/{id}/refresh` — verify 405 Method Not Allowed
- [ ] `DELETE /api/v1/sources/{id}/refresh` — verify 405 Method Not Allowed
- [ ] `GET /api/v1/sources/{id}/job/retry` — verify 405 Method Not Allowed

### 4.25 Sitemap — Invalid URL

- [ ] `POST /api/v1/chatbots/{id}/sitemap/discover` with `{"sitemap_url": "not-a-url"}`
- [ ] Verify 400 Bad Request
- [ ] `POST /api/v1/chatbots/{id}/sitemap/discover` with `{"sitemap_url": "https://example.com/not-a-sitemap.xml"}`
- [ ] Verify 400 Bad Request with parse error details

### 4.26 PDF — Invalid Content Type

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=pdf` and a file that is not a PDF (e.g., `.txt` renamed to `.pdf`)
- [ ] Verify 400 Bad Request (content-type validation fails)
- [ ] Upload a PDF with `Content-Type: image/png` — verify 400 Bad Request

### 4.27 PDF — OCR Support (Pro+ Plans)

- [ ] Upload a scanned PDF (image-only, no text layer)
- [ ] With a Pro+ plan, verify OCR is applied and text is extracted
- [ ] Verify chunks contain the OCR-extracted text
- [ ] With a Free plan, verify processing fails or returns empty content

### 4.28 Source — Invalid Source Type

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=invalid`
- [ ] Verify 400 Bad Request

### 4.29 Source — Missing Required Fields

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=url` but no `source_url` — verify 400
- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=text` but no `text` field — verify 400
- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=pdf` but no `file` — verify 400

### 4.30 Processing — Storage Service Unavailable

- [ ] Shut down the R2/S3 storage service
- [ ] Upload a PDF — verify 503 Service Unavailable (`storage_missing`)
- [ ] Submit text — verify 503 Service Unavailable (`storage_missing`)

### 4.31 Concurrent Refresh and Delete

- [ ] Create a URL source, wait for completion
- [ ] Send `POST /api/v1/sources/{id}/refresh` and immediately `DELETE /api/v1/sources/{id}`
- [ ] Verify deletion succeeds (204)
- [ ] Verify no orphaned vectors remain in Qdrant
- [ ] Verify no stale processing errors in logs

### 4.32 Large PDF — Multi-Chunk Verification

- [ ] Upload a 50-page PDF document
- [ ] Wait for processing to complete
- [ ] Verify `chunk_count` reflects the expected number of chunks
- [ ] Scroll through all chunks via pagination
- [ ] Verify no content is lost between chunks
- [ ] Verify chunk overlap is present (~15% tail overlap)

### 4.33 Large URL — Multi-Chunk Verification

- [ ] Create a URL source pointing to a very long article (5000+ words)
- [ ] Wait for processing to complete
- [ ] Verify multiple chunks were created
- [ ] Verify chunks respect sentence boundaries (no mid-sentence splits)
- [ ] Verify tail overlap between consecutive chunks

### 4.34 Empty Text Source

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=text` and `text=""`
- [ ] Verify 400 Bad Request
- [ ] Submit text with only whitespace — verify 400 Bad Request

### 4.35 Empty PDF

- [ ] Upload a 0-byte PDF file
- [ ] Verify file size check passes (0 < maxSizeMB)
- [ ] Wait for processing
- [ ] Verify job fails at `parse_content` step with `ERR_EMPTY_CONTENT`

### 4.36 Source — Storage Quota Exceeded

- [ ] Configure a plan with a small storage quota
- [ ] Upload PDFs until storage quota is exhausted
- [ ] Next upload returns 402 Payment Required

### 4.37 Source — Plan Fallback Limits

- [ ] Configure a plan with `FilesMaxFilesPerBot: 0` and `FilesMaxSizeMB: 0`
- [ ] Verify safe fallback limits are applied: max 5 files, max 10MB per file
- [ ] Configure a plan with `ScrapingMaxURLsPerBot: 0`
- [ ] Verify safe fallback of 5 URLs per bot
- [ ] Configure a plan with `FilesMaxTextLength: 0`
- [ ] Verify safe fallback of 400,000 characters

### 4.38 Source — Discovered URL Depth Limit

- [ ] Create a chatbot with `discovery_mode: "auto"`
- [ ] Add a URL source (user-added, `is_discovered: false`)
- [ ] After processing, verify sub-pages were discovered and auto-created
- [ ] Verify discovered sources have `is_discovered: true`
- [ ] Process one of the discovered sources
- [ ] Verify the discovered source does NOT discover further sub-pages (1-level depth limit)

### 4.39 Source — Bulk Create with Mixed Results

- [ ] `POST /api/v1/chatbots/{id}/sources/bulk` with 5 URLs where 2 are duplicates and 1 is empty
- [ ] Verify response: `created_count: 2`, `skipped_count: 3`, `errors: []`
- [ ] Verify only 2 new sources were created

### 4.40 Source — Bulk Create Exceeds Limits

- [ ] Configure plan with `ScrapingMaxURLsPerBot: 3`
- [ ] Create 2 URL sources manually
- [ ] Bulk create with 5 URLs — verify only 1 is created (available slot)
- [ ] Verify response `created_count: 1`

### 4.41 Source — Bulk Create Empty Array

- [ ] `POST /api/v1/chatbots/{id}/sources/bulk` with `{"urls": []}`
- [ ] Verify 400 Bad Request with `ERR_NO_URLS_PROVIDED`

### 4.42 Source — Bulk Create Invalid JSON

- [ ] `POST /api/v1/chatbots/{id}/sources/bulk` with malformed body
- [ ] Verify 400 Bad Request with `ERR_INVALID_REQUEST_BODY`

### 4.43 Source — Sitemap Discovery with Path Filters

- [ ] Configure chatbot with `include_paths: ["/products/"]` and `exclude_paths: ["/products/out-of-stock/"]`
- [ ] `POST /api/v1/chatbots/{id}/sitemap/discover` with a sitemap containing `/products/`, `/products/out-of-stock/`, `/about/` URLs
- [ ] Verify only `/products/` URLs (excluding `/products/out-of-stock/`) are returned
- [ ] Verify `total_count` reflects the filtered count

### 4.44 Source — Sitemap Discovery Deduplication

- [ ] `POST /api/v1/chatbots/{id}/sitemap/discover` with a sitemap containing duplicate URLs
- [ ] Verify `urls` array has no duplicates
- [ ] Verify `total_count` reflects the deduplicated count

### 4.45 Source — Sitemap Discovery Nested Sitemaps

- [ ] `POST /api/v1/chatbots/{id}/sitemap/discover` with a sitemap index that references child sitemaps
- [ ] Verify all URLs from all child sitemaps are discovered and flattened
- [ ] Verify `total_count` reflects the aggregate count

### 4.46 Source — Delete with Vectors

- [ ] Create a source, wait for completion
- [ ] Verify vectors exist in Qdrant (chunks endpoint returns data)
- [ ] `DELETE /api/v1/sources/{id}`
- [ ] `GET /api/v1/sources/{id}/chunks` — verify 404 (source deleted)
- [ ] Verify vectors are removed from Qdrant (direct Qdrant query returns no points for source ID)

### 4.47 Source — Delete Non-Existent

- [ ] `DELETE /api/v1/sources/non-existent-id`
- [ ] Verify 404 Not Found

### 4.48 Source — Get Non-Existent

- [ ] `GET /api/v1/sources/non-existent-id`
- [ ] Verify 404 Not Found

### 4.49 Source — Get Chunks for Non-Existent Source

- [ ] `GET /api/v1/sources/non-existent-id/chunks`
- [ ] Verify 404 Not Found

### 4.50 Source — Refresh Non-Existent

- [ ] `POST /api/v1/sources/non-existent-id/refresh`
- [ ] Verify 404 Not Found

### 4.51 Source — Refresh Non-URL Source

- [ ] Create a PDF source, wait for completion
- [ ] `POST /api/v1/sources/{id}/refresh`
- [ ] Verify 400 Bad Request with `ERR_ONLY_URL_REFRESH`
- [ ] Repeat with text source — same result

### 4.52 Source — Refresh While Processing

- [ ] Create a URL source
- [ ] Immediately `POST /api/v1/sources/{id}/refresh` (while still `pending`)
- [ ] Verify 409 Conflict with `ERR_SOURCE_ALREADY_PROCESSING`
- [ ] Wait for processing to start (`status: "processing"`)
- [ ] `POST /api/v1/sources/{id}/refresh` — verify 409 Conflict

### 4.53 Source — ETag Caching After Refresh

- [ ] `GET /api/v1/sources/{id}` — note ETag
- [ ] Refresh the source
- [ ] `GET /api/v1/sources/{id}` — verify ETag has changed (new `processed_at` or `chunk_count`)
- [ ] `GET /api/v1/sources/{id}` with old `If-None-Match` — verify 200 (stale ETag)
- [ ] `GET /api/v1/sources/{id}` with new ETag — verify 304

### 4.54 Source — ETag for Unprocessed Source

- [ ] Create a source (status: `pending`, no `processed_at`, `chunk_count: 0`)
- [ ] `GET /api/v1/sources/{id}` — verify ETag is `pending--0`
- [ ] `GET /api/v1/sources/{id}` with `If-None-Match: pending--0` — verify 304

### 4.55 Source — Delete During Processing

- [ ] Create a URL source
- [ ] Immediately delete it while it's still `pending`
- [ ] Verify 204 No Content
- [ ] Verify processing job completes but finds source deleted (logged, no crash)
- [ ] Verify no orphaned vectors in Qdrant

### 4.56 Source — Delete and Re-create

- [ ] Create a URL source, wait for completion
- [ ] Delete the source
- [ ] Wait for cooldown period
- [ ] Re-create the same URL
- [ ] Verify it's treated as a new source (new ID, fresh processing)

### 4.57 Source — Invalid UUID Format

- [ ] `GET /api/v1/sources/not-a-uuid` — verify 400 Bad Request
- [ ] `DELETE /api/v1/sources/not-a-uuid` — verify 400 Bad Request
- [ ] `POST /api/v1/sources/not-a-uuid/refresh` — verify 400 Bad Request
- [ ] `GET /api/v1/sources/not-a-uuid/chunks` — verify 400 Bad Request
- [ ] `GET /api/v1/sources/not-a-uuid/job` — verify 400 Bad Request
- [ ] `POST /api/v1/sources/not-a-uuid/job/retry` — verify 400 Bad Request

### 4.58 Source — Invalid Chatbot ID

- [ ] `GET /api/v1/chatbots/non-existent/sources` — verify 404
- [ ] `POST /api/v1/chatbots/non-existent/sources` — verify 404
- [ ] `POST /api/v1/chatbots/non-existent/sources/bulk` — verify 404
- [ ] `GET /api/v1/chatbots/non-existent/pending-urls` — verify 404
- [ ] `POST /api/v1/chatbots/non-existent/sitemap/discover` — verify 404

### 4.59 Source — File Path Storage Key Format

- [ ] Upload a PDF to a chatbot with org and workspace IDs
- [ ] Verify the storage key follows the pattern: `{orgID}/{wsID}/{botID}/{filename}`
- [ ] Verify the file is retrievable from R2/S3 using the stored `file_path`

### 4.60 Source — Legacy Bot Fallback Key

- [ ] Upload a PDF to a chatbot without org/workspace IDs (legacy)
- [ ] Verify the storage key follows the fallback pattern: `sources/{filename}`

### 4.61 Source — Delete Re-aggregates Suggestions

- [ ] Create 2 sources with different content, wait for completion
- [ ] Note chatbot's `suggested_questions`
- [ ] Delete one source
- [ ] Verify `ReAggregateSuggestionsForChatbot` is called (background goroutine)
- [ ] Verify chatbot's `suggested_questions` is updated (may have fewer questions)

### 4.62 Source — Delete with Storage Error

- [ ] Upload a PDF source
- [ ] Make R2/S3 storage unavailable
- [ ] Delete the source
- [ ] Verify 204 No Content (storage error is logged but does not block deletion)
- [ ] Verify source is soft-deleted in database
- [ ] Verify vectors are deleted from Qdrant

### 4.63 Source — Delete with Qdrant Error

- [ ] Make Qdrant unavailable
- [ ] Delete a source
- [ ] Verify 204 No Content (Qdrant error is logged as warning, does not block deletion)
- [ ] Verify source is soft-deleted in database

### 4.64 Source — Bulk Create with SSRF Validation

- [ ] `POST /api/v1/chatbots/{id}/sources/bulk` with `{"urls": ["http://10.0.0.1/admin"]}`
- [ ] Verify the URL is blocked by SSRF validation (logged as `ssrf_blocked`)
- [ ] Verify the URL is skipped (counted in `skipped_count`, not `created_count`)

### 4.65 Source — URL Normalization

- [ ] Create a URL source with `https://Example.COM/Path/`
- [ ] Verify the stored URL is normalized (lowercase host, trailing slash handling)
- [ ] Create the same URL with different casing — verify 409 Conflict (normalized match)

### 4.66 Source — Discovered URL Does Not Crawl Further

- [ ] Create a chatbot with `discovery_mode: "auto"`
- [ ] Add a URL source (user-added)
- [ ] After processing, verify sub-pages were discovered with `is_discovered: true`
- [ ] Process a discovered source
- [ ] Verify the discovered source does NOT discover further sub-pages (logged as `url_discovery_skipped_is_discovered`)

### 4.67 Source — Discovery with Plan Limit

- [ ] Configure plan with `ScrapingMaxURLsPerBot: 3` and `ScrapingMaxPagesPerCrawl: 5`
- [ ] Add a URL source (1/3 slots used)
- [ ] After processing, verify at most 2 discovered sub-pages were added (2 remaining slots)
- [ ] Verify the discovery log shows `limit_reached` if more than 2 links were found

### 4.68 Source — Discovery with Single URL Plan

- [ ] Configure plan with `ScrapingMaxURLsPerBot: 1`
- [ ] Add a URL source
- [ ] After processing, verify no sub-pages were discovered (logged as `url_discovery_skipped_single_url_plan`)

### 4.69 Source — Discovery with Zero Crawl Limit

- [ ] Configure plan with `ScrapingMaxPagesPerCrawl: 0`
- [ ] Add a URL source
- [ ] After processing, verify no sub-pages were discovered (logged as `url_discovery_skipped_no_crawl_limit`)

### 4.70 Source — Discovery Path Filter Error

- [ ] Configure chatbot with invalid regex in `include_paths`
- [ ] Add a URL source
- [ ] Verify discovery continues without filter (logged as `url_discovery_path_filter_failed`)
- [ ] Verify sub-pages are discovered without path filtering

### 4.71 Source — Discovery HTML Fetch Failure

- [ ] Add a URL source that fails HTML fetch but succeeds at text extraction
- [ ] Verify processing completes (text is extracted)
- [ ] Verify no sub-pages were discovered (logged as `url_processing_fetch_html_failed`)
- [ ] Verify source status is `completed`

### 4.72 Source — Discovery No Links Found

- [ ] Add a URL source pointing to a page with no internal links
- [ ] After processing, verify no sub-pages were discovered (logged as `url_discovery_no_links_found`)

### 4.73 Source — Discovery Extract Links Failure

- [ ] Add a URL source where link extraction fails
- [ ] Verify processing continues (discovery failure is non-fatal)
- [ ] Verify source status is `completed`

### 4.74 Source — Create with Maximum Allowed Fields

- [ ] Create a URL source with all optional fields populated
- [ ] Verify all fields are stored correctly

### 4.75 Source — Create with Special Characters in URL

- [ ] Create a URL source with URL-encoded characters (`https://example.com/path%20with%20spaces`)
- [ ] Verify URL is stored and processed correctly
- [ ] Create a URL source with Unicode in path (`https://example.com/über-cool`)
- [ ] Verify URL is normalized and processed

### 4.76 Source — Delete with File Cleanup

- [ ] Upload a PDF, wait for completion
- [ ] Verify the file exists in R2/S3 storage
- [ ] Delete the source
- [ ] Verify the file is deleted from R2/S3 storage (logged, best-effort)
- [ ] Verify source is soft-deleted even if file deletion fails

### 4.77 Source — Delete Without Storage

- [ ] Configure the server without a storage service (`Storage: nil`)
- [ ] Create a URL source (no file storage needed)
- [ ] Delete the source
- [ ] Verify 204 No Content (no storage deletion attempted)

### 4.78 Source — Bulk Create with Mixed Source Types

- [ ] `POST /api/v1/chatbots/{id}/sources/bulk` — verify only URL sources are supported
- [ ] Verify the handler only creates `source_type: "url"` sources

### 4.79 Source — Create with Empty Form

- [ ] `POST /api/v1/chatbots/{id}/sources` with no form fields
- [ ] Verify 400 Bad Request (defaults to `source_type=pdf` but no file provided)

### 4.80 Source — Create with Invalid Content-Type

- [ ] `POST /api/v1/chatbots/{id}/sources` with `Content-Type: application/json` instead of `multipart/form-data`
- [ ] Verify 400 Bad Request (multipart form parse fails)

### 4.81 Source — List with Multiple Sources

- [ ] Create 5 sources of mixed types (2 URL, 2 PDF, 1 text)
- [ ] `GET /api/v1/chatbots/{id}/sources` — verify all 5 are returned
- [ ] Verify sources are ordered by creation date (newest first or as implemented)

### 4.82 Source — List with Deleted Sources

- [ ] Create 3 sources
- [ ] Delete 1 source
- [ ] `GET /api/v1/chatbots/{id}/sources` — verify only 2 sources are returned
- [ ] Verify the deleted source is not in the list

### 4.83 Source — List Empty Chatbot

- [ ] Create a new chatbot with no sources
- [ ] `GET /api/v1/chatbots/{id}/sources` — verify `[]` (empty array)

### 4.84 Source — Create with Maximum Text Length

- [ ] Submit text at exactly the plan's `FilesMaxTextLength` limit
- [ ] Verify 201 Created
- [ ] Submit text at `FilesMaxTextLength + 1` — verify 413

### 4.85 Source — Create with Maximum File Size

- [ ] Upload a PDF at exactly the plan's `FilesMaxSizeMB` limit
- [ ] Verify 201 Created
- [ ] Upload a PDF at `FilesMaxSizeMB + 1 byte` — verify 413

### 4.86 Source — Create with Maximum File Count

- [ ] Upload PDFs until `FilesMaxFilesPerBot` is reached
- [ ] Next upload returns 403 Forbidden with `ERR_PDF_LIMIT_REACHED`

### 4.87 Source — Refresh with Monthly Quota

- [ ] Configure plan with `RefreshMaxMonthly: 2`
- [ ] Refresh a source twice (both succeed)
- [ ] Third refresh returns 402 Payment Required with `ERR_MONTHLY_REFRESH_EXCEEDED`

### 4.88 Source — Refresh Cooldown Timing

- [ ] Configure plan with cooldown period
- [ ] Refresh a source
- [ ] Immediately check `Retry-After` header on next refresh attempt (429)
- [ ] Verify `Retry-After` value matches the configured cooldown

### 4.89 Source — Discovered URL Does Not Count Toward Manual Limit

- [ ] Configure plan with `ScrapingMaxURLsPerBot: 2`
- [ ] Add a URL source manually (1/2)
- [ ] After discovery, verify discovered URLs are created (may exceed 2 total)
- [ ] Verify the limit check uses `maxTotal = maxPagesPerCrawl + maxURLsPerBot`
- [ ] Verify manual creation is still limited to 2
