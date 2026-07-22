> Extracted from: sources.md — Section: 5. Edge Cases

# 5. Edge Cases

| # | Edge Case | Expected Behavior |
|---|-----------|-------------------|
| 1 | SSRF attempt with private IP | 403 Forbidden, logged as `ssrf_blocked` |
| 2 | Duplicate URL (exact match) | 409 Conflict with `ERR_DUPLICATE_URL` |
| 3 | Duplicate URL (trailing slash variant) | 409 Conflict (normalized match) |
| 4 | Duplicate PDF content (same file) | 409 Conflict with `ERR_DUPLICATE_CONTENT` |
| 5 | Duplicate text content | 409 Conflict with `ERR_DUPLICATE_CONTENT` |
| 6 | Delete-and-readd within cooldown | 429 Too Many Requests with `Retry-After` |
| 7 | URL limit reached | 403 Forbidden with `ERR_URL_LIMIT_REACHED` |
| 8 | PDF file count limit reached | 403 Forbidden with `ERR_PDF_LIMIT_REACHED` |
| 9 | PDF file size exceeds limit | 413 Request Entity Too Large with `ERR_FILE_TOO_LARGE` |
| 10 | Text length exceeds limit | 413 Request Entity Too Large with `ERR_TEXT_TOO_LONG` |
| 11 | Monthly ingestion quota exceeded | 402 Payment Required |
| 12 | Monthly refresh quota exceeded | 402 Payment Required with `ERR_MONTHLY_REFRESH_EXCEEDED` |
| 13 | Refresh not enabled on plan | 403 Forbidden with `ERR_PLAN_REFRESH_UNAVAILABLE` |
| 14 | Refresh while source is processing | 409 Conflict with `ERR_SOURCE_ALREADY_PROCESSING` |
| 15 | Refresh non-URL source | 400 Bad Request with `ERR_ONLY_URL_REFRESH` |
| 16 | Refresh cooldown active | 429 Too Many Requests with `Retry-After` |
| 17 | Storage service unavailable | 503 Service Unavailable |
| 38 | Cross-tenant access | 404 Not Found (source doesn't exist for this user) |
| 39 | Unauthenticated access | 401 Unauthorized |
| 40 | Invalid UUID format | 400 Bad Request |
| 41 | Invalid source type | 400 Bad Request |
| 42 | Missing required fields | 400 Bad Request |
| 43 | Method not allowed | 405 Method Not Allowed |
| 45 | Storage quota exceeded | 402 Payment Required |
| 46 | Invalid sitemap URL | 400 Bad Request |
| 47 | Unparseable sitemap XML | 400 Bad Request with parse error details |
| 48 | Sitemap with nested sitemap index | All child sitemaps are parsed and flattened |
| 49 | Bulk create with empty array | 400 Bad Request with `ERR_NO_URLS_PROVIDED` |
| 50 | Bulk create with invalid JSON | 400 Bad Request with `ERR_INVALID_REQUEST_BODY` |
| 58 | URL with fragment identifier | Fragment stripped during normalization |
| 59 | URL with query parameters | Query parameters preserved |
| 60 | URL with international characters | Normalized (Punycode for IDN) |
| 61 | URL with port number | Port preserved in stored URL |
| 62 | URL with trailing slash | Normalized (consistent trailing slash handling) |
| 63 | URL with uppercase host | Normalized to lowercase |
| 64 | PDF with non-PDF content-type | 400 Bad Request |
| 65 | Text with only whitespace | 400 Bad Request |
| 66 | File upload exceeds 52MB multipart limit | 400 Bad Request (multipart parse fails) |
