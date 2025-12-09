# 14. Scraper Tests

> **Priority**: Medium  
> **Test Count**: 12  
> **Source Files**: `internal/scraper/`

---

## 14.1 Path Filtering

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| PTH-001 | Include path `/blog/*` | Only blog paths | ✅ |
| PTH-002 | Exclude path `/admin/*` | Admin excluded | ✅ |
| PTH-003 | Glob pattern matching | Wildcards work | ✅ |
| PTH-004 | Path normalization | Trailing slashes | ✅ |

---

## 14.2 CSS Selector Extraction

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| CSS-001 | Single selector `article` | Article content only | ✅ |
| CSS-002 | Multiple selectors | Multiple sections | ✅ |
| CSS-003 | Invalid selector | Fallback to full page | ✅ |
| CSS-004 | Nested selectors `.content > p` | Nested works | ✅ |

---

## 14.3 Dynamic Content

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| DYN-001 | JavaScript-rendered page | Rod browser fallback | ✅ |
| DYN-002 | Headless browser timeout | Error handled | ✅ |
| DYN-003 | Encoding detection | UTF-8 from meta tag | ✅ |
| DYN-004 | Caching HTML content | Cache hit/miss | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/scraper/path_filter_test.go` | Path filtering |
| `internal/scraper/selector_extractor_test.go` | CSS selectors |
| `internal/scraper/sitemap_parser_test.go` | Sitemaps |
| `internal/scraper/browser_test.go` | Dynamic content (JS, Timeouts) |
| `internal/scraper/encoding_test.go` | Text encoding |
| `internal/scraper/cache_test.go` | Caching |
