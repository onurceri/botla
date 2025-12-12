## Security-Critical Regression Suite

This document summarizes a minimal automated regression suite for security‑critical behavior in the Botla backend and frontend. It is derived directly from the comprehensive checklist in `.qoder/quests/project-analysis-checklist.md`.

Focus areas:

- Plan enforcement (models and premium features)
- Secure embed domain/token enforcement
- Quotas (chat tokens, ingestions, embeddings, refresh)
- Analytics integrity and multi-tenant access control

---

## 1. Plan Enforcement

### 1.1 Model and Token Limits

Code:

- Handler: `internal/api/handlers/chat.go:73-104`
  - Uses `db.GetPlanByUserID` to load `PlanConfig`.
  - Enforces `AllowedModels` by coercing `cbot.Model` if needed.
  - Checks `plan.Config.Chat.MaxMonthlyTokens` via `db.GetMonthlyTokenUsage`.

Tests (recommended):

- **PLAN-001: Free cannot select Pro/Ultra models**
  - Setup: user on `free` plan with `AllowedModels = ["openai/gpt-4o-mini"]`.
  - Action: `PUT /api/v1/chatbots/:id` with `{"model":"openai/gpt-4o"}`.
  - Assert: DB-stored chatbot `model` is an allowed model (first entry from `AllowedModels`).
- **PLAN-002: Pro cannot select Ultra-only model**
  - Setup: user on `pro` with `AllowedModels = ["openai/gpt-4o-mini","openai/gpt-4o"]`.
  - Action: `PUT /api/v1/chatbots/:id` with `{"model":"anthropic/claude-3.5-sonnet"}`.
  - Assert: stored `model` is coerced back to allowed model.

### 1.2 Feature Gating by Plan

Code:

- `updateChatbot` applies plan checks before mutating the chatbot:
  - Branding:
    - Hide branding: `internal/api/handlers/chatbot_item.go:87-100`
    - Custom branding: `internal/api/handlers/chatbot_item.go:102-114`
  - Refresh policy: `internal/api/handlers/chatbot_item.go:116-124`
  - Discovery: `internal/api/handlers/chatbot_item.go:126-137`
  - Secure embed: `internal/api/handlers/chatbot_item.go:139-151`
  - Guardrails / fallback: `internal/api/handlers/chatbot_item.go:153-165`
  - Handoff: `internal/api/handlers/chatbot_item.go:167-179`

Tests (table-driven integration test on `PUT /api/v1/chatbots/:id`):

- **PLAN-010: Branding permissions**
  - Free + `{"hide_branding": true}` → 403 with `feature: "hide_branding"`.
  - Pro/Ultra + `{"hide_branding": true}` → 200.
  - Free/Pro + `custom_branding` → 403 with `feature: "custom_branding"`.
  - Ultra + `custom_branding` → 200.
- **PLAN-011: Refresh policy**
  - Plan with `Refresh.Enabled=false` + `{"refresh_policy":"auto"}` → 403 (`ErrPlanRefreshUnavailable`).
  - Plan with `Refresh.Enabled=true` + `{"refresh_policy":"auto"}` + `refresh_frequency` → 200 and `next_refresh_at` set.
- **PLAN-012: Discovery mode**
  - Plan with `Scraping.MaxPagesPerCrawl <= 0` + `{"discovery_mode":"auto"}` → 403.
  - Plan with `Scraping.MaxPagesPerCrawl > 0` + `{"discovery_mode":"auto"}` → 200.
- **PLAN-013: Guardrails and handoff**
  - Plan with `Guardrails.CanUseEscalateFallback=false` + `{"threshold_config":{"fallback_mode":"escalate"}}` → 403.
  - Plan with `Guardrails.CanUseEscalateFallback=true` + `{"handoff_enabled":true}` → 200.

### 1.3 Max Chatbots per Plan

Plan config:

- `PlanConfig.MaxChatbots` from `db/migrations/000025_update_plan_guardrails.up.sql:1`.

Test (if not implemented yet, this test will drive implementation):

- **PLAN-020: Max chatbots limit**
  - For `free` (`max_chatbots = 1`):
    - First `POST /api/v1/chatbots` → 200, chatbot created.
    - Second `POST /api/v1/chatbots` → error (define contract, e.g. 403 with upgrade hint).

---

## 2. Secure Embed

### 2.1 Plan-Based Gating

Config:

- `plans.config.security.secure_embed_enabled`:
  - Set per plan in `db/migrations/000023_secure_embed_config.up.sql:1`.
  - Modeled by `PlanConfig.Security` in `internal/models/plan.go:22`.

