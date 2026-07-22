> Extracted from: background-processing.md — Section: 4. Test Scenarios

# 4. Test Scenarios

### 4.1 Processing Pipeline — Step-by-Step Progress

- [ ] Create a URL source
- [ ] Immediately `GET /api/v1/sources/{id}/job` — verify `status: "pending"`, `progress_percent: 0`
- [ ] Shortly after, verify `status: "running"`, `current_step: "fetch_source"`, `progress_percent: 10`
- [ ] Then `current_step: "parse_content"`, `progress_percent: 30`
- [ ] Then `current_step: "chunk_text"`, `progress_percent: 50`
- [ ] Then `current_step: "embed_chunks"`, `progress_percent: 80`
- [ ] Then `current_step: "store_vectors"`, `progress_percent: 100`
- [ ] Finally `status: "completed"`, `progress_percent: 100`

### 4.2 Processing Failure — Invalid URL

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_url=https://this-domain-does-not-exist-12345.com`
- [ ] Wait for processing to fail
- [ ] `GET /api/v1/sources/{id}/job` shows `status: "failed"`, `error_code: "ERR_SCRAPE_NETWORK"`, `failed_step: "parse_content"`
- [ ] Source status is `failed` with `error_message` set

### 4.3 Processing Failure — Dynamic JS Required

- [ ] Create a URL source pointing to a JS-heavy SPA site
- [ ] Use a plan with `ScrapingDynamicEnabled: false`
- [ ] Wait for processing to fail
- [ ] Verify `error_code: "ERR_DYNAMIC_REQUIRED"` and `failed_step: "parse_content"`

### 4.4 Processing Failure — PDF Parse Error

- [ ] Upload a corrupted/non-PDF file with `.pdf` extension
- [ ] Wait for processing to fail
- [ ] Verify `error_code: "ERR_PDF_PARSE_FAILED"` and `failed_step: "parse_content"`

### 4.5 Processing Failure — Empty Content

- [ ] Create a URL source pointing to a page with no extractable text (e.g., image-only page)
- [ ] Wait for processing to fail
- [ ] Verify `error_code: "ERR_EMPTY_CONTENT"` and `failed_step: "parse_content"`

### 4.6 Processing Failure — Queue Full

- [ ] Configure worker pool with 1 worker and queue buffer of 1
- [ ] Enqueue multiple sources rapidly
- [ ] Verify one source fails with `error_code: "QUEUE_FULL"` and `failed_step: "fetch_source"`

### 4.7 Processing — Auto-Retry

- [ ] Point a URL source to a server that returns 503 temporarily
- [ ] Verify the job is retried (check `retry_count` increments)
- [ ] After the server recovers, verify the job completes
- [ ] Verify backoff timing: 2s, 4s, 8s between retries

### 4.8 Processing — Max Retries Exceeded

- [ ] Point a URL source to a permanently unreachable server
- [ ] Wait for all 3 retries to exhaust
- [ ] Verify `status: "failed"`, `error_code: "ERR_SCRAPE_NETWORK"`, `retry_count: 3`
- [ ] Verify source status is `failed` with `error_message`

### 4.9 Processing — Concurrent Sources

- [ ] Create 10 URL sources simultaneously
- [ ] Verify all are enqueued and processed (may take time depending on worker count)
- [ ] Verify all reach `completed` or `failed` status
- [ ] Verify no sources are stuck in `pending` or `processing` indefinitely

### 4.10 Processing — Startup Recovery

- [ ] Create a source and enqueue it
- [ ] Simulate a crash (kill the server) while a job is `running`
- [ ] Restart the server
- [ ] Verify the stuck job is recovered and re-enqueued (logged as `recovering_stale_running_job`)
- [ ] Verify the source eventually reaches `completed` or `failed`

### 4.11 Chunk Inspection — Content Integrity

