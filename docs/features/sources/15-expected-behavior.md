> Extracted from: sources.md — Section: 6. Expected Behavior

# 6. Expected Behavior

### Success Conditions

| Condition | Indicator |
|-----------|-----------|
| Source created successfully | 201 Created with `{"id": "source-uuid"}` |
| Source listed | 200 OK with array of DataSource objects |
| Source retrieved | 200 OK with full DataSource object |
| Source deleted | 204 No Content |
| Source refreshed | 202 Accepted with `{"id": "source-uuid"}` |
| Chunks retrieved | 200 OK with `{"chunks": [...], "next_cursor": "..."}` |
| Sitemap discovered | 200 OK with `{"urls": [...], "total_count": N}` |
| ETag match | 304 Not Modified (no body) |
| Processing completed | Source `status: "completed"`, `chunk_count > 0`, `processed_at` set |
| Content unchanged on refresh | Job completes with `Skipped: true`, same `chunk_count` |
| Bulk create (partial success) | 200 OK with `created_count`, `skipped_count`, `errors` |
| Bulk create (all created) | 201 Created with `created_count`, `skipped_count`, `errors` |

### Failure Conditions

| Condition | Indicator |
|-----------|-----------|
| Unauthenticated request | 401 Unauthorized |
| Cross-tenant access | 404 Not Found |
| Invalid UUID | 400 Bad Request |
| Method not allowed | 405 Method Not Allowed |
| Duplicate URL | 409 Conflict with `ERR_DUPLICATE_URL` |
| Duplicate content (PDF/text) | 409 Conflict with `ERR_DUPLICATE_CONTENT` |
| URL limit reached | 403 Forbidden with `ERR_URL_LIMIT_REACHED` |
| PDF limit reached | 403 Forbidden with `ERR_PDF_LIMIT_REACHED` |
| File too large | 413 Request Entity Too Large with `ERR_FILE_TOO_LARGE` |
| Text too long | 413 Request Entity Too Large with `ERR_TEXT_TOO_LONG` |
| SSRF blocked | 403 Forbidden with `ERR_BLOCKED_URL` |
| Re-add cooldown | 429 Too Many Requests with `Retry-After` |
| Monthly ingestion exceeded | 402 Payment Required |
| Monthly refresh exceeded | 402 Payment Required with `ERR_MONTHLY_REFRESH_EXCEEDED` |
| Refresh not available on plan | 403 Forbidden with `ERR_PLAN_REFRESH_UNAVAILABLE` |
| Source already processing | 409 Conflict with `ERR_SOURCE_ALREADY_PROCESSING` |
| Only URL refresh allowed | 400 Bad Request with `ERR_ONLY_URL_REFRESH` |
| Refresh cooldown | 429 Too Many Requests with `Retry-After` |
| Storage unavailable | 503 Service Unavailable |
| Invalid request body | 400 Bad Request with `ERR_INVALID_REQUEST_BODY` |
| No URLs provided (bulk) | 400 Bad Request with `ERR_NO_URLS_PROVIDED` |
| Invalid sitemap URL | 400 Bad Request |
| Sitemap parse failed | 400 Bad Request with error details |
