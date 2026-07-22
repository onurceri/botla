> Extracted from: background-processing.md — Section: 2.6 Background Processing — Duplicate Detection

# 2.6 Background Processing — Duplicate Detection

- [ ] URL sources: duplicate detection via normalized URL comparison (returns 409 `ERR_DUPLICATE_URL`)
- [ ] PDF sources: SHA-256 hash of file content (returns 409 `ERR_DUPLICATE_CONTENT`)
- [ ] Text sources: SHA-256 hash of text content (returns 409 `ERR_DUPLICATE_CONTENT`)
- [ ] URL refresh: SHA-256 hash of scraped content skips re-embedding if unchanged
- [ ] Bulk create: duplicate URLs are silently skipped (counted in `skipped_count`)