- [ ] Create a text source with known content
- [ ] After processing, `GET /api/v1/sources/{id}/chunks`
- [ ] Verify chunks contain the original text split at sentence/paragraph boundaries
- [ ] Verify no content is lost (reconstruct original text from chunks)
- [ ] Verify chunk overlap (~15% tail overlap between consecutive chunks)
- [ ] Verify each chunk has a deterministic point ID via `MakePointID(sourceID, index)`

### 4.12 Chunk Inspection — Token Counts

- [ ] Create a text source with content of known token count
- [ ] After processing, inspect chunks
- [ ] Verify each chunk's `token_count` is ≤ 512 (target token size)
- [ ] Verify chunks respect sentence boundaries (no mid-sentence splits)
- [ ] Verify chunks respect paragraph boundaries (paragraphs are not split unless exceeding token limit)

### 4.13 Suggestion Re-aggregation After Deletion

- [ ] Create 3 URL sources, wait for all to complete
- [ ] Note the chatbot's `suggested_questions` (6 items)
- [ ] Delete one source
- [ ] Verify suggestions are re-aggregated (background goroutine)
- [ ] Verify the chatbot's `suggested_questions` may have changed (fewer or different questions)

### 4.14 Suggestion Re-aggregation After New Source

- [ ] Create a chatbot with no sources — verify `suggested_questions` is empty
- [ ] Add a URL source, wait for completion
- [ ] Verify `suggested_questions` is populated with questions from the source
- [ ] Add a second source with different content
- [ ] Verify `suggested_questions` is updated with questions from both sources
- [ ] Verify weighted selection: source with more chunks contributes more questions

### 4.15 Concurrent Processing — Multiple Source Types

- [ ] Create a URL source, a PDF source, and a text source simultaneously
- [ ] Verify all three are processed concurrently (depending on worker availability)
- [ ] Verify each reaches `completed` status
- [ ] Verify the chatbot can answer questions from all three sources

### 4.16 Processing — Plan Limit Enforcement During Discovery

- [ ] Configure a plan with `ScrapingMaxURLsPerBot: 2` and `ScrapingMaxPagesPerCrawl: 3`
- [ ] Add a URL source (1/2 slots used)
- [ ] After processing, verify at most 1 discovered sub-page was added (2nd slot)
- [ ] Add another URL source manually (2/2 slots used)
- [ ] Add a third URL source — verify 403 Forbidden with `ERR_URL_LIMIT_REACHED`

### 4.17 Processing — Discovery with Path Filters

- [ ] Configure chatbot with `include_paths: ["/docs/"]` and `exclude_paths: ["/docs/archive/"]`
- [ ] Add a URL source that links to `/docs/`, `/docs/archive/`, `/blog/` pages
- [ ] After processing, verify only `/docs/` URLs (excluding `/docs/archive/`) were discovered
- [ ] Verify `/blog/` URLs were excluded by the include filter

### 4.18 Processing — Discovery Disabled Mode

- [ ] Configure chatbot with `discovery_mode: "disabled"`
- [ ] Add a URL source with many internal links
- [ ] After processing, verify no sub-pages were discovered
- [ ] Verify only the original source exists

### 4.19 Processing — CSS Selector Whitelist

- [ ] Configure chatbot with `selector_whitelist: [".main-content", "article"]`
- [ ] Add a URL source pointing to a page with diverse content
- [ ] After processing, verify only content from `.main-content` and `article` elements was extracted
- [ ] Verify chunks do not contain content from excluded elements (nav, footer, sidebar)

### 4.20 Processing — Qdrant Unavailable

- [ ] Shut down Qdrant
- [ ] Create a text source and wait for processing
- [ ] Verify job fails at `embed_chunks` step with `ERR_EMBEDDING_FAILED`

### 4.21 Processing — OpenAI API Error

- [ ] Configure an invalid OpenAI API key
- [ ] Create a text source
- [ ] Verify job fails at `embed_chunks` step with `ERR_EMBEDDING_FAILED`

### 4.22 Processing — Suggestion Generation Failure

- [ ] Configure an invalid OpenAI API key for LLM calls (but valid for embeddings)
- [ ] Create a text source
- [ ] Verify processing completes (embedding succeeds)
- [ ] Verify `extract_metadata_failed` is logged but processing is not blocked
- [ ] Verify source status is `completed` despite metadata extraction failure

