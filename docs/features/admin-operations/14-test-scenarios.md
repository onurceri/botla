> Extracted from: admin-operations.md — Section: 3. Test Scenarios

# 3. Test Scenarios

### 3.1 Admin — Authentication & Authorization

**Scenario 3.1.1 — Non-admin cannot access admin endpoints**
- [ ] Authenticate as regular (non-platform-admin) user
- [ ] Attempt all `GET /api/v1/admin/*` endpoints
- [ ] Verify 403 Forbidden on all

**Scenario 3.1.2 — Unauthenticated access to admin endpoints**
- [ ] Send requests to admin endpoints without JWT cookie
- [ ] Verify 401 Unauthorized

**Scenario 3.1.3 — Platform admin can access all admin endpoints**
- [ ] Authenticate as platform admin
- [ ] Verify all admin endpoints return 200 (or appropriate non-403 status)

### 3.2 Admin — Users CRUD

**Scenario 3.2.1 — List users (default pagination)**
- [ ] Send `GET /api/v1/admin/users`
- [ ] Verify 200 OK with `users` array and `total`
- [ ] Verify default `limit=20`
- [ ] Verify users contain expected fields

**Scenario 3.2.2 — List users with filters**
- [ ] Send `GET /api/v1/admin/users?email=test@example.com`
- [ ] Verify filtered results match email
- [ ] Send `GET /api/v1/admin/users?is_platform_admin=true`
- [ ] Verify only admins returned
- [ ] Send `GET /api/v1/admin/users?plan_id=<plan_id>`
- [ ] Verify only users on that plan returned
- [ ] Combine filters — verify AND logic

**Scenario 3.2.3 — List users with custom pagination**
- [ ] Send `GET /api/v1/admin/users?limit=5&offset=10`
- [ ] Verify 5 users returned starting from offset 10
- [ ] Verify `total` reflects total count (not page size)

**Scenario 3.2.4 — Get user by ID**
- [ ] Send `GET /api/v1/admin/users/{valid_id}`
- [ ] Verify 200 OK with user details
- [ ] Send with non-existent UUID — verify 404

**Scenario 3.2.5 — Update user**
- [ ] Send `PATCH /api/v1/admin/users/{id}` with valid JSON body
- [ ] Verify 200 OK with `{"status": "updated"}`
- [ ] Verify changes persisted (fetch user to confirm)
- [ ] Verify audit log entry created for `update_user` action
- [ ] Send with empty body — verify 400 Bad Request
- [ ] Send with non-existent ID — verify 500 (sql.ErrNoRows not handled in UpdateUser — note: this uses `AdminUpdateUser` which may not return sql.ErrNoRows)

### 3.3 Admin — Organizations CRUD

**Scenario 3.3.1 — List organizations (default pagination)**
- [ ] Send `GET /api/v1/admin/organizations`
- [ ] Verify 200 OK with `organizations` array and `total`
- [ ] Verify default `limit=20`

**Scenario 3.3.2 — List organizations with filters**
- [ ] Filter by `name` — verify partial match
- [ ] Filter by `plan_id` — verify only orgs on that plan

**Scenario 3.3.3 — Get organization by ID**
- [ ] Send for valid ID — verify 200 with org details
- [ ] Send for non-existent ID — verify 404

### 3.4 Admin — Chatbots

**Scenario 3.4.1 — List chatbots (default pagination)**
- [ ] Send `GET /api/v1/admin/chatbots`
- [ ] Verify 200 OK with `chatbots` array and `total`

**Scenario 3.4.2 — List chatbots with filters**
- [ ] Filter by `name`, `organization_id`, `owner_id`
- [ ] Verify filtered results match

**Scenario 3.4.3 — Get chatbot by ID**
- [ ] Send for valid ID — verify 200 with chatbot details
- [ ] Send for non-existent — verify 404

**Scenario 3.4.4 — Force refresh chatbot**
- [ ] Send `POST /api/v1/admin/chatbots/{id}/force-refresh`
- [ ] Verify 200 OK with `status: "refreshing"`, `sources_reset`, `sources_queued`
- [ ] Verify all chatbot sources reset to `pending` status
- [ ] Verify Qdrant vectors deleted for chatbot
- [ ] Verify chunk counts reset to 0
- [ ] Verify sources queued for re-processing
- [ ] Verify audit log entry created for `force_refresh_chatbot`
- [ ] Send for non-existent chatbot — verify 404

### 3.5 Admin — Sources

**Scenario 3.5.1 — List sources (default pagination)**
- [ ] Send `GET /api/v1/admin/sources`
- [ ] Verify 200 OK with `sources` array and `total`

**Scenario 3.5.2 — List sources with filters**
- [ ] Filter by `chatbot_id`, `source_type`, `status`, `owner_id`
- [ ] Verify correct filtering