Code:

- `updateChatbot`:
  - If `secure_embed_enabled` is requested:
    - Loads plan via `db.GetPlanByUserID`.
    - Requires `plan.Config.Security.SecureEmbedEnabled == true`, else 403 (`feature: "secure_embed"`).

Test:

- **SEC-001: Free cannot enable secure embed**
  - Setup: user with `plan_code = "free"`.
  - Action: `PUT /api/v1/chatbots/:id` with `{"secure_embed_enabled": true}`.
  - Assert: 403 and JSON body includes `feature: "secure_embed"`.

### 2.2 Public Enforcement (Origin + Token)

Code:

- Public chat handler for widget:
  - Origin/domain check and token validation in `internal/api/handlers/public.go:193-246`.
  - Only enforced when `cbot.SecureEmbedEnabled` is true.

Existing tests:

- `internal/integration/chatbot_secure_embed_test.go`:
  - Verifies updating chatbot with secure embed fields and reading them back.
- `internal/integration/public_secure_embed_test.go`:
  - Without `X-Embed-Token` → 401.
  - With invalid token → 401.
  - With valid token + allowed Origin → 200.
  - With valid token + disallowed Origin → 403.

Regression IDs:

- **SEC-010: Secure embed update/persistence** (reuse `TestChatbot_SecureEmbed_UpdateAndGet`).
- **SEC-011: Public secure embed enforcement** (reuse `TestPublic_SecureEmbed_Enforcement`).

### 2.3 Widget Behavior

Code:

- `widget/src/widgetApp.tsx:91-126`:
  - Resolves `embedToken` via `embedTokenUrl`.
  - Optionally obtains CAPTCHA token via `window.getCaptchaToken`.
  - Sends POST to `/api/v1/public/chatbots/:id/chat` with `X-Embed-Token`.

Tests:

- **SEC-020: Widget secure embed happy path**
  - Covered by `frontend/e2e/widget-embed-secure.spec.ts`: verifies widget loads, auto-opens, and can send messages with embed token logic.

---

## 3. Quotas (Tokens, Ingestions, Embeddings, Refresh)

### 3.1 Monthly Chat Tokens

Code:

- Aggregation:
  - `db.GetMonthlyTokenUsage` sums `analytics.total_tokens_used` for all chatbots of a user, current month only (`internal/db/analytics.go:44`).
- Write path:
  - `db.IncrementAnalytics` is called from chat service (`internal/services/chat_service.go:271-326`).
- Enforcement:
  - Authenticated chat: `internal/api/handlers/chat.go:73-104`.
  - Public chat: `internal/api/handlers/public.go:229-253`.

Existing tests:

- `internal/integration/usage_tracking_test.go`:
  - **QTA-001 to QTA-005:** verifies `IncrementAnalytics` upsert behavior and `GetMonthlyTokenUsage` across chatbots and months.
- `internal/integration/quota_enforcement_test.go`:
  - **QTA-010: TestQuota_ChatTokensExceeded**
    - Lowers `max_monthly_tokens` on free plan.
    - Pre-populates analytics to exceed limit.
    - Validates chat returns localized `ERR_MONTHLY_TOKENS_EXCEEDED` (HTTP 402/PaymentRequired).

### 3.2 Ingestions and Embedding Tokens

Code:

- Storage:
  - `usage_ingestions` table and helpers in `internal/db/usage_ingestions.go:1-53`.
- Write path:
  - After URL processing, `URLProcessor.Process` calls:
    - `db.IncrementSuccessfulIngestion` and `db.AddEmbeddingTokens` (`internal/processing/url_processor.go:173-196`).
- Enforcement:
  - `SourcesHandlers.checkIngestionQuota` uses `plan.Config.MaxMonthlyIngestions` and `db.GetMonthlyIngestionUsage` (`internal/api/handlers/source_utils.go:18-26`).
  - Bulk URL endpoint enforces both per-bot URL limit and remaining monthly ingestions (`internal/api/handlers/source_bulk.go:66-104`).

Existing tests:

- `internal/integration/quota_enforcement_test.go`:
  - **QTA-020: TestQuota_IngestionExceeded**
    - Lowers `max_monthly_ingestions` for free plan.
    - Pre-increments `usage_ingestions`.
    - Attempts new ingestion and expects a payment-required style error.

### 3.3 Refresh Quota

Config:

- `plans.config.refresh` (`enabled`, `max_monthly`) from `db/migrations/000007_add_refresh_tracking.up.sql:5-24`.

Code:

- Refresh count is tracked in `usage_ingestions.refresh_count`.
- `getUserUsage` includes `refresh_count` in `/api/v1/me` (`internal/api/handlers/me.go:144`).

Test to add:

- **QTA-030: Refresh quota exceeded**
  - Setup: plan with `refresh.max_monthly = 1`.
  - Create chatbot and URL source, perform one refresh.
  - Second refresh attempt should return an explicit error (e.g. `ErrMonthlyRefreshExceeded`).

---

## 4. Analytics Integrity

### 4.1 Overview and Trends

Code:

- Overview:
  - `db.GetAnalyticsOverview` aggregates total messages, conversations, tokens, thumbs up/down, handoffs, and computes `FeedbackRate` for last 30 days (`internal/db/analytics.go:111-145`).
  - Model: `models.AnalyticsOverview` (`internal/models/analytics.go:13`).
- Trends:
  - `db.GetAnalyticsTrends` returns daily `AnalyticsPoint` (messages, conversations, tokens, thumbs up/down, handoffs) (`internal/db/analytics.go:257-291`).
  - Wrapped by `AnalyticsService.GetChatbotTrends` (`internal/services/analytics_service.go:28`).

Existing tests:

- `internal/integration/analytics_full_coverage_test.go`:
  - Verifies for today:
    - Messages >= 2
    - Conversations >= 1
    - Tokens > 0
    - ThumbsUp >= 1
    - Handoffs >= 1
  - Also checks chatbot-specific `/api/v1/chatbots/:id/analytics/trends`.

### 4.2 Handoff Analytics

Code:

- `HandoffService` increments analytics in a background goroutine (`internal/services/handoff_service.go:162-176`).

Existing tests:

- `internal/integration/handoff_test.go`:
  - After a handoff request, polls `/api/v1/analytics` until `handoffs > 0`.

### 4.3 Unanswered Queries

Code:

- `db.TrackUnansweredQuery` inserts or increments `unanswered_queries` per `(chatbot_id, query)` (`internal/db/analytics.go:198-216`).
- Chat service sets `isUnanswered` for low RAG tier and records query in a background goroutine (`internal/services/chat_service.go:271-326`).

Test to add:

- **ANL-030: Unanswered queries tracked**
  - Setup: configure RAG to produce `TierLow` result.
  - Send chat with a unique message.
  - Assert one row in `unanswered_queries` for that chatbot and query, with `occurrence_count = 1`.

### 4.4 Source Usage Analytics and Access Control

Code:

- Source usage query:
  - `db.GetSourceUsageStats` computes usage per source (times used, average relevance, positive/negative feedback, last used) (`internal/db/source_analytics.go:7-52`).
- Handler:
  - `AnalyticsHandlers.GetSourceUsage`:
    - Parses `chatbot_id` from path.
    - Resolves user from context.
    - Enforces access based on:
      - Workspace membership (`bot.WorkspaceID`).
      - Organization membership (`bot.OrganizationID`).
      - Personal ownership (`bot.UserID`).
    - Accepts `days` query param (default 30, max 365) and returns stats as JSON (`internal/api/handlers/analytics.go:222-296`).

Existing tests:

- `internal/db/source_analytics_test.go`:
  - Unit test for `GetSourceUsageStats` (single source scenario).

Tests to add:

- **ANL-040: Source usage endpoint access control**
  - Personal chatbot:
    - Owner user: `GET /api/v1/chatbots/:id/analytics/source-usage` → 200.
    - Different user: same call → 403.
  - Org/workspace chatbot:
    - Member of organization/workspace: 200.
    - Non-member: 403.

---

## 5. Running the Suite

### 5.1 Backend

- Run all integration tests:
  - `go test ./internal/integration/...`
- Optionally tag the new security-focused tests with a build tag (e.g. `//go:build security`) and run:
  - `go test -tags=security ./internal/integration/...`

### 5.2 Frontend

- Key tests to keep green:
  - Plan and usage display:
    - `frontend/src/pages/__tests__/PlanPage.hidden-limits.test.tsx`
  - Secure embed configuration UI:
    - `frontend/src/pages/__tests__/ChatbotDetailPage.secure-embed.test.tsx`
  - Secure embed widget behavior:
    - `frontend/e2e/widget-embed-secure.spec.ts`

These tests together form a minimal but high-value regression suite that protects plan enforcement, secure embed, quota logic, and analytics integrity from regressions.

