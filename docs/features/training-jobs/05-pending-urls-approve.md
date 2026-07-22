> Extracted from: training-jobs.md — Section: 2.4 Pending URLs — Approve

# 2.4 Pending URLs — Approve

- [ ] `POST /api/v1/chatbots/{id}/pending-urls/approve` accepts `url_ids` array in JSON body
- [ ] Creates a `DataSource` for each approved URL with `is_discovered: true` (1-level depth limit)
- [ ] Each created source is enqueued for processing
- [ ] Pending URL status is updated to `selected`
- [ ] Response includes `approved_count` and `sources_created`
- [ ] Empty `url_ids` returns 400 Bad Request
