## Plan Limits, Secure Embed, and Analytics – Current State

### Plan Configuration and Enforcement
- Plans (`free`, `pro`, `ultra`) store limits in `plans.config` as `PlanConfig` (`internal/models/plan.go:22`), extended by migrations:
  - Base scraping/files/chat config: `db/migrations/000003_update_plan_configs.up.sql:8`
  - File count limits (`max_files_total`): `db/migrations/000004_add_max_files_total.up.sql:3`
  - Ingestion and embedding limits (`max_monthly_ingestions`, `max_monthly_embedding_tokens`, `min_readd_cooldown_minutes`): `db/migrations/000006_extend_plan_config_ingestions.up.sql:1`
  - Refresh limits per plan: `db/migrations/000007_add_refresh_tracking.up.sql:5`
  - Guardrails and max_chatbots: `db/migrations/000025_update_plan_guardrails.up.sql:1`
  - Secure embed availability: `db/migrations/000023_secure_embed_config.up.sql:1`
- The `/api/v1/me` handler exposes plan + usage:
  - `planInfo` and `getPlanInfo` load `PlanConfig` and apply defaults (`internal/api/handlers/me.go:43`, `97`).
  - `getUserUsage` aggregates usage (files, URLs, tokens, ingestions, embedding tokens, refresh count) from DB (`internal/api/handlers/me.go:144`).
- Frontend `PlanPage` reads `config` and `usage` and visualizes:
  - Token, file, storage, URL, ingestion, embedding, and refresh limits vs usage (`frontend/src/pages/PlanPage.tsx:197`, `226`, `271`, `319`, `372`).

### Quotas and Usage Tracking
- Monthly chat token usage:
  - Aggregated from `analytics.total_tokens_used` by user: `internal/db/analytics.go:44`.
  - Written via `db.IncrementAnalytics` from chat service (`internal/services/chat_service.go:271`).
  - Enforced in authenticated chat: `internal/api/handlers/chat.go:73-104`.
  - Enforced in public chat (widget): `internal/api/handlers/public.go:229-253`.
- Monthly ingestion and embedding tokens:
  - Stored per user in `usage_ingestions` with `sources_count` and `embedding_tokens` (`internal/db/usage_ingestions.go:1`).
  - `IncrementSuccessfulIngestion` and `AddEmbeddingTokens` are called at the end of URL processing (`internal/processing/url_processor.go:173-196`).
  - Quota enforcement:
    - Single-source create: `SourcesHandlers.checkIngestionQuota` uses `plan.Config.MaxMonthlyIngestions` and `GetMonthlyIngestionUsage` (`internal/api/handlers/source_utils.go:18`).
    - Bulk URL create additionally caps URLs by bot URL limit and remaining monthly ingestions (`internal/api/handlers/source_bulk.go:66-104`).

### Secure Embed – Plan Gating and Public Enforcement
- Plan-driven availability:
  - `plans.config.security.secure_embed_enabled` is set per plan (`db/migrations/000023_secure_embed_config.up.sql:1`).
  - `PlanConfig.Security` is modeled in `internal/models/plan.go:22`.
  - Chatbot update validates secure embed against the user’s plan:
    - If `secure_embed_enabled` is requested but plan `Security.SecureEmbedEnabled` is false, returns 403 with `feature: "secure_embed"` (`internal/api/handlers/chatbot_item.go:139-151`).
- Backend enforcement at public chat endpoint:
  - When `cbot.SecureEmbedEnabled` is true, `internal/api/handlers/public.go` enforces:
    - Allowed domains via `Origin` header and comma-separated `AllowedDomains` (`internal/api/handlers/public.go:193-205`).
    - Embed token via `X-Embed-Token` signed with `EmbedSecret` and containing `chatbot_id` (`internal/api/handlers/public.go:205-246`).
  - Failure behavior:
    - Missing or invalid token → 401.
    - Valid token but origin not in `allowed_domains` → 403.
- Widget behavior:
  - The widget fetches the embed token (and optional CAPTCHA token) before POSTing chat:
    - `widget/src/widgetApp.tsx:91-126` handles `embedToken`, `embedTokenUrl`, and `captchaSiteKey`.
  - Frontend hides secure embed UI for free plans and shows it for Pro/Ultra, aligned with plan code.
- Tests:
  - `internal/integration/chatbot_secure_embed_test.go:8-78` verifies enabling secure embed on a chatbot and persistence via GET.
  - `internal/integration/public_secure_embed_test.go:1-120` verifies:
    - 401 without token
    - 401 with invalid token
    - 200 with valid token and allowed origin
    - 403 with valid token but disallowed origin
  - `frontend/e2e/widget-embed-secure.spec.ts:29-50` exercises the secure embed widget with auto-open, embed token URL, and CAPTCHA stub.

### Analytics – Overview, Trends, Source Usage, Unanswered
- Schema extensions:
  - `analytics.total_tokens_used`, `handoff_count`, `avg_response_time_ms` and indexes for trend queries: `db/migrations/000020_advanced_analytics.up.sql:1`.
  - Message-level columns `messages.confidence_score` and `messages.sources_used`: same migration.
  - `unanswered_queries` table for low-confidence questions: `db/migrations/000020_advanced_analytics.up.sql:13`.
- Overview (per chatbot, last 30 days):
  - `db.GetAnalyticsOverview` sums messages, conversations, tokens, thumbs up/down, handoffs and derives `FeedbackRate` as satisfaction (`internal/db/analytics.go:111-145`).
  - Model: `models.AnalyticsOverview` (`internal/models/analytics.go:13`).
  - Service: `AnalyticsService.GetChatbotOverview` (`internal/services/analytics_service.go:22`).
