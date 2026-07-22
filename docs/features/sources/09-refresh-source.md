> Extracted from: sources.md — Section: 2.8 Source Management — Refresh Source

# 2.8 Source Management — Refresh Source

- [ ] `POST /api/v1/sources/{id}/refresh` re-ingests a URL source
- [ ] Only URL sources can be refreshed (PDF/text returns 400 Bad Request with `ERR_ONLY_URL_REFRESH`)
- [ ] Already-processing source returns 409 Conflict with `ERR_SOURCE_ALREADY_PROCESSING`
- [ ] Plan-level refresh feature gate check returns 403 Forbidden with `ERR_PLAN_REFRESH_UNAVAILABLE`
- [ ] Monthly refresh quota check returns 402 Payment Required with `ERR_MONTHLY_REFRESH_EXCEEDED`
- [ ] Cooldown period check returns 429 Too Many Requests with `Retry-After` header
- [ ] Source is re-enqueued for processing; old vectors are deleted before re-embedding
- [ ] Content hash comparison skips re-embedding if content unchanged (returns `Skipped: true`)
- [ ] Returns 202 Accepted with source ID
