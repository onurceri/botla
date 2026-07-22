> Extracted from: training-jobs.md — Section: 2.5 Pending URLs — Reject

# 2.5 Pending URLs — Reject

- [ ] `POST /api/v1/chatbots/{id}/pending-urls/reject` accepts `url_ids` array
- [ ] Pending URL status is updated to `rejected`
- [ ] Response includes `rejected_count`
- [ ] Empty `url_ids` returns 400 Bad Request