- Trends:
  - `db.GetAnalyticsTrends` returns daily `AnalyticsPoint` (messages, conversations, tokens, thumbs_up, thumbs_down, handoffs) with optional `avg_response_time_ms` (`internal/db/analytics.go:257-291`).
  - Model: `models.TrendData` with `Daily []DailyAnalytics` (`internal/models/analytics.go:21`).
  - Service: `AnalyticsService.GetChatbotTrends` (`internal/services/analytics_service.go:28`).
  - Integration test `internal/integration/analytics_full_coverage_test.go:150` validates message count, conversation count, tokens, thumbs up, and handoffs for the current day.
- Source usage analytics:
  - Query: `db.GetSourceUsageStats` aggregates `TimesUsed`, `AvgRelevance`, `PositiveFeedback`, `NegativeFeedback`, `LastUsed` per source ID for a chatbot and a `days` window (`internal/db/source_analytics.go:7-52`).
  - Handler: `AnalyticsHandlers.GetSourceUsage` parses `botID`, checks access (personal/workspace/org), accepts a `days` parameter (default 30, max 365), and returns JSON-encoded stats (`internal/api/handlers/analytics.go:222-296`).
  - Note: Only sources that have been used in at least one message appear; “never used” sources are not currently reported.
- Unanswered queries:
  - Low-confidence responses are tracked via `db.TrackUnansweredQuery`, which inserts or increments `unanswered_queries` per `(chatbot_id, query)` (`internal/db/analytics.go:198-216`).
  - Chat service marks `isUnanswered` for low-tier RAG results and records queries asynchronously (`internal/services/chat_service.go:271-326`).

### Multi-Tenancy and Access Control for Analytics
- Chatbots may belong to:
  - A personal user (legacy / non-org),
  - An organization,
  - A workspace within an organization (`db/migrations/000018_multi_tenant.up.sql`).
- `GetSourceUsage` enforces:
  - Workspace: load workspace by `bot.WorkspaceID`, then require membership in `ws.OrganizationID` (`internal/api/handlers/analytics.go:240-248`).
  - Organization: require membership in `bot.OrganizationID` (`internal/api/handlers/analytics.go:249-255`).
  - Personal: require `bot.UserID == userID` (`internal/api/handlers/analytics.go:256-262`).
  - Failing any of these returns 403 (`internal/api/handlers/analytics.go:266-272`).
- Other analytics endpoints follow the same or similar pattern (user-scoped or org/workspace-scoped via middleware and services).

### Scraper, Timeouts, and Cache
- URL processing:
  - `URLProcessor.Process` orchestrates scraping, discovery, chunking, embedding, and ingestion accounting (`internal/processing/url_processor.go:39-208`).
  - Static scraping via Colly: `ScrapeURL` in `internal/scraper/worker.go:45-83`.
  - Dynamic fallback via headless browser: `ScrapeURLWithFallback` and `ScrapeDynamicURL` in `internal/scraper/worker.go:209-260` and `internal/scraper/browser.go:137-175`.
- Error handling:
  - HTTP errors (including 404) propagate through Colly `OnError` and are logged as `scraper_error` and `url_processing_scrape_failed`.
  - Timeouts:
    - Static: driven by `CollectorConfig.Timeout` in `internal/scraper/colly.go:15-21`.
    - Dynamic: `ScrapeDynamicURL` uses `context.WithTimeout` with `NavTimeout` (`internal/scraper/browser.go:137-155`); timeout errors propagate back to the processor.
  - In both cases, the source is marked failed with an error message; the system does not crash.
- Robots.txt:
  - There is currently no robots.txt parsing or enforcement in the scraper or URL processor.
  - The testing checklist has been updated to mark robots.txt compliance as `(PLANNED)` rather than existing behavior.
- Cache and Redis:
  - A `Cache` interface with `MemoryCache` and `RedisCache` is defined in `internal/scraper/cache.go:13-78`.
  - `NewCache` chooses Redis when `REDIS_URL` is set, otherwise falls back to in-memory (`internal/scraper/cache.go:80-85`).
  - Scrapers use this cache to avoid re-scraping and to store results with TTL; failures fall back gracefully and do not crash the application.

### Planned vs Implemented Highlights
- Implemented:
  - Plan-based limits for scraping, files, chat tokens, ingestions, embeddings, refresh, guardrails, secure embed, branding.
  - Enforcement in backend for:
    - Chat token quotas (authenticated and public),
    - Ingestion quotas (single and bulk),
    - Secure embed enablement and public enforcement,
    - Guardrail features and handoff enablement.
  - Analytics:
    - Overview metrics (messages, conversations, tokens, feedback, handoffs, feedback rate),
    - Trends per day (messages, conversations, tokens, feedback, handoffs, avg response time),
    - Source usage per source (usage count, relevance, feedback, last used),
    - Unanswered query tracking.
  - Scraper timeout and error handling, ingestion metadata extraction, and Redis-backed caching with safe fallback.
- Planned (not yet implemented but listed in checklist):
  - Robots.txt-aware scraping.
  - CAPTCHA token propagation and backend validation for secure embed.
  - Source usage for “never used” sources.
  - Advanced analytics such as action execution tracking, failed message tracking, peak usage charts, and export to CSV/JSON.

