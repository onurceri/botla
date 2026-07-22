> Extracted from: plans-entitlements.md — Section: 3. Test Scenarios

# 3. Test Scenarios

### 3.1 Plan Listing & Retrieval

**Scenario 3.1.1 — List all plans (unauthenticated)**
- [ ] Send `GET /api/v1/plans`
- [ ] Verify response is 200 OK
- [ ] Verify response is a JSON array
- [ ] Verify each plan object contains `code`, `price`, `currency`, `limits`, `features`
- [ ] Verify limits include `max_chatbots`, `max_monthly_ingestions`, `max_monthly_embedding_tokens`
- [ ] Verify features include `scraping`, `files`, `chat`, `refresh`, `security`, `guardrails`, `branding`, `rate_limits`
- [ ] Verify Free plan shows `chat_default_model: "openai/gpt-4o-mini"`
- [ ] Verify Pro plan shows 2 allowed models (`gpt-4o-mini`, `gpt-4o`)
- [ ] Verify Ultra plan shows 3 allowed models (`gpt-4o-mini`, `gpt-4o`, `gpt-5`)
- [ ] Verify all plans are cached after first request (subsequent request does not hit DB)

**Scenario 3.1.2 — Get plan by code (valid)**
- [ ] Send `GET /api/v1/plans/free`
- [ ] Verify 200 OK with `code: "free"`
- [ ] Verify limits match `DefaultPlanLimits()` values
- [ ] Repeat for `pro` and `ultra` codes
- [ ] Verify plan cached individually

**Scenario 3.1.3 — Get plan by code (invalid)**
- [ ] Send `GET /api/v1/plans/nonexistent`
- [ ] Verify 404 Not Found
- [ ] Send `GET /api/v1/plans/` (empty code)
- [ ] Verify 400 Bad Request

**Scenario 3.1.4 — Get current user's plan**
- [ ] Authenticate as user on Free plan
- [ ] Send `GET /api/v1/me/plan`
- [ ] Verify 200 OK with `code: "free"`
- [ ] Verify response includes `limits`, `features`, `available_models`
- [ ] Verify `available_models` contains only Free-allowed models
- [ ] Verify `features.rate_limits.endpoints` includes `chat` and `sources` with correct RPM/Windows
- [ ] Repeat with user on Pro plan — verify different limits + additional models
- [ ] Repeat with user on Ultra plan — verify highest limits + GPT-5 availability

**Scenario 3.1.5 — Get plan for unauthenticated user**
- [ ] Send `GET /api/v1/me/plan` without JWT cookie
- [ ] Verify 401 Unauthorized

**Scenario 3.1.6 — Get plan for user with no plan assigned**
- [ ] Create user with `plan_id = NULL`
- [ ] Authenticate as that user
- [ ] Send `GET /api/v1/me/plan`
- [ ] Verify 404 Not Found

### 3.2 Usage Tracking

**Scenario 3.2.1 — Get current usage**
- [ ] Authenticate as user with existing chat activity
- [ ] Send `GET /api/v1/me/usage`
- [ ] Verify 200 OK with usage stats
- [ ] Verify token counts are accurate

**Scenario 3.2.2 — Usage for zero-activity user**
- [ ] Authenticate as freshly created user with no activity
- [ ] Send `GET /api/v1/me/usage`
- [ ] Verify 200 OK with zero/empty values (no error)

### 3.3 Plan Defaults & Fallback

**Scenario 3.3.1 — Plan limits with zero values use defaults**
- [ ] Configure a plan with `FilesMaxFilesTotal = 0`, `FilesMaxTextLength = 0`, `ChatMinResponseTokenLimit = 0`, `ChatMaxResponseTokenLimit = 0`
- [ ] Fetch `GET /api/v1/me/plan` for user on that plan
- [ ] Verify `applyLimitsDefaults` applies: `FilesMaxFilesTotal` mapped by plan code (Free=5, Pro=100, Ultra=1000), `FilesMaxTextLength = 400000`, `MinResponseTokenLimit != 0`, `MaxResponseTokenLimit != 0`

### 3.4 Plan Validation at Startup

**Scenario 3.4.1 — Valid plans pass startup validation**
- [ ] Ensure all plan limits pass `PlanLimits.Validate()`
- [ ] Call `PlanService.ValidateAllPlans`
- [ ] Verify returns nil

**Scenario 3.4.2 — Invalid plan fails validation**
- [ ] Set `MaxChatbots = 0` on a plan
- [ ] Call `PlanService.ValidateAllPlans`
- [ ] Verify returns error: `max_chatbots must be >= 1`
- [ ] Set `FilesMaxSizeMB = 0`
- [ ] Verify returns error: `files_max_size_mb must be > 0`
- [ ] Set `RateLimitsRequestsPerMinute = 0`
- [ ] Verify returns error: `rate_limits_requests_per_minute must be >= 1`
- [ ] Set `ChatMaxResponseTokenLimit < ChatMinResponseTokenLimit`
- [ ] Verify returns error about min/max mismatch
- [ ] Set multiple invalid fields simultaneously
- [ ] Verify joined error contains all validation failures