**Scenario 3.5.3 — Get source stats**
- [ ] Send `GET /api/v1/admin/sources/stats`
- [ ] Verify 200 OK with aggregated source statistics

**Scenario 3.5.4 — Get source types**
- [ ] Send `GET /api/v1/admin/sources/types`
- [ ] Verify 200 OK with `types: ["url", "file", "pdf", "text"]` and `statuses: ["pending", "processing", "ready", "failed"]`

**Scenario 3.5.5 — Get source by ID**
- [ ] Send for valid ID — verify 200 with source details
- [ ] Send for non-existent — verify 404

**Scenario 3.5.6 — Reprocess source**
- [ ] Send `POST /api/v1/admin/sources/{id}/reprocess`
- [ ] Verify 200 OK with `status: "reprocessing"`, `queued: true`
- [ ] Verify source status reset to `pending`
- [ ] Verify existing Qdrant vectors deleted for source
- [ ] Verify source queued for processing
- [ ] Verify audit log entry created
- [ ] Send for non-existent source — verify 404

### 3.6 Admin — Health Check

**Scenario 3.6.1 — Detailed health — all dependencies healthy**
- [ ] Ensure all services (Postgres, Redis, Qdrant, OpenAI, OpenRouter, Storage) are up
- [ ] Send `GET /api/v1/admin/health/detailed`
- [ ] Verify 200 OK
- [ ] Verify `status: "healthy"`
- [ ] Verify all dependencies have `status: "ok"`
- [ ] Verify each dependency includes `name`, `status`, `latency_ms`, `checked_at`
- [ ] Verify `version`, `uptime`, `environment`, `last_updated` present

**Scenario 3.6.2 — Health — degraded dependencies**
- [ ] Simulate high latency on one dependency (e.g., inject >100ms for Postgres)
- [ ] Send health check
- [ ] Verify `status: "degraded"`
- [ ] Verify problematic dependency shows `status: "degraded"`

**Scenario 3.6.3 — Health — non-critical dependency down (OpenAI/OpenRouter)**
- [ ] Revoke OpenAI API key or make endpoint unreachable
- [ ] Send health check
- [ ] Verify `status: "degraded"` (not unhealthy — non-critical)
- [ ] Verify OpenAI dependency shows `status: "down"`

**Scenario 3.6.4 — Health — critical dependency down (PostgreSQL)**
- [ ] Stop PostgreSQL
- [ ] Send health check
- [ ] Verify 503 Service Unavailable
- [ ] Verify `status: "unhealthy"`
- [ ] Verify Postgres shows `status: "down"`

**Scenario 3.6.5 — Health — Redis not configured**
- [ ] Set Redis client to nil/not configured
- [ ] Send health check
- [ ] Verify Redis shows `status: "down"` with message `not configured`

**Scenario 3.6.6 — Health — OpenRouter not configured**
- [ ] Clear `OPENROUTER_API_KEY` env var
- [ ] Send health check
- [ ] Verify OpenRouter shows `status: "down"` with message `API key not configured`

**Scenario 3.6.7 — Health — Storage not configured**
- [ ] Clear R2 settings (`R2_ACCOUNT_ID`, `R2_BUCKET_NAME`)
- [ ] Send health check
- [ ] Verify Storage shows `status: "down"` with message `storage not configured`

**Scenario 3.6.8 — Health caching**
- [ ] Send health check (populates cache with 1-minute TTL)
- [ ] Take down one dependency
- [ ] Send health check again without `?refresh=true`
- [ ] Verify stale cached result returned (not the actual down state)
- [ ] Send with `?refresh=true`
- [ ] Verify fresh (down) state returned

**Scenario 3.6.9 — Health check timeout**
- [ ] Simulate unresponsive dependency (e.g., Qdrant hangs)
- [ ] Send health check — verify context timeout at 10 seconds
- [ ] Verify overall status computed from available results

### 3.7 Admin — Queues

**Scenario 3.7.1 — Get queue stats**
- [ ] Send `GET /api/v1/admin/queues`
- [ ] Verify 200 OK with queue statistics

**Scenario 3.7.2 — Get stuck jobs (default threshold)**
- [ ] Send `GET /api/v1/admin/queues/stuck`
- [ ] Verify 200 OK with jobs processing > 30 minutes

**Scenario 3.7.3 — Get stuck jobs (custom threshold)**
- [ ] Send `GET /api/v1/admin/queues/stuck?threshold=5m`
- [ ] Verify jobs processing > 5 minutes
- [ ] Send with invalid duration — verify uses default 30m

