> Extracted from: training-jobs.md — Section: 2.3 Pending URLs — List

# 2.3 Pending URLs — List

- [ ] `GET /api/v1/chatbots/{id}/pending-urls` returns paginated list of discovered URLs awaiting approval
- [ ] Supports `page` and `per_page` query parameters (per_page max 100)
- [ ] Response includes `urls` array, `total`, `page`, `per_page`
- [ ] Each URL entry includes: id, url, discovered_at
- [ ] Empty list returns `urls: []` with total=0