**Scenario 3.4.3 — Missing limits on plan**
- [ ] Create plan with `Limits = nil`
- [ ] Call `PlanService.ValidateAllPlans`
- [ ] Verify error: `plan "<code>": missing limits`

**Scenario 3.4.4 — Empty plans list (no plans in DB)**
- [ ] Truncate plans table
- [ ] Call `PlanService.ValidateAllPlans`
- [ ] Verify returns nil (no plans = no validation errors)

### 3.5 Cache Invalidation

**Scenario 3.5.1 — Invalidate single plan cache**
- [ ] Fetch plan via `GET /api/v1/plans/pro` (populates cache)
- [ ] Modify plan limits directly in DB
- [ ] Call `POST /api/v1/admin/plans/{id}/cache-invalidate`
- [ ] Verify 200 OK
- [ ] Fetch plan again — verify updated limits returned (cache miss → DB fetch)

**Scenario 3.5.2 — Invalidate all plans cache after limit update**
- [ ] Fetch `GET /api/v1/plans` (populates `plans:all` cache)
- [ ] Update a plan's limits via `PATCH /api/v1/admin/plans/{id}/limits`
- [ ] Verify response includes cache invalidation
- [ ] Fetch `GET /api/v1/plans` again — verify updated values
- [ ] Fetch specific plan — verify updated values

**Scenario 3.5.3 — Cache expiry (24h TTL)**
- [ ] Mock time to simulate 24h+ passage
- [ ] Verify cache entry expired and next request hits DB

**Scenario 3.5.4 — Redis unavailable — memory cache fallback**
- [ ] Shut down Redis
- [ ] Fetch plans — verify still returns data (memory cache)
- [ ] Wait for memory cache expiry
- [ ] Fetch plans — verify still returns data (falls through to DB)

### 3.6 Quota Enforcement

**Scenario 3.6.1 — Token reservation succeeds within quota**
- [ ] User has `ChatMaxMonthlyTokens = 100000`
- [ ] Call `QuotaEnforcer.ReserveTokens(ctx, userID, 1000, 100000)`
- [ ] Verify no error
- [ ] Verify usage repo `ReserveChatTokens` was called with correct params

**Scenario 3.6.2 — Token reservation exceeds quota**
- [ ] User has already used 95,000 of 100,000 monthly tokens
- [ ] Call `QuotaEnforcer.ReserveTokens(ctx, userID, 10000, 100000)`
- [ ] Verify error returned (quota exceeded)

**Scenario 3.6.3 — No quota enforcement when limit is zero**
- [ ] Set `ChatMaxMonthlyTokens = 0`
- [ ] Call `QuotaEnforcer.ReserveTokens(ctx, userID, 1000000, 0)`
- [ ] Verify nil (no enforcement)

**Scenario 3.6.4 — Token adjustment after completion**
- [ ] Reserve 1000 tokens
- [ ] Call `QuotaEnforcer.AdjustTokens(ctx, userID, 1000, 800)`
- [ ] Verify usage adjusted by -200 (refund)
- [ ] Call `QuotaEnforcer.AdjustTokens(ctx, userID, 1000, 1200)`
- [ ] Verify usage adjusted by +200 (overage)

**Scenario 3.6.5 — Token refund on error**
- [ ] Reserve 500 tokens
- [ ] Call `QuotaEnforcer.RefundTokens(ctx, userID, 500)`
- [ ] Verify usage adjusted by -500

**Scenario 3.6.6 — Chatbot creation blocked by plan limit**
- [ ] User on Free plan (MaxChatbots = 1) with 1 existing chatbot
- [ ] Attempt to create 2nd chatbot
- [ ] Verify creation fails with quota/plan limit error

**Scenario 3.6.7 — Chatbot creation allowed within limit**
- [ ] User on Pro plan (MaxChatbots = 10) with 3 existing chatbots
- [ ] Attempt to create 4th chatbot
- [ ] Verify creation succeeds

**Scenario 3.6.8 — File upload blocked by plan limits**
- [ ] User on Free plan (FilesMaxFilesPerBot = 1) with 1 existing file on bot
- [ ] Attempt to upload 2nd file — verify failure
- [ ] Attempt to upload file exceeding FilesMaxSizeMB (5MB) — verify failure
- [ ] Attempt to upload when FilesTotalStorageMB (10MB) would be exceeded — verify failure

**Scenario 3.6.9 — Rate limit enforcement**
- [ ] Send requests at a rate exceeding plan's `RateLimitsRequestsPerMinute`
- [ ] Verify 429 Too Many Requests
- [ ] Verify chat endpoint respects `RateLimitsChatRPM`
- [ ] Verify sources endpoint respects `RateLimitsSourcesRPM`
- [ ] Verify rate limit resets after `RateLimitsWindowSeconds`

---