### 4.23 Processing — Suggestion Re-aggregation Panic Recovery

- [ ] Force a panic during suggestion re-aggregation (e.g., nil chatbot repo)
- [ ] Verify the panic is recovered (logged as `re_aggregate_panic`)
- [ ] Verify the server does not crash
- [ ] Verify the source deletion/processing still completes

### 4.24 Concurrent Source Operations

- [ ] While a source is `processing`, attempt to delete it
- [ ] Verify deletion succeeds (204) — vectors are cleaned up
- [ ] Verify the processing job completes but finds source deleted (logged, no crash)
- [ ] While a source is `processing`, attempt to refresh it
- [ ] Verify 409 Conflict with `ERR_SOURCE_ALREADY_PROCESSING`

### 4.25 Worker Pool — Max Workers

- [ ] Configure worker pool with 16 workers
- [ ] Enqueue 20 sources simultaneously
- [ ] Verify 16 are processed concurrently, 4 wait in queue
- [ ] Verify all eventually complete

### 4.26 Worker Pool — Graceful Shutdown

- [ ] Enqueue several long-processing sources
- [ ] Call `Stop()` on the queue
- [ ] Verify workers finish their current job before shutting down
- [ ] Verify no jobs are lost (remaining jobs stay in queue for next startup)

### 4.27 Processing — URL with CSS Selectors

- [ ] Configure chatbot with `selector_whitelist: ["h1", ".summary"]`
- [ ] Add a URL source pointing to a page with diverse content
- [ ] After processing, verify only `h1` and `.summary` content was extracted
- [ ] Verify chunks do not contain content from other elements

### 4.28 Processing — URL with Dynamic Rendering

- [ ] Configure a plan with `ScrapingDynamicEnabled: true`
- [ ] Add a URL source pointing to a JS-rendered SPA
- [ ] After processing, verify content was extracted (go-rod dynamic rendering)
- [ ] Verify chunks contain the JS-rendered content

### 4.29 Processing — URL Dynamic Rendering Not Available

- [ ] Configure a plan with `ScrapingDynamicEnabled: false`
- [ ] Add a URL source pointing to a JS-rendered SPA
- [ ] Wait for processing to fail
- [ ] Verify `error_code: "ERR_DYNAMIC_REQUIRED"` and `failed_step: "parse_content"`

### 4.30 Processing — URL with Redirect

- [ ] Create a URL source pointing to a URL that 301 redirects
- [ ] Verify processing follows the redirect and extracts content from the final destination
- [ ] Verify the source URL in the database remains the original (not the redirect target)

### 4.31 Processing — URL with Authentication

- [ ] Create a URL source pointing to a page behind a login wall
- [ ] Verify processing fails with appropriate error (empty content or scrape error)
- [ ] Verify no credentials are leaked in logs

### 4.32 Processing — Very Large Content

- [ ] Create a text source with 1MB of content
- [ ] Verify processing completes
- [ ] Verify multiple chunks were created (512 tokens each)
- [ ] Verify no OOM or timeout errors

### 4.33 Processing — Special Characters

- [ ] Create a text source with Unicode, emoji, HTML entities, and special characters
- [ ] Verify processing completes
- [ ] Verify chunks preserve the special characters correctly
- [ ] Verify embedding does not fail on special characters

### 4.34 Processing — Turkish Content

- [ ] Create a text source with Turkish content (including İ, ı, ş, ç, ö, ü, ğ characters)
- [ ] Verify `text.NormalizeTR` is applied
- [ ] Verify chunks preserve Turkish characters correctly
- [ ] Verify Turkish sentence tokenization works (abbreviations, sentence boundaries)

### 4.35 Processing — Multiple Languages

- [ ] Create a text source with mixed English and Turkish content
- [ ] Verify processing completes
- [ ] Verify chunks preserve both languages correctly

### 4.36 Processing — URL with Query Parameters

