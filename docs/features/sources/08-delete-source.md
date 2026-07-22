> Extracted from: sources.md — Section: 2.7 Source Management — Delete Source

# 2.7 Source Management — Delete Source

- [ ] `DELETE /api/v1/sources/{id}` soft-deletes the source (sets `deleted_at`)
- [ ] Associated vectors are deleted from Qdrant (best-effort, logged on failure)
- [ ] Associated file is deleted from R2/S3 storage (if file-based source)
- [ ] Suggestion re-aggregation is triggered in a background goroutine (panic-recovered)
- [ ] Returns 204 No Content on success
- [ ] Already-deleted source returns 404