**Scenario 3.7.4 — Retry job**
- [ ] Send `POST /api/v1/admin/queues/{id}/retry`
- [ ] Verify 200 OK
- [ ] Verify source updated for refresh
- [ ] Verify audit log created
- [ ] Send with non-existent ID — verify 500

**Scenario 3.7.5 — Delete job**
- [ ] Send `DELETE /api/v1/admin/queues/{id}`
- [ ] Verify 200 OK
- [ ] Verify source deleted from DB
- [ ] Verify audit log created
- [ ] Send with non-existent ID — verify 500

### 3.8 Admin — Errors

**Scenario 3.8.1 — List errors**
- [ ] Send `GET /api/v1/admin/errors`
- [ ] Verify 200 OK with `data` array, `total`, `page`, `per_page`
- [ ] Verify default `limit=20`

**Scenario 3.8.2 — List errors with severity filter**
- [ ] Send `GET /api/v1/admin/errors?severity=error`
- [ ] Verify only error-level entries returned
- [ ] Repeat for `warn`, `info`, `fatal`

**Scenario 3.8.3 — Get error stats**
- [ ] Send `GET /api/v1/admin/errors/stats`
- [ ] Verify 200 OK with summary statistics

**Scenario 3.8.4 — Get error detail**
- [ ] Send for valid error ID — verify 200 with full error detail
- [ ] Send for non-existent — verify 404

### 3.9 Admin — Audit Logs

**Scenario 3.9.1 — List audit logs**
- [ ] Send `GET /api/v1/admin/audit-logs`
- [ ] Verify 200 OK with `data`, `total`, `page`, `per_page`
- [ ] Verify default `limit=50`

**Scenario 3.9.2 — List audit logs with filters**
- [ ] Filter by `admin_user_id` — verify only actions by that admin
- [ ] Filter by `action` — verify only matching actions
- [ ] Filter by `target_type` — verify only matching target types
- [ ] Filter by date range (`start_date`, `end_date` in RFC3339) — verify correct range
- [ ] Combine multiple filters — verify AND logic

**Scenario 3.9.3 — Audit log entries created by admin actions**
- [ ] Perform admin action (update user, reprocess source, etc.)
- [ ] Verify corresponding audit log entry exists with correct `action`, `target_type`, `target_id`, `details`, `ip_address`, `user_agent`

### 3.10 Admin — Privacy (KVKK)

**Scenario 3.10.1 — List privacy requests**
- [ ] Send `GET /api/v1/admin/privacy/requests`
- [ ] Verify 200 OK with `data`, `total`, `page`, `per_page`
- [ ] Verify default `page=1`, `limit=20`

**Scenario 3.10.2 — List privacy requests with status filter**
- [ ] Filter by `status=pending` — verify only pending
- [ ] Filter by `status=completed` — verify only completed
- [ ] Filter by `status=denied` — verify only denied
- [ ] Filter by `status=processing` — verify only processing

**Scenario 3.10.3 — Get privacy request by ID**
- [ ] Send for valid ID — verify 200 with request details
- [ ] Send for non-existent — verify 404

**Scenario 3.10.4 — Approve export request**
- [ ] Create pending export request
- [ ] Send `PATCH /api/v1/admin/privacy/requests/{id}` with `{"action": "approve"}`
- [ ] Verify 200 OK
- [ ] Verify `PrivacyService.ProcessExportRequest` called
- [ ] Verify audit log entry `process_privacy_request`

**Scenario 3.10.5 — Approve deletion request**
- [ ] Create pending deletion request
- [ ] Send `PATCH /api/v1/admin/privacy/requests/{id}` with `{"action": "approve"}`
- [ ] Verify 200
- [ ] Verify `PrivacyService.ProcessDeletion` called

**Scenario 3.10.6 — Deny privacy request**
- [ ] Send `PATCH /api/v1/admin/privacy/requests/{id}` with `{"action": "deny", "denial_reason": "Insufficient information"}`
- [ ] Verify 200
- [ ] Verify status changed to `denied`
- [ ] Verify denial reason stored
- [ ] Verify audit log created

**Scenario 3.10.7 — Process privacy request — invalid action**
- [ ] Send with `{"action": "invalid"}`
- [ ] Verify 400 Bad Request
- [ ] Send with empty body — verify 400

**Scenario 3.10.8 — Download privacy export (success)**
- [ ] Create completed export request with valid export URL
- [ ] Send `GET /api/v1/admin/privacy/requests/{id}/download`
- [ ] Verify 200 with file content
- [ ] Verify `Content-Type: application/json`
- [ ] Verify `Content-Disposition: attachment` header
- [ ] Verify `Cache-Control: no-store`

**Scenario 3.10.9 — Download privacy export — not export type**
- [ ] Create deletion request
- [ ] Download — verify 400 Bad Request

