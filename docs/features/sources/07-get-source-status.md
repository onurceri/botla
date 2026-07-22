> Extracted from: sources.md — Section: 2.6 Source Management — Get Source Status

# 2.6 Source Management — Get Source Status

- [ ] `GET /api/v1/sources/{id}` returns full source object with all fields
- [ ] ETag support: `If-None-Match` header returns 304 Not Modified when content unchanged
- [ ] ETag is computed from `status + processed_at + chunk_count`
- [ ] Response includes `Cache-Control: private, must-revalidate`
- [ ] Soft-deleted sources return 404 Not Found
- [ ] Cross-tenant access returns 404
