> Extracted from: training-jobs.md — Section: 4. Test Scenarios

# 4. Test Scenarios

### 4.1 URL Discovery — Pending Mode

- [ ] Create a chatbot with `discovery_mode: "pending"`
- [ ] Add a URL source that links to other pages
- [ ] After processing, `GET /api/v1/chatbots/{id}/pending-urls` shows discovered URLs
- [ ] `POST /api/v1/chatbots/{id}/pending-urls/approve` with selected URL IDs
- [ ] Verify sources are created and enqueued for processing
- [ ] `POST /api/v1/chatbots/{id}/pending-urls/reject` with remaining URL IDs
- [ ] Verify rejected URLs no longer appear in pending list
- [ ] `DELETE /api/v1/chatbots/{id}/pending-urls/clear` removes all remaining pending URLs

### 4.2 Training Job — Retry Failed Job

- [ ] Create a source with an invalid URL (e.g., `https://nonexistent.invalid`)
- [ ] Wait for processing to fail
- [ ] `GET /api/v1/sources/{id}/job` shows `status: "failed"`, `error_code`, `error_message`, `failed_step`
- [ ] `POST /api/v1/sources/{id}/job/retry`
- [ ] Verify 202 Accepted
- [ ] Poll job until it fails again or completes
- [ ] Verify `retry_count` is incremented

### 4.3 Training Job — Retry Non-Failed Job

- [ ] Create a source and immediately try to retry while pending
- [ ] `POST /api/v1/sources/{id}/job/retry`
- [ ] Verify 400 Bad Request with `ERR_JOB_NOT_FAILED`

### 4.4 Pending URLs — Pagination

- [ ] Configure a chatbot with `discovery_mode: "pending"` and add a URL with many sub-pages
- [ ] `GET /api/v1/chatbots/{id}/pending-urls?page=1&per_page=5` — verify 5 URLs returned
- [ ] `GET /api/v1/chatbots/{id}/pending-urls?page=2&per_page=5` — verify next 5 URLs
- [ ] Verify `total` reflects the full count
- [ ] `GET /api/v1/chatbots/{id}/pending-urls?per_page=200` — verify capped at 100

### 4.5 Cross-Tenant — Pending URLs

- [ ] Create Chatbot A (User A) with pending URLs
- [ ] Authenticate as User B
- [ ] `GET /api/v1/chatbots/{chatbotA_id}/pending-urls` — verify 404
- [ ] `POST /api/v1/chatbots/{chatbotA_id}/pending-urls/approve` — verify 404
- [ ] `POST /api/v1/chatbots/{chatbotA_id}/pending-urls/reject` — verify 404
- [ ] `DELETE /api/v1/chatbots/{chatbotA_id}/pending-urls/clear` — verify 404

### 4.6 Pending URLs — Approve Creates Discovered Sources

- [ ] Configure chatbot with `discovery_mode: "pending"`
- [ ] Add a URL source, wait for discovery
- [ ] `POST /api/v1/chatbots/{id}/pending-urls/approve` with discovered URL IDs
- [ ] Verify created sources have `is_discovered: true`
- [ ] Verify created sources are enqueued and processed
- [ ] Verify processed discovered sources do NOT crawl further

### 4.7 Pending URLs — Reject and Clear

- [ ] Discover pending URLs
- [ ] Reject some URLs — verify they no longer appear in pending list
- [ ] Clear all remaining — verify `cleared_count` matches
- [ ] Verify pending list is now empty

### 4.8 Training Job — No Job Exists

- [ ] Create a source and immediately `GET /api/v1/sources/{id}/job` before processing starts
- [ ] Verify response maps source status to job status (no `job_id` field)
- [ ] Verify `progress_percent` reflects source status

### 4.9 Training Job — Retry Non-Existent

- [ ] `POST /api/v1/sources/non-existent-id/job/retry`
- [ ] Verify 404 Not Found

### 4.10 Pending URLs — Approve with Non-Existent IDs

- [ ] `POST /api/v1/chatbots/{id}/pending-urls/approve` with `{"url_ids": ["non-existent-id"]}`
- [ ] Verify 200 OK with `approved_count: 0`, `sources_created: 0`

### 4.11 Pending URLs — Reject with Non-Existent IDs

- [ ] `POST /api/v1/chatbots/{id}/pending-urls/reject` with `{"url_ids": ["non-existent-id"]}`
- [ ] Verify 200 OK with `rejected_count: 0`

### 4.12 Pending URLs — Clear Empty

- [ ] `DELETE /api/v1/chatbots/{id}/pending-urls/clear` when no pending URLs exist
- [ ] Verify 200 OK with `cleared_count: 0`

### 4.13 Pending URLs — Approve with Duplicate URLs

- [ ] Discover pending URLs
- [ ] Manually create a source with one of the discovered URLs
- [ ] Approve the same URL from pending list
- [ ] Verify the duplicate is skipped (source creation fails silently, logged as warning)
- [ ] Verify `sources_created` is less than `approved_count`

### 4.14 Pending URLs — Approve All

- [ ] Discover 10 pending URLs
- [ ] `POST /api/v1/chatbots/{id}/pending-urls/approve` with all 10 URL IDs
- [ ] Verify all 10 are approved and sources are created
- [ ] Verify all 10 sources are enqueued and processed

### 4.15 Pending URLs — Reject All

- [ ] Discover 5 pending URLs
- [ ] `POST /api/v1/chatbots/{id}/pending-urls/reject` with all 5 URL IDs
- [ ] Verify all 5 are rejected
- [ ] Verify pending list is now empty
