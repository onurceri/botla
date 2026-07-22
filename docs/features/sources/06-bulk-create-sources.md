> Extracted from: sources.md — Section: 2.5 Source Management — Bulk Create Sources

# 2.5 Source Management — Bulk Create Sources

- [ ] `POST /api/v1/chatbots/{id}/sources/bulk` accepts JSON body with `urls` array
- [ ] Empty `urls` array returns 400 Bad Request with `ERR_NO_URLS_PROVIDED`
- [ ] URL count per-bot limit is checked against current + requested count
- [ ] Monthly ingestion quota is checked
- [ ] Input is trimmed to available quota (both per-bot and monthly)
- [ ] Duplicate URLs are skipped (not counted as errors)
- [ ] Response includes `created_count`, `skipped_count`, and `errors` array
- [ ] Returns 201 Created if at least one source was created, 200 OK if all were skipped