- [ ] Create a URL source with query parameters (`https://example.com/page?utm_source=test&ref=123`)
- [ ] Verify URL is normalized and stored
- [ ] Verify processing extracts content correctly
- [ ] Verify duplicate detection works with/without query params (normalized comparison)

### 4.37 Processing — URL Fragment

- [ ] Create a URL source with a fragment (`https://example.com/page#section`)
- [ ] Verify fragment is stripped during normalization
- [ ] Verify processing extracts content from the page

### 4.38 Processing — PDF with Images Only

- [ ] Upload a PDF that contains only images (no text layer)
- [ ] With a Free plan (no OCR), verify processing fails with `ERR_EMPTY_CONTENT`
- [ ] With a Pro+ plan (OCR enabled), verify text is extracted from images

### 4.39 Processing — PDF Password Protected

- [ ] Upload a password-protected PDF
- [ ] Verify processing fails at `parse_content` step with `ERR_PDF_PARSE_FAILED`

### 4.40 Processing — URL Timeout

- [ ] Create a URL source pointing to a very slow server (30s+ response time)
- [ ] Verify processing fails with `ERR_SCRAPE_TIMEOUT` at `parse_content` step

### 4.41 Processing — URL Forbidden

- [ ] Create a URL source pointing to a server that returns 403
- [ ] Verify processing fails with `ERR_SCRAPE_FORBIDDEN` at `parse_content` step

### 4.42 Processing — URL Network Error

- [ ] Create a URL source pointing to a non-routable IP
- [ ] Verify processing fails with `ERR_SCRAPE_NETWORK` at `parse_content` step

### 4.43 Processing — URL Empty Content

- [ ] Create a URL source pointing to a page with no visible text (e.g., blank page, image gallery)
- [ ] Verify processing fails with `ERR_EMPTY_CONTENT` at `parse_content` step

### 4.44 Processing — Multiple Workers

- [ ] Configure 4 workers
- [ ] Enqueue 4 sources simultaneously
- [ ] Verify all 4 are processed in parallel (check `started_at` timestamps overlap)
- [ ] Verify total processing time is ~1/4 of sequential time

### 4.45 Processing — Worker Count Capped at 16

- [ ] Configure worker pool with 20 workers
- [ ] Verify only 16 workers are created (logged as `worker_count_capped`)
- [ ] Verify `WorkerCount()` returns 16

### 4.46 Processing — Source Type Dispatch

- [ ] Create a URL source — verify `URLProcessor.ProcessWithSteps` is called
- [ ] Create a PDF source — verify `PDFProcessor.ProcessWithSteps` is called
- [ ] Create a text source — verify `TextProcessor.ProcessWithSteps` is called
- [ ] Verify unknown source type returns `ERR_UNKNOWN_SOURCE_TYPE`

### 4.47 Processing — Step Resume After Crash

- [ ] Create a URL source
- [ ] Let processing reach `chunk_text` step
- [ ] Simulate a crash
- [ ] Restart server — verify job is recovered
- [ ] Verify processing resumes from `chunk_text` (not from `fetch_source`)
- [ ] Verify no duplicate vectors are created

### 4.48 Processing — Step Progress Accuracy

- [ ] Monitor `GET /api/v1/sources/{id}/job` during processing
- [ ] Verify `progress_percent` matches the current step:
  - `fetch_source` → 10%
  - `parse_content` → 30%
  - `chunk_text` → 50%
  - `embed_chunks` → 80%
  - `store_vectors` → 100%

### 4.49 Processing — Embedding Batch Size

- [ ] Create a source with enough content to produce 30+ chunks
- [ ] Verify embeddings are created in batches of 25 chunks
- [ ] Verify all chunks are successfully embedded and stored

### 4.50 Processing — Rate Limiting Between Jobs

- [ ] Enqueue 3 sources
- [ ] Verify there is a ~500ms delay between job processing starts
- [ ] Verify this delay is NOT applied in test environment (`GO_ENV=test`)

### 4.51 Processing — URL with Non-HTML Content

