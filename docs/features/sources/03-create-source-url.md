> Extracted from: sources.md — Section: 2.2 Source Management — Create Source (URL)

# 2.2 Source Management — Create Source (URL)

- [ ] `POST /api/v1/chatbots/{id}/sources` with `source_type=url` and `source_url` creates a new URL source
- [ ] SSRF validation blocks private IP ranges (10.x, 172.16-31.x, 192.168.x, 127.x.x.x)
- [ ] URL is normalized (trailing slash handling) before storage
- [ ] Duplicate URL detection returns 409 Conflict
- [ ] Delete-and-readd cooldown check returns 429 Too Many Requests with `Retry-After` header
- [ ] URL count per-bot limit check returns 403 Forbidden with `ERR_URL_LIMIT_REACHED`
- [ ] Monthly ingestion quota check returns 402 Payment Required
- [ ] Source is created with `status: "pending"` and immediately enqueued for processing
- [ ] Response returns 201 Created with the new source ID