**Scenario 3.10.10 — Download privacy export — not completed**
- [ ] Create pending export request
- [ ] Download — verify 400 Bad Request

**Scenario 3.10.11 — Download privacy export — expired**
- [ ] Create completed export with past expiry
- [ ] Download — verify 410 Gone

**Scenario 3.10.12 — Get signed download URL**
- [ ] Create completed export request with valid export URL
- [ ] Send `GET /api/v1/admin/privacy/requests/{id}/download-url`
- [ ] Verify 200 with `url` containing signed URL
- [ ] Verify signed URL expires in 1 hour

**Scenario 3.10.13 — Generate user export**
- [ ] Send `POST /api/v1/admin/privacy/export/{userId}`
- [ ] Verify 200 OK with export object
- [ ] Verify audit log `generate_user_export` created

**Scenario 3.10.14 — Download data export**
- [ ] Send `GET /api/v1/admin/privacy/exports/{id}/download` for completed export
- [ ] Verify 200 with file content
- [ ] Send for non-completed export — verify 400
- [ ] Send for expired export — verify 410
- [ ] Send for non-existent — verify 404

### 3.11 Admin — Plans Management

**Scenario 3.11.1 — List plans (admin)**
- [ ] Send `GET /api/v1/admin/plans`
- [ ] Verify 200 OK with `plans` array and `total`
- [ ] Verify each plan summary includes: `id`, `code`, `name`, `status`, `billing_cycle`, `price`, `currency`, `trial_days`, `max_chatbots`, `max_monthly_ingestions`, `files_max_size_mb`, `chat_allowed_models_count`

**Scenario 3.11.2 — Get plan detail (admin)**
- [ ] Send `GET /api/v1/admin/plans/{id}`
- [ ] Verify 200 OK with `plan`, `limits`, and `translations`
- [ ] Verify `plan` contains full plan info
- [ ] Verify `limits` contains all 50+ PlanLimitsDetail fields
- [ ] Send for non-existent plan — verify 404

**Scenario 3.11.3 — Update plan limits (partial)**
- [ ] Send `PATCH /api/v1/admin/plans/{id}/limits` with `{"max_chatbots": 50}`
- [ ] Verify 200 OK with `status: "updated"`
- [ ] Verify only `max_chatbots` changed (other limits unchanged)
- [ ] Verify cache invalidated for this plan
- [ ] Fetch `GET /api/v1/plans/{code}` — verify new limit reflected

**Scenario 3.11.4 — Update plan limits — multiple fields**
- [ ] Send update with changes to multiple categories simultaneously
- [ ] Verify all changes applied correctly
- [ ] Verify cache invalidated

**Scenario 3.11.5 — Update plan limits — boolean toggles**
- [ ] Toggle `scraping_dynamic_enabled`, `refresh_enabled`, `security_secure_embed_enabled`, `branding_can_hide_branding`
- [ ] Verify each toggle reflected in API response

**Scenario 3.11.6 — Update plan limits — string/slice fields**
- [ ] Update `chat_default_model` — verify new default model
- [ ] Update `chat_allowed_models` — verify new model list

**Scenario 3.11.7 — Update plan limits — empty body**
- [ ] Send PATCH with `{}` — verify 400 Bad Request

**Scenario 3.11.8 — Update plan limits — non-existent plan**
- [ ] Send PATCH for non-existent ID — verify 404

**Scenario 3.11.9 — Invalidate plan cache**
- [ ] Send `POST /api/v1/admin/plans/{id}/cache-invalidate`
- [ ] Verify 200 OK with `status: "cache_invalidated"`
- [ ] Send for non-existent plan — verify 404

### 3.12 Admin — Model Registry

**Scenario 3.12.1 — Get available models by plan**
- [ ] User on Free plan (allowed: `gpt-6-luna`, `gpt-4o-mini`) — verify both returned
- [ ] User on Pro plan (allowed: `gpt-6-luna`, `gpt-4o-mini`, `gpt-4o`) — verify all three returned
- [ ] User on Ultra plan (allowed: `gpt-6-luna`, `gpt-4o-mini`, `gpt-4o`, `gpt-5`) — verify all four returned

**Scenario 3.12.2 — Resolve API model ID**
- [ ] Call `GetAPIModelID("gpt-4o-mini")` — verify returns `openai/gpt-4o-mini`
- [ ] Call with non-existent model — verify error

**Scenario 3.12.3 — Empty allowed models returns all active models**
- [ ] Call `GetAvailableModels(ctx, [])` — verify returns all active models from DB

**Scenario 3.12.4 — Inactive models excluded**
- [ ] Mark a model as `is_active = false` in `ai_models` table
- [ ] Call `GetAvailableModels` — verify inactive model not returned even if in allowed list

---