- [ ] Create a URL source pointing to a JSON API endpoint
- [ ] Verify processing extracts the raw JSON as text content
- [ ] Verify chunks contain the JSON content

### 4.52 Processing — URL with Binary Content

- [ ] Create a URL source pointing to a binary file (e.g., `.zip`, `.exe`)
- [ ] Verify processing fails with `ERR_EMPTY_CONTENT` or `ERR_SCRAPE_NETWORK`

### 4.53 Processing — URL with Very Long Page

- [ ] Create a URL source pointing to a very long page (100K+ characters)
- [ ] Verify processing completes
- [ ] Verify multiple chunks are created
- [ ] Verify no single chunk exceeds 512 tokens

### 4.54 Processing — Content Hash on Refresh

- [ ] Create a URL source, wait for completion
- [ ] Note the source's `hash` field
- [ ] Refresh the source (content unchanged on remote)
- [ ] Verify the job completes with `Skipped: true` (content unchanged)
- [ ] Verify `chunk_count` remains the same
- [ ] Verify old vectors were NOT deleted (no re-embedding needed)

### 4.55 Processing — Content Changed on Refresh

- [ ] Create a URL source pointing to a page you can modify
- [ ] Wait for completion, note `hash` and `chunk_count`
- [ ] Modify the remote page content
- [ ] Refresh the source
- [ ] Verify old vectors were deleted before re-embedding
- [ ] Verify new `hash` differs from old `hash`
- [ ] Verify `chunk_count` may differ

### 4.56 Processing — Embedding Token Usage Tracking

- [ ] Create a text source
- [ ] After processing, verify `usageRepo.AddEmbeddingTokens` was called
- [ ] Verify the token count is accurate (sum of all chunk token counts)
- [ ] Verify `usageRepo.IncrementSuccessfulIngestion` was called

### 4.57 Processing — PDF Token Usage Tracking

- [ ] Upload a PDF
- [ ] After processing, verify embedding tokens and ingestion count are tracked
- [ ] Verify the same tracking occurs for URL and text sources

### 4.58 Processing — URL with Fragment Identifier

- [ ] Create a URL source with `https://example.com/page#section`
- [ ] Verify the fragment is stripped during normalization
- [ ] Verify the stored URL does not include the fragment
- [ ] Verify processing extracts content from the page

### 4.59 Processing — URL with Authentication Parameters

- [ ] Create a URL source with `https://example.com/page?token=secret&api_key=123`
- [ ] Verify the URL is stored with query parameters intact
- [ ] Verify processing extracts content (query params are passed to the scraper)

### 4.60 Processing — URL with International Characters

- [ ] Create a URL source with internationalized domain name (IDN) like `https://münchen.de`
- [ ] Verify URL is normalized (Punycode encoding)
- [ ] Verify processing extracts content

### 4.61 Processing — URL with Port Number

- [ ] Create a URL source with `https://example.com:8080/page`
- [ ] Verify processing connects to the correct port
- [ ] Verify the URL is stored with the port number

### 4.62 Processing — URL with robots.txt Restrictions

- [ ] Create a URL source pointing to a page disallowed by `robots.txt`
- [ ] Verify the scraper respects `robots.txt` (if implemented) or note the behavior

### 4.63 Processing — URL with Rate Limiting

- [ ] Create multiple URL sources pointing to the same domain
- [ ] Verify the scraper respects rate limits (if implemented)
- [ ] Verify no sources fail due to rate limiting from the target domain

### 4.64 Processing — URL with HTTPS

- [ ] Create a URL source with `https://example.com`
- [ ] Verify processing uses HTTPS
- [ ] Verify content is extracted correctly

### 4.65 Processing — URL with HTTP (Non-HTTPS)

- [ ] Create a URL source with `http://example.com`
- [ ] Verify processing works with HTTP
- [ ] Note: SSRF validation still applies

### 4.66 Processing — URL with Non-Standard Port

- [ ] Create a URL source with `https://example.com:3000/api`
- [ ] Verify processing connects to port 3000
- [ ] Verify content is extracted
