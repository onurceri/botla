# 08. Source Ingestion Tests

> **Priority**: Critical  
> **Test Count**: 22  
> **Source Files**: `internal/api/handlers/source*.go`, `internal/scraper/`, `internal/processing/`

---

## 8.1 Text Upload

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| SRC-001 | Upload small text (<1KB) | Stored and processed | ✅ |
| SRC-002 | Upload with Turkish content | Chars preserved | ✅ |
| SRC-003 | Upload PDF (with fitz) | Text extracted | ✅ |
| SRC-004 | Upload PDF (OCR fallback) | Text extracted via Tesseract | ✅ |
| SRC-005 | Upload unsupported file type | 400 Bad Request | ✅ |
| SRC-006 | Upload exceeding size limit | 413/402 error | ✅ |

---

## 8.2 URL Ingestion

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| URL-001 | Ingest valid static page | Content extracted | ✅ |
| URL-002 | Ingest dynamic page (JS) | Browser fallback works | ✅ |
| URL-003 | Ingest with path filter (include) | Only matching paths | ✅ |
| URL-004 | Ingest with path filter (exclude) | Excluded paths skipped | ✅ |
| URL-005 | Ingest with CSS selector | Only selected content | ✅ |
| URL-006 | Ingest 404 URL | Error handled | ✅ |
| URL-007 | Ingest with timeout | Error message returned | ✅ |
| URL-008 | Duplicate URL detection | 409 or skip | ✅ |

---

## 8.3 Sitemap Import

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| SMP-001 | Parse valid XML sitemap | URLs extracted | ✅ |
| SMP-002 | Parse sitemap index | All sitemaps processed | ✅ |
| SMP-003 | Parse gzipped sitemap | Decompressed | ✅ |
| SMP-004 | Invalid sitemap URL | Error returned | ✅ |
| SMP-005 | Bulk import from sitemap | Sources created | ✅ |

---

## 8.4 Refresh & Discovery

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| RFR-001 | Manual refresh URL source | Re-fetched | ✅ |
| RFR-002 | Refresh unchanged (hash match) | Skip re-embedding | ✅ |
| RFR-003 | Refresh cooldown active | Error returned | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/integration/sources_test.go` | CRUD |
| `internal/integration/source_refresh_test.go` | Refresh |
| `internal/integration/source_refresh_hash_test.go` | Refresh optimization (hash) |
| `internal/integration/source_size_limit_test.go` | File size limits |
| `internal/scraper/sitemap_parser_test.go` | Sitemap |
| `internal/scraper/sitemap_gzip_test.go` | Sitemap Gzip |
