# botla.app Comprehensive Feature Documentation

> **Generated**: July 2026
> **Scope**: Complete feature inventory across Backend (Go), Frontend (React), and Widget (Preact)

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Backend Features (Go)](#2-backend-features-go)
3. [Frontend Features (React)](#3-frontend-features-react)
4. [Widget Features (Preact)](#4-widget-features-preact)
5. [Database Schema](#5-database-schema)
6. [API Reference](#6-api-reference)
7. [Security & Authentication](#7-security--authentication)
8. [RAG Pipeline](#8-rag-pipeline)
9. [Plans & Entitlements](#9-plans--entitlements)

---

## 1. Architecture Overview

### 1.1 Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Backend** | Go 1.25+ | REST API, Business Logic, RAG Pipeline |
| **Database** | PostgreSQL 16 | Primary data storage |
| **Cache** | Redis | Rate limiting, session caching, plan cache |
| **Vector DB** | Qdrant | Semantic search embeddings |
| **Storage** | Cloudflare R2 / AWS S3 | File storage |
| **Frontend Dashboard** | React 19 + Vite 7 | Admin dashboard |
| **Widget** | Preact 10 + Vite 7 | Embeddable chat widget |
| **Infrastructure** | Docker Compose | Local development |
| **CI/CD** | Cloudflare Pages | Widget deployment |

### 1.2 Project Structure

```
botla-app/                          # Module: github.com/onurceri/botla-app
├── cmd/
│   ├── server/                     # Backend entrypoint (dependency wiring)
│   └── cli/                        # Admin CLI (promote/demote platform admins)
├── internal/                       # Private application code
│   ├── api/
│   │   ├── handlers/               # HTTP request handlers (44 source files)
│   │   ├── router/                 # Route registration (6 files: router, auth, chatbot, source, public, org, admin)
│   │   └── guards/                 # Admin-only route guards
│   ├── auth/                       # JWT & bcrypt password utilities
│   ├── db/                         # DB connection pool setup (DSN, pooling)
│   ├── integration/                # Integration tests
│   ├── models/                     # Domain models (20 files)
│   ├── processing/                 # Source ingestion queue + per-type processors
│   ├── rag/                        # RAG pipeline (embeddings, LLM, vectors, tools)
│   ├── repository/                 # Data access: interfaces + Postgres impls + mocks (squirrel)
│   ├── scraper/                    # Web scraping (colly static, go-rod dynamic, sitemap)
│   ├── pdf/                        # PDF processing (fitz CGO, stub fallback)
│   ├── services/                   # Business logic (24 source files)
│   └── workers/                    # Bounded worker pool for async/analytics jobs
├── pkg/
│   ├── config/                     # Env-driven configuration (all env vars)
│   ├── logger/                     # Structured JSON logging
│   ├── middleware/                  # HTTP middleware stack (13 files)
│   ├── storage/                    # S3/R2 file storage
│   ├── ratelimit/                  # Redis-backed rate limiting
│   ├── tokenizer/                  # Sentence-aware token counting
│   ├── urlutil/                    # URL helpers + SSRF protection
│   └── errors/                     # Error wrapping (pkgerrors)
├── db/migrations/                  # SQL migrations (golang-migrate, 2 versions)
├── api/openapi.yaml                # Public API spec
├── frontend/                       # React dashboard workspace
├── widget/                         # Preact embeddable widget workspace
└── packages/ui-shared/             # Shared UI components workspace (@botla/ui-shared)
```

---

## 2. Backend Features (Go)

### 2.1 Authentication & Authorization

#### 2.1.1 User Management

| Feature | Description |
|---------|-------------|
| **Registration** | Email/password with bcrypt hashing, password strength validation |
| **Login** | JWT token generation, refresh token rotation |
| **Password Validation** | 8+ chars, uppercase, lowercase, digit, special character |
| **Email Validation** | RFC 5322 format checking |
| **Platform Admin** | Superuser flag for system administration |

#### 2.1.2 JWT Authentication

```
// Token Types
- Access Token:  1 hour expiry (HttpOnly cookie: botla_token)
- Refresh Token: 7 day expiry (HttpOnly cookie: botla_refresh_token, rotated on use)

// Token Claims
- UserID: string (UUID)
- IsPlatformAdmin: bool
- TokenType: "access" | "refresh"
```

#### 2.1.3 Refresh Token Management

- SHA-256 hashed storage in database
- Token revocation support
- Automatic rotation on refresh
- HttpOnly cookie-based delivery (NOT localStorage/Bearer header)
- Frontend uses `withCredentials: true` with `/api/v1/me` as auth source of truth

### 2.2 Chatbot Management

#### 2.2.1 Chatbot Model (`internal/models/chatbot.go`) — 61 fields

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| UserID | string | Owner user ID |
| WorkspaceID | *string | Workspace membership |
| OrganizationID | *string | Organization ownership |
| Name | string | Chatbot display name |
| Description | *string | Optional description |
| CustomInstruction | string | User-editable persona instructions |
| LanguageCode | string | UI language (tr, en) |
| Model | string | LLM model identifier |
| Temperature | float32 | LLM temperature (0.0-2.0) |
| MaxTokens | int | Response token limit |
| SecureEmbedEnabled | bool | Require embed token for chat |
| AllowedDomains | *string | Comma-separated allowed domains |
| EmbedSecret | *string | Secret for signed embed tokens |
| ConfidenceThreshold | float64 | Default confidence threshold |
| ThresholdConfig | *ThresholdConfig | Tiered threshold configuration |
| FallbackMessages | *FallbackMessages | Custom fallback messages |
| TopicRestrictions | *TopicConfig | Allowed/blocked topics |
| HandoffEnabled | bool | Enable human handoff |
| HandoffType | string | Type of handoff |
| HandoffConfig | *HandoffConfig | Handoff configuration |
| DiscoveryMode | string | auto, pending, disabled |
| IncludePaths | []string | URL patterns to include |
| ExcludePaths | []string | URL patterns to exclude |
| SelectorWhitelist | []string | CSS selectors for content |
| RefreshPolicy | string | manual, auto |
| RefreshFrequency | *string | daily, weekly, monthly |
| SuggestedQuestions | []string | AI-generated questions |
| ManualQuestions | []string | Admin-defined questions |
| SuggestionsEnabled | bool | Show suggestion carousel |
| HideBranding | bool | Hide "Powered by Botla" |
| CustomBranding | *CustomBranding | Logo, text, link |
| ThemeColor, WelcomeMessage, Position, BotMessageColor, UserMessageColor, BotMessageTextColor, UserMessageTextColor, ChatFontFamily, ChatHeaderColor, ChatHeaderTextColor, ChatBackgroundColor, BubbleRadius, InputBackgroundColor, InputTextColor, SendButtonColor, BotIcon, BotDisplayName | various | Appearance customization |

#### 2.2.2 Chatbot CRUD Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List/Create Chatbots | `ChatbotHandlers` | `GET/POST /api/v1/chatbots` |
| Get/Update/Delete Chatbot | `ChatbotItemHandlers` | `GET/PUT/DELETE /api/v1/chatbots/:id` |
| Update Basic Info | `ChatbotHandlers.UpdateBasicInfo` | `PUT /api/v1/chatbots/:id/basic-info` |
| Update Appearance | `ChatbotHandlers.UpdateAppearance` | `PUT /api/v1/chatbots/:id/appearance` |
| Update Model Settings | `ChatbotHandlers.UpdateModelSettings` | `PUT /api/v1/chatbots/:id/model` |
| Update Security | `ChatbotHandlers.UpdateSecuritySettings` | `PUT /api/v1/chatbots/:id/security` |
| Update Guardrails | `ChatbotHandlers.UpdateGuardrails` | `PUT /api/v1/chatbots/:id/guardrails` |
| Update Handoff | `ChatbotHandlers.UpdateHandoff` | `PUT /api/v1/chatbots/:id/handoff` |
| Update Refresh | `ChatbotHandlers.UpdateRefresh` | `PUT /api/v1/chatbots/:id/refresh` |
| Update Scraping Config | `ChatbotHandlers.UpdateScrapingConfig` | `PUT /api/v1/chatbots/:id/scraping` |
| Regenerate Suggestions | `SuggestionHandlers.Regenerate` | `POST /api/v1/chatbots/:id/suggestions/regenerate` |
| Get Suggestion Status | `SuggestionHandlers.Status` | `GET /api/v1/chatbots/:id/suggestions/status` |
| Sitemap Discovery | `SourceHandlers.DiscoverSitemap` | `POST /api/v1/chatbots/:id/sitemap/discover` |
| Chatbot Context | `ChatbotContextHandler` | Context management endpoints |

### 2.3 Source Management (Data Ingestion)

#### 2.3.1 Source Types

| Type | Description |
|------|-------------|
| URL | Web page scraping (static + dynamic JS) |
| PDF | PDF document processing (fitz CGO, OCR for Pro+) |
| Text | Raw text content |
| Sitemap | XML sitemap crawling + discovery |

#### 2.3.2 Source Source Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List Sources | `SourceHandlers.ChatbotSources` | `GET /api/v1/chatbots/:id/sources` |
| Create Source | `SourceHandlers.ChatbotSources` | `POST /api/v1/chatbots/:id/sources` |
| Bulk Create | `SourceHandlers.BulkCreateSources` | `POST /api/v1/chatbots/:id/sources/bulk` |
| Get Source | `SourceHandlers.GetSourceStatusOrDelete` | `GET /api/v1/sources/:id` |
| Delete Source | `SourceHandlers.GetSourceStatusOrDelete` | `DELETE /api/v1/sources/:id` |
| Refresh Source | `SourceHandlers.RefreshSource` | `POST /api/v1/sources/:id/refresh` |
| Get Chunks | `SourceHandlers.GetSourceChunks` | `GET /api/v1/sources/:id/chunks` |
| Get Job Status | `TrainingJobHandlers.GetJobStatus` | `GET /api/v1/sources/:id/job` |
| Retry Job | `TrainingJobHandlers.RetryJob` | `POST /api/v1/sources/:id/job/retry` |
| List Pending URLs | `PendingURLHandlers.List` | `GET /api/v1/chatbots/:id/pending-urls` |
| Approve Pending | `PendingURLHandlers.Approve` | `POST /api/v1/chatbots/:id/pending-urls/approve` |
| Reject Pending | `PendingURLHandlers.Reject` | `POST /api/v1/chatbots/:id/pending-urls/reject` |
| Clear Pending | `PendingURLHandlers.Clear` | `DELETE /api/v1/chatbots/:id/pending-urls/clear` |

### 2.4 Chat Operations

#### 2.4.1 Chat Request (`internal/models/chat.go`)

| Field | Type | Description |
|-------|------|-------------|
| Message | string | User message (max 4000 chars) |
| SessionID | string | Conversation session identifier |

#### 2.4.2 Chat Response (`internal/models/chat.go`)

| Field | Type | Description |
|-------|------|-------------|
| Response | string | Bot response (markdown) |
| MessageID | string | Message UUID |
| ConversationID | string | Conversation UUID |
| Sources | []SourceCitation | Source references |
| TokensUsed | int | Token consumption |
| HandoffRequestID | *string | Human handoff request ID |

#### 2.4.3 Chat Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| Protected Chat | `ChatHandlers.Chat` | `POST /api/v1/chatbots/:id/chat` |
| Public Chat | `PublicHandlers.Chat` | `POST /api/v1/public/chatbots/:id/chat` |
| Get Public Config | `PublicHandlers.Config` | `GET /api/v1/public/chatbots/:id/config` |
| Submit Feedback | `PublicHandlers.SubmitFeedback` | `POST /api/v1/public/chatbots/:id/feedback` |
| Human Handoff | `HandoffHandlers.PublicRequestHandoff` | `POST /api/v1/public/chatbots/:id/handoff` |
| Submit Contact | `HandoffHandlers.PublicSubmitEmail` | `POST /api/v1/public/chatbots/:id/handoff/:request_id/contact` |

### 2.5 Conversations & Messages

#### 2.5.1 Conversation Model (`internal/models/conversation.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| ChatbotID | string | Parent chatbot ID |
| SessionID | string | External session identifier |
| UserID | *string | User ID (for protected chats) |
| StartedAt | time.Time | Conversation start |
| LastMessageAt | *time.Time | Last activity |

#### 2.5.2 Message Model (`internal/models/message.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| ConversationID | string | Parent conversation ID |
| Role | string | user, assistant, system |
| Content | string | Message content (markdown) |
| Tokens | *int | Token count |
| Feedback | *bool | Thumbs up/down |
| SourcesJSON | *string | JSON array of source citations |
| ToolCallsJSON | *string | JSON of tool invocations |
| CreatedAt | time.Time | Message timestamp |

### 2.6 Actions (Smart Tools)

#### 2.6.1 Action Model (`internal/models/action.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| ChatbotID | string | Parent chatbot ID |
| Name | string | Action name (API-safe) |
| Description | string | Human-readable description |
| Type | string | http, function |
| Endpoint | *string | API endpoint URL |
| Method | *string | HTTP method |
| Headers | *string | JSON headers |
| Body | *string | Request body template |
| Parameters | *string | JSON array of parameters |
| IsEnabled | bool | Whether action is active |

#### 2.6.2 Action Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List Actions | `ActionHandlers.List` | `GET /api/v1/chatbots/:id/actions` |
| Get Action | `ActionHandlers.Get` | `GET /api/v1/chatbots/:id/actions/:actionId` |
| Create Action | `ActionHandlers.Create` | `POST /api/v1/chatbots/:id/actions` |
| Update Action | `ActionHandlers.Update` | `PUT /api/v1/chatbots/:id/actions/:actionId` |
| Delete Action | `ActionHandlers.Delete` | `DELETE /api/v1/chatbots/:id/actions/:actionId` |
| Get Logs | `ActionHandlers.GetLogs` | `GET /api/v1/chatbots/:id/actions/logs` |

### 2.7 Organization & Workspace Management

#### 2.7.1 Organization Model (`internal/models/organization.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| Name | string | Organization name |
| Slug | string | URL-safe identifier |
| OwnerID | string | Owner user ID |
| CreatedAt | time.Time | Creation timestamp |

#### 2.7.2 Organization/Workspace Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List/Create Organizations | `OrganizationHandlers` | `GET/POST /api/v1/organizations` |
| Update Organization | `OrganizationHandlers.Update` | `PATCH /api/v1/organizations/:id` |
| Delete Organization | `OrganizationHandlers.Delete` | `DELETE /api/v1/organizations/:id` |
| Get/Add/Remove/Update Members | `OrganizationHandlers` | `GET/POST/DELETE/PATCH /api/v1/organizations/:id/members` |
| List/Create Workspaces | `WorkspaceHandlers` | `GET/POST /api/v1/organizations/:id/workspaces` |
| Update/Delete Workspace | `WorkspaceHandlers` | `PATCH/DELETE /api/v1/organizations/:id/workspaces/:wsID` |

### 2.8 User & Profile Management

#### 2.8.1 User Model (`internal/models/user.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| Email | string | User email (unique) |
| PasswordHash | string | bcrypt hashed password |
| FullName | string | User's full name |
| AvatarURL | *string | Profile picture URL |
| PlanID | string | Subscription plan ID |
| IsPlatformAdmin | bool | Admin flag |
| LanguageCode | string | Preferred language |
| CreatedAt | time.Time | Registration timestamp |
| UpdatedAt | time.Time | Last update |
| DeletedAt | *time.Time | Soft delete |
| OnboardingData | *OnboardingData | Onboarding completion state |

#### 2.8.2 User Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| Get Profile | `MeHandlers.Me` | `GET /api/v1/me` |
| Get Plan | `PlanHandlers.GetPlan` | `GET /api/v1/me/plan` |
| Get Usage | `UsageHandlers.GetUsage` | `GET /api/v1/me/usage` |
| Get Onboarding State | `OnboardingHandlers.Get` | `GET /api/v1/me/onboarding` |
| Update Onboarding | `OnboardingHandlers.Update` | `PUT /api/v1/me/onboarding` |
| Skip Onboarding | `OnboardingHandlers.Skip` | `POST /api/v1/me/onboarding/skip` |
| Complete Onboarding | `OnboardingHandlers.Complete` | `POST /api/v1/me/onboarding/complete` |
| List Privacy Consents | `UserPrivacyHandlers` | `GET /api/v1/me/privacy/consents` |
| Update Privacy Consents | `UserPrivacyHandlers` | `PATCH /api/v1/me/privacy/consents` |
| Request Data Export | `UserPrivacyHandlers` | `POST /api/v1/me/privacy/export` |
| Request Data Correction | `UserPrivacyHandlers` | `POST /api/v1/me/privacy/correction` |
| Request Account Deletion | `UserPrivacyHandlers` | `POST /api/v1/me/privacy/delete` |
| List Privacy Requests | `UserPrivacyHandlers.List` | `GET /api/v1/me/privacy/requests` |
| Download Export | `UserPrivacyHandlers.Download` | `GET /api/v1/me/privacy/requests/:id/download` |
| Protected Endpoint | `ProtectedHandler` | `ANY /api/v1/protected` |

### 2.9 Plans & Entitlements

#### 2.9.1 Plan Model (`internal/models/plan.go` + `plan_limits.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| Name | string | Plan display name |
| Code | string | Plan code (free, starter, pro, ultra) |
| Limits | *PlanLimits | Typed 72+ field limits config (replacing JSONB) |
| IsActive | bool | Plan active status |
| SortOrder | int | Display ordering |

`PlanLimits` covers: max chatbots, scraping config, files config, chat config, RAG config, guardrails config, branding config, rate limits, security config, refresh config, and more. Validated at startup with defaults for Free/Starter/Pro/Ultra.

#### 2.9.2 Plan Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List Plans | `PlansHandlers.GetAllPlans` | `GET /api/v1/plans` |
| Get Plan By Code | `PlansHandlers.GetPlanByCode` | `GET /api/v1/plans/:code` |
| Get Current Plan | `PlanHandlers.GetPlan` | `GET /api/v1/me/plan` |
| Admin CRUD Plans | `AdminPlanHandlers` | `GET/PATCH /api/v1/admin/plans` |
| Update Plan Limits | `AdminPlanHandlers.UpdateLimits` | `PATCH /api/v1/admin/plans/:id/limits` |
| Invalidate Cache | `AdminPlanHandlers.InvalidateCache` | `POST /api/v1/admin/plans/:id/cache-invalidate` |

### 2.10 Analytics

#### 2.10.1 Analytics Model (`internal/models/analytics.go`)

| Field | Type | Description |
|-------|------|-------------|
| ChatbotID | string | Parent chatbot ID |
| Date | time.Time | Aggregation date |
| MessageCount | int | Total messages |
| PositiveFeedback | int | Thumbs up count |
| NegativeFeedback | int | Thumbs down count |
| TokenUsage | int | Total tokens consumed |
| EmbeddingTokens | int | Embedding tokens |
| ChatTokens | int | Chat completion tokens |
| ConversationCount | int | Unique conversations |
| SourceCount | int | Total sources |

#### 2.10.2 Analytics Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| Global Analytics | `AnalyticsHandlers.GetAnalytics` | `GET /api/v1/analytics` |
| Chatbot Overview | `AnalyticsHandlers.GetChatbotAnalyticsOverview` | `GET /api/v1/chatbots/:id/analytics/overview` |
| Chatbot Trends | `AnalyticsHandlers.GetChatbotAnalyticsTrends` | `GET /api/v1/chatbots/:id/analytics/trends` |
| Source Usage | `AnalyticsHandlers.GetSourceUsage` | `GET /api/v1/chatbots/:id/analytics/sources` |

### 2.11 Human Handoff

#### 2.11.1 Handoff Model (`internal/models/handoff.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| ChatbotID | string | Parent chatbot ID |
| ConversationID | string | Parent conversation ID |
| MessageID | string | Original message ID |
| Status | string | pending, contacted, resolved |
| ContactEmail | *string | User-provided email |
| ContactedAt | *time.Time | Agent contact time |
| ResolvedAt | *time.Time | Resolution time |
| Notes | *string | Agent notes |

#### 2.11.2 Handoff Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| Public Request | `HandoffHandlers.PublicRequestHandoff` | `POST /api/v1/public/chatbots/:id/handoff` |
| Submit Contact | `HandoffHandlers.PublicSubmitEmail` | `POST /api/v1/public/chatbots/:id/handoff/:requestId/contact` |
| List Requests | `HandoffHandlers.ListHandoffRequests` | `GET /api/v1/chatbots/:id/handoff-requests` |
| Get Detail | `HandoffHandlers.GetHandoffRequestDetail` | `GET /api/v1/chatbots/:id/handoff-requests/:requestId` |
| Update Status | `HandoffHandlers.UpdateHandoffRequest` | `PATCH /api/v1/chatbots/:id/handoff-requests/:requestId` |

### 2.12 Guardrails

#### 2.12.1 Guardrails Model (`internal/models/guardrail.go`)

| Field | Type | Description |
|-------|------|-------------|
| ChatbotID | string | Parent chatbot ID |
| TopicRestrictions | *TopicConfig | Allowed/blocked topics |
| FallbackMessages | *FallbackMessages | Custom fallback texts |
| ThresholdConfig | *ThresholdConfig | Confidence thresholds |

### 2.13 Training Jobs

#### 2.13.1 Training Job Model (`internal/models/training_job.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| SourceID | string | Associated source ID |
| ChatbotID | string | Parent chatbot ID |
| Status | string | pending, running, completed, failed, cancelled |
| CurrentStep | string | fetch_source, parse_content, chunk_text, embed_chunks, store_vectors |
| ProgressPercent | int | 0-100 |
| ErrorCode | *string | Error code if failed |
| ErrorMessage | *string | Error details |
| StartedAt | *time.Time | Job start time |
| CompletedAt | *time.Time | Job completion time |

### 2.14 Background Processing

#### 2.14.1 Source Queue (`internal/processing/`)

- `sources_queue.go` — Main ingestion queue (bounded worker pool)
- `queue_manager.go` — Queue manager orchestration
- `job_processor.go` — Job processing pipeline
- `url_processor.go` — URL source processor (scrape → chunk → embed → upsert)
- `pdf_processor.go` — PDF source processor (fitz CGO + OCR)
- `text_processor.go` — Raw text source processor
- `suggestions.go` — AI suggestion generation for sources
- `interfaces.go` — Processor interfaces

#### 2.14.2 Refresh Scheduler (`internal/services/refresh_scheduler.go`)

- Scheduled content refresh for sources
- Daily, weekly, monthly refresh policies
- Automatic cooldown enforcement

#### 2.14.3 Retention Job (`internal/services/retention_job.go`)

- KVKK/GDPR compliance data cleanup
- Automatic deletion of old data
- Configurable retention periods

#### 2.14.4 Worker Pool (`internal/workers/pool.go`)

- Generic bounded worker pool for async/analytics jobs
- Configurable worker count (default: 4 for processing, 10 for analytics)

### 2.15 Web Scraping

#### 2.15.1 Scraper (`internal/scraper/` — 12 source files)

| Feature | Description |
|---------|-------------|
| **Static Scraping** | Colly-based HTML content extraction |
| **Dynamic Scraping** | go-rod browser automation for JS-rendered content |
| **Sitemap Parsing** | XML sitemap crawling |
| **CSS Selector Targeting** | Specific element extraction |
| **Path Filtering** | Include/exclude URL patterns |
| **Discovery Mode** | Auto-discover subpages |
| **Browser Pool** | Configurable pool of headless Chromium instances |
| **Cache Layer** | Scrape cache with ETag support |
| **Encoding Detection** | Automatic charset detection/handling |

#### 2.15.2 SSRF Protection (`pkg/urlutil/ssrf.go`)

- Blocks private IP ranges (10.x, 172.16.x, 192.168.x)
- Blocks localhost and metadata endpoints
- Supports allowlist for internal services
- Configurable strict/relaxed mode

### 2.16 PDF Processing

| Feature | Description |
|---------|-------------|
| **Text Extraction** | go-fitz-based PDF text extraction (CGO + fitz build tag) |
| **Stub Fallback** | No-CGO fallback when fitz unavailable |
| **OCR Support** | Tesseract for image-based PDFs (Pro+) |
| **Multi-page** | Handles multi-page documents |

### 2.17 Admin Operations

#### 2.17.1 Admin Endpoints

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| Overview Stats | `AdminHandlers.GetOverviewStats` | `GET /api/v1/admin/stats/overview` |
| List Users | `AdminHandlers.ListUsers` | `GET /api/v1/admin/users` |
| Get/Update User | `AdminHandlers` | `GET/PATCH /api/v1/admin/users/:id` |
| List Organizations | `AdminHandlers.ListOrganizations` | `GET /api/v1/admin/organizations` |
| Get Organization | `AdminHandlers.GetOrganization` | `GET /api/v1/admin/organizations/:id` |
| List Chatbots | `AdminChatbotHandlers.List` | `GET /api/v1/admin/chatbots` |
| Get Chatbot | `AdminChatbotHandlers.Get` | `GET /api/v1/admin/chatbots/:id` |
| Force Refresh | `AdminChatbotHandlers.ForceRefresh` | `POST /api/v1/admin/chatbots/:id/force-refresh` |
| List Sources | `AdminSourceHandlers.List` | `GET /api/v1/admin/sources` |
| Source Stats | `AdminSourceHandlers.GetSourceStats` | `GET /api/v1/admin/sources/stats` |
| Source Types | `AdminSourceHandlers.GetSourceTypes` | `GET /api/v1/admin/sources/types` |
| Reprocess Source | `AdminSourceHandlers.Reprocess` | `POST /api/v1/admin/sources/:id/reprocess` |
| Detailed Health | `AdminHealthHandlers.GetDetailedHealth` | `GET /api/v1/admin/health/detailed` |
| List Queues | `AdminQueueHandlers.GetQueues` | `GET /api/v1/admin/queues` |
| Stuck Jobs | `AdminQueueHandlers.GetStuckJobs` | `GET /api/v1/admin/queues/stuck` |
| Retry Job | `AdminQueueHandlers.RetryJob` | `POST /api/v1/admin/queues/:id/retry` |
| Delete Job | `AdminQueueHandlers.DeleteJob` | `DELETE /api/v1/admin/queues/:id` |
| List Errors | `AdminErrorHandlers.List` | `GET /api/v1/admin/errors` |
| Error Stats | `AdminErrorHandlers.GetErrorStats` | `GET /api/v1/admin/errors/stats` |
| Get Error | `AdminErrorHandlers.GetError` | `GET /api/v1/admin/errors/:id` |
| List Audit Logs | `AdminAuditHandlers` | `GET /api/v1/admin/audit-logs` |
| List Privacy Requests | `AdminPrivacyHandlers` | `GET /api/v1/admin/privacy/requests` |
| Process Request | `AdminPrivacyHandlers.Process` | `PATCH /api/v1/admin/privacy/requests/:id` |
| Generate Export | `AdminPrivacyHandlers.GenerateExport` | `POST /api/v1/admin/privacy/export/:userId` |
| List/Get Plans | `AdminPlanHandlers` | `GET /api/v1/admin/plans` |

### 2.18 Services Layer (`internal/services/` — 24 source files)

| Service | Description |
|---------|-------------|
| `action_service.go` | Action CRUD + execution logic |
| `admin_service.go` | Admin dashboard operations |
| `analytics_service.go` | Analytics aggregation |
| `chat_context.go` / `chat_context_builder.go` | Chat context assembly |
| `chat_fallback.go` | Fallback handling (static/smart/escalate) |
| `chat_helpers.go` / `chat_pipeline.go` | Chat pipeline orchestration |
| `chat_prompts.go` | Prompt templates (always English) |
| `chat_service.go` | Main chat service |
| `chatbot_service.go` | Chatbot business logic |
| `errors.go` | Service error definitions |
| `guardrail_service.go` | Guardrails/confidence thresholds |
| `handoff_service.go` | Human handoff workflow |
| `model_registry.go` | LLM model resolution |
| `organization_service.go` | Organization management |
| `plan_service.go` | Plan operations + Redis cache |
| `privacy_service.go` | GDPR/KVKK privacy logic |
| `quota_enforcer.go` | Quota enforcement |
| `rag_service.go` | RAG admin service (vector deletion) |
| `refresh_scheduler.go` | Auto-refresh scheduling |
| `retention_job.go` | Data retention cleanup |
| `workspace_service.go` | Workspace management |

### 2.19 Repository Layer (`internal/repository/` — 19 source files)

| Repository | Interface |
|------------|-----------|
| `action_repo.go` | ActionRepository |
| `admin_chatbot_repo.go` | AdminChatbotRepository |
| `admin_repo.go` | AdminRepository |
| `analytics_repo.go` | AnalyticsRepository |
| `chatbot_repo.go` | ChatbotRepository |
| `conversation_repo.go` | ConversationRepository |
| `handoff_repo.go` | HandoffRepository |
| `organization_repo.go` | OrganizationRepository |
| `pending_url_repo.go` | PendingURLRepository |
| `plan_repo.go` | PlanRepository |
| `privacy_repo.go` | PrivacyRepository |
| `queue_repo.go` | QueueRepository |
| `source_repo.go` | SourceRepository |
| `suggestion_job_repo.go` | SuggestionJobRepository |
| `training_job_repo.go` | TrainingJobRepository |
| `usage_repo.go` | UsageRepository |
| `user_repo.go` | UserRepository |
| `pagination.go` | Pagination helpers |
| `interfaces.go` | All repository interfaces + types |

### 2.20 RAG Pipeline (`internal/rag/` — ~20 source files)

| File | Purpose |
|------|---------|
| `llm_client.go` | LLM client interface |
| `openai.go` | OpenAI LLM + embedding implementation |
| `openrouter.go` | OpenRouter LLM implementation |
| `client_factory.go` | LLM provider resolution (OpenAI vs OpenRouter) |
| `circuit_breaker.go` | Gobreaker-based circuit breaker for LLM calls |
| `qdrant.go` | Qdrant HTTP vector client |
| `embedding.go` | Embedding client (OpenAI text-embedding-3-small) |
| `chunker.go` | Text chunking logic |
| `search.go` | Tiered similarity search with thresholds |
| `tokens.go` | Token counting utilities |
| `tools.go` | Tool definitions for function calling |
| `tool_name_generator.go` | Unique tool name generation |
| `tool_executor.go` | Tool execution logic |
| `llm_metrics.go` | LLM metrics tracking |
| `topic_extractor.go` | Topic extraction from queries |
| `subsystem.go` | RAG subsystem wiring |

### 2.21 Middleware Stack (`pkg/middleware/` — 13 source files)

Middleware chain order (outermost first):
```
RequestID → SecurityHeaders → Recovery → RequestLogger → MaxBytes(52MB) → PlanLoader → DeletedAccount → RateLimit → mux
```

| Middleware | Purpose |
|-----------|---------|
| `auth.go` | `AuthMiddleware` (JWT cookie verification), `OptionalAuthMiddleware` (public endpoints) |
| `cors.go` | CORS policy enforcement |
| `deleted_account.go` | Blocks deleted users |
| `maxbytes.go` | Request body size limiter (52MB) |
| `organization.go` | `RequireOrganizationAccess`, `ExtractTenantContext` |
| `plan_context.go` | Plan context injection |
| `plan_loader.go` | Plan loader middleware |
| `ratelimit.go` | Redis-backed rate limiter |
| `recovery.go` | Panic recovery |
| `request_id.go` | RequestID injection |
| `requestlog.go` | Structured request logging |
| `security.go` | SecurityHeaders (CSP, HSTS, etc.) |

### 2.22 RAG Pipeline (detail)

#### 2.22.1 Pipeline Stages

```
[Source Input]
      ↓
[Content Fetch] → [HTML Parsing / PDF Text Extraction / Sitemap Parse]
      ↓
[Chunking] → [Sentence-aware splitting with 15% overlap]
      ↓
[Embedding] → [OpenAI text-embedding-3-small, 25-chunk batches]
      ↓
[Vector Storage] → [Qdrant collection, deterministic point IDs: MakePointID(sourceID, index)]
      ↓
[Query] → [Embed query vector]
      ↓
[Search] → [Tiered similarity search (High/Medium/Low)]
      ↓
[LLM Generation] → [OpenAI GPT-4o / GPT-4o-mini via circuit breaker]
      ↓
[Response] → [With citations and sources]
```

#### 2.22.2 Chunking Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Target Tokens** | 512 | Desired chunk size |
| **Overlap** | ~15% | Tail overlap for context |
| **Boundaries** | Paragraph/Sentence | Preserves semantic units |
| **Abbreviation Handling** | Configurable | Prevents sentence splitting |

#### 2.22.3 Search Tiers

| Tier | Threshold | Description |
|------|-----------|-------------|
| **High** | ≥ 0.50 | Strong match |
| **Medium** | 0.30 - 0.49 | Weak match (with warning) |
| **Low** | < 0.30 | No good match (fallback) |

#### 2.22.4 Fallback Modes

| Mode | Trigger | Behavior |
|------|---------|----------|
| **Static** | Low confidence | Generic "I don't know" |
| **Smart** | Low confidence | AI redirect based on capabilities |
| **Escalate** | Low confidence | Human handoff workflow |

#### 2.22.5 Topic Extraction

During ingestion, each source undergoes:
- **Capability Summary** — What the source covers
- **Suggested Questions** — Auto-generated Q&A
- **Language Detection** — Content language

#### 2.22.6 Circuit Breaker

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Failure Ratio** | 0.50 | Trip on 50% failure |
| **Timeout** | 30s | Half-open state duration |
| **Requests** | 3 | Requests in half-open |
| **Interval** | 60s | Sample window |

---

## 3. Frontend Features (React)

### 3.1 Stack

| Category | Technology |
|----------|------------|
| **Framework** | React 19.2, TypeScript 5.9 |
| **Build** | Vite 7.2 |
| **Routing** | React Router DOM v7.9 |
| **Server State** | TanStack React Query v5 |
| **HTTP** | Axios (withCredentials: true) |
| **UI** | shadcn/ui (New York), Radix UI primitives |
| **Styling** | Tailwind CSS v4, CVA, clsx, tailwind-merge |
| **Icons** | Lucide React |
| **Charts** | Recharts |
| **Animations** | Framer Motion |
| **Forms** | React Hook Form + Zod |
| **Testing** | Vitest + Testing Library, Playwright E2E |
| **Analytics** | PostHog (conditional) |
| **Markdown** | markdown-to-jsx |
| **Color Picker** | react-colorful |
| **Toast** | Sonner |
| **Shared** | @botla/ui-shared (workspace) |

### 3.2 Routing Structure

#### 3.2.1 Public Routes

| Path | Component | Description |
|------|-----------|-------------|
| `/` | `LandingPage` | Marketing landing page |
| `/login` | `LoginPage` | User login |
| `/register` | `RegisterPage` | User registration |

#### 3.2.2 Protected Routes (Dashboard, wrapped in `PrivateRoute` + `OrganizationProvider` + `DashboardLayout`)

| Path | Component | Description |
|------|-----------|-------------|
| `/onboarding` | `OnboardingPage` | First-time setup (no DashboardLayout) |
| `/dashboard` | `DashboardPage` | Index redirect |
| `/dashboard/chatbots` | `ChatbotsPage` | List all chatbots |
| `/dashboard/chatbots/:id` | `ChatbotDetailPage` | Index redirects to settings |
| `/dashboard/chatbots/:id/settings` | `SettingsTab` | Chatbot settings |
| `/dashboard/chatbots/:id/security` | `SecurityTab` | Security configuration |
| `/dashboard/chatbots/:id/sources` | `SourcesTab` | Source management |
| `/dashboard/chatbots/:id/actions` | `ActionsTab` | Smart actions |
| `/dashboard/chatbots/:id/playground` | `PlaygroundTab` | Chat playground |
| `/dashboard/chatbots/:id/deploy` | `DeployTab` | Deployment/embed code |
| `/dashboard/chatbots/:id/insights` | `InsightsTab` | Analytics & insights |
| `/dashboard/settings` | (redirect) | Redirects to profile |
| `/dashboard/settings/profile` | `ProfilePage` | User profile |
| `/dashboard/settings/organization` | `OrganizationSettingsPage` | Organization settings |
| `/dashboard/settings/workspace` | `WorkspaceSettingsPage` | Workspace settings |
| `/dashboard/settings/plan` | `PlanPage` | Subscription management |
| `/dashboard/settings/privacy` | `PrivacySettingsPage` | Privacy controls (KVKK/GDPR) |

#### 3.2.3 Admin Routes (wrapped in `PrivateRoute` + `AdminRoute` + `AdminLayout`)

| Path | Component | Description |
|------|-----------|-------------|
| `/admin` | `AdminDashboardPage` | Admin dashboard |
| `/admin/users` | `AdminUsersPage` | User management |
| `/admin/organizations` | `AdminOrganizationsPage` | Organization management |
| `/admin/chatbots` | `AdminChatbotsPage` | Chatbot management |
| `/admin/sources` | `AdminSourcesPage` | Source monitoring |
| `/admin/system` | `AdminSystemPage` | System health |
| `/admin/queues` | `AdminQueuesPage` | Job queue status |
| `/admin/errors` | `AdminErrorsPage` | Error log viewer |
| `/admin/audit` | `AdminAuditPage` | Audit logs |
| `/admin/privacy` | `AdminPrivacyPage` | Privacy compliance |
| `/admin/plans` | `AdminPlansPage` | Plan management |

### 3.3 Design System

- **Apple-inspired design** with amber/gold primary (`#f59e0b`)
- Neutral backgrounds (`#fafafa`, `#1d1d1f`)
- SF Pro Display font stack
- Glassmorphism effects (glass, glass-card, glass-panel)
- Custom easing curves (ease-out-expo, ease-spring)
- Premium card styles (card-elevated, card-interactive)
- Professional sidebar styles

### 3.4 Internationalization

| Language | Status |
|----------|--------|
| **Turkish (tr)** | Full support — hardcoded translation objects in `src/i18n/` (landing, errors, privacy) |
| **English (en)** | Not implemented — no i18n framework, no English translations |

The app has Turkish-only UI strings with no i18n framework. No react-i18next or similar library.

### 3.5 State Management

| Approach | Library | Usage |
|----------|---------|-------|
| **Server State** | TanStack Query v5 | Primary state management (staleTime: 0, refetchOnWindowFocus: true, retry: 1) |
| **Auth State** | React Context (AuthContext) | Wraps entire app, uses React Query for `/api/v1/me` |
| **Org/Workspace State** | React Context (OrganizationContext) | Wraps protected routes, localStorage persistence |
| **Local State** | useState, useReducer | Component state |
| **URL State** | React Router | Route parameters |

No Redux, Zustand, MobX, or Jotai.

### 3.6 API Layer (`frontend/src/api/`)

| Module | Features |
|--------|----------|
| `client.ts` | Axios instance (withCredentials: true), interceptors, token refresh queue, session expiry handling |
| `auth.ts` | Login, register, logout, me |
| `chatbot.ts` | CRUD operations, appearance, model, security, guardrails, handoff, scraping, refresh, context |
| `source.ts` | Source CRUD, upload, refresh, chunks |
| `chat.ts` | Chat messages, feedback |
| `action.ts` | Action CRUD, logs |
| `organization.ts` | Organization management |
| `workspace.ts` | Workspace management |
| `analytics.ts` | Usage stats, insights, trends |
| `handoff.ts` | Handoff requests |
| `plan.ts` | Plans, limits |
| `admin.ts` | Admin operations |
| `onboarding.ts` | Onboarding flow |

### 3.7 UI Components (`src/components/ui/` — 18 shadcn-style components)

Button, Card, Input, Dialog, Tabs, Toast (sonner), Tooltip, Select, Switch, Badge, Avatar, Spinner, Alert, Label, Progress, Textarea, DropdownMenu, ColorPicker, ErrorBoundary, StatusBadge, PlanBadge, Table

### 3.8 Pages (23 page files)

| Area | Files |
|------|-------|
| **Public** | LandingPage, LoginPage, RegisterPage |
| **Dashboard** | DashboardPage, ChatbotsPage, ChatbotDetailPage, OnboardingPage, ProfilePage, PlanPage, PrivacySettingsPage |
| **Admin** | AdminLayout, AdminDashboardPage, AdminUsersPage, AdminOrganizationsPage, AdminChatbotsPage, AdminSourcesPage, AdminSystemPage, AdminQueuesPage, AdminErrorsPage, AdminAuditPage, AdminPrivacyPage, AdminPlansPage |

### 3.9 Feature Structure (`src/features/`)

| Feature | Components | Hooks | Tests |
|---------|-----------|-------|-------|
| **chatbot/** | NewChatbotForm, IngestionProgress, SourceCard, IdentitySection, appearance, security, scraper, refresh, guardrails, handoff, etc. | useAutoSave, useChatbotForm, usePreview, useSourceOps, useSourceUpload, useSuggestionPolling, useToastErrors | 25+ test files |
| **analytics/** | ChatbotAnalytics, SourceUsageStats | — | 2 test files |
| **admin/** | HealthPanel, StatsCard, AdminRoute | — | 3 test files |
| **organization/** | OrganizationContext, OrganizationSwitcher, dialogs | — | — |

### 3.10 Tests

| Type | Framework | Count |
|------|-----------|-------|
| Unit | Vitest + Testing Library | ~72 test files |
| E2E | Playwright | 12 spec files |

### 3.11 Onboarding Wizard (`src/components/onboarding/`)

Multi-step wizard with 4 steps: StepBotName, StepPersonality, StepDataSource, StepComplete. Includes StepIndicator, NavigationButtons, and useOnboardingWizard hook.

---

## 4. Widget Features (Preact)

### 4.1 Stack

| Category | Technology |
|----------|------------|
| **Framework** | Preact 10.27 |
| **Build** | Vite 7 (library mode, IIFE output) |
| **Styling** | Tailwind CSS v4 (via PostCSS) |
| **Markdown** | react-markdown + remark-gfm + rehype-sanitize |
| **Testing** | Vitest + Testing Library, Playwright E2E |
| **Deployment** | Cloudflare Pages (wrangler) |
| **Shared** | @botla/ui-shared (workspace) |

### 4.2 Architecture

The widget is deployed as a **single IIFE bundle** (`widget.js`) to **Cloudflare Pages**. It is embedded via `<script>` tag with `data-bot` attributes.

Key architectural decisions:
- **Shadow DOM isolation** — full CSS reset inside shadow root
- **No state management library** — pure React hooks + localStorage
- **No router** — single-page widget
- **Turkish-first UI** — all UI strings in Turkish
- **playground support** — postMessage-based live config updates with origin validation

### 4.3 Entry Points

| File | Description |
|------|-------------|
| `src/widget.tsx` | IIFE entry: `mount()` / `unmount()`, shadow DOM setup, postMessage listener, safe-area detection |
| `src/widgetApp.tsx` | Main `WidgetApp` Preact component managing all state and rendering |

### 4.4 Components

| Component | File | Description |
|-----------|------|-------------|
| `ChatBubble` | `ChatBubble.tsx` | Floating launcher button with unread badge, pulse animation |
| `ChatDrawer` | `ChatDrawer.tsx` | Main chat panel: header, message list, input, branding, suggestions, loading |
| `Message` | `Message.tsx` | Message bubble with markdown (react-markdown), GFM, sanitization, feedback thumbs, handoff |
| `Suggestions` | `Suggestions.tsx` | Suggested questions carousel with nav |
| `ConfirmModal` | `ConfirmModal.tsx` | Reusable modal with icon, title, confirm/cancel, focus trap |

### 4.5 Widget Configuration (data-* attributes)

| Parameter | Type | Description |
|-----------|------|-------------|
| `chatbot-id` | string | Chatbot UUID (required) |
| `api-base` | string | API base URL override |
| `color` | string | Theme color (hex) |
| `welcome` | string | Welcome message override |
| `position` | string | bottom-left or bottom-right |
| `auto-open` | boolean | Auto-open on load |
| `panel-height` / `panel-width` | string | Panel dimensions (CSS) |
| All appearance colors | string | Header, bubble, text, input, send button colors |
| `bot-name` / `bot-icon` | string | Custom bot identity |
| `hide-branding` | boolean | Hide "Powered by Botla" |
| `custom-branding` | JSON | Custom branding object |
| `suggestions` | JSON | Override suggestions |
| `embed-token-url` | string | URL to fetch embed token |
| `captcha-site-key` | string | hCaptcha site key |
| `reset-session` | boolean | Reset session on load |
| `session-id` | string | Override session ID |
| `preview-mode` | boolean | Preview/embed mode |
| `position-strategy` | string | fixed or absolute |
| `use-url-overrides` | boolean | Use URL param overrides |

### 4.6 State & Session Management

All state lives in `WidgetApp` via `useState`/`useRef`. Session persistence via `localStorage` (key prefix: `chatbot_session_`).

| Module | File | Description |
|--------|------|-------------|
| Session utils | `utils/session.ts` | `getSession`, `saveSession`, `clearSession`, `ensureSession` |
| Logger | `utils/logger.ts` | Console logger with debug toggle |
| Sanitize | `utils/sanitize.ts` | `sanitizeUrl`, `sanitizeMessage`, `sanitizeMarkdown` (XSS prevention) |

### 4.7 API Client

| File | Description |
|------|-------------|
| `src/api/chat.ts` | `sendMessage()` — POST to `/api/v1/public/chatbots/:id/chat`, handles embed tokens, captcha, session, rate limiting |

Direct API calls in `widgetApp.tsx`:
- Config fetch: `GET /api/v1/public/chatbots/:id/config`
- Chat: `POST /api/v1/public/chatbots/:id/chat`
- Feedback: `POST /api/v1/public/chatbots/:id/feedback`
- Handoff: `POST /api/v1/public/chatbots/:id/handoff/:requestId/contact`
- Embed token: `GET <embedTokenUrl>` (optional)

### 4.8 Internationalization

| Language | Status |
|----------|--------|
| **Turkish (tr)** | Default — all UI strings |
| **English (en)** | Error messages only (in `src/i18n/errors.ts`) |

### 4.9 Widget Features

| Feature | Description |
|---------|-------------|
| Real-time Chat | POST-based messaging with streaming |
| Markdown Rendering | react-markdown with GFM + sanitization |
| Suggestions | AI-generated question carousel |
| Feedback | Thumbs up/down on messages |
| Human Handoff | Email capture for support |
| Session Persistence | localStorage for conversation history |
| Typing Indicator | Animated typing dots |
| Unread Badge | Message count on launcher |
| Custom Styling | CSS variables for theming |
| Shadow DOM | Style isolation with aggressive CSS reset |
| Google Fonts | Dynamic font loading |
| Captcha Support | hCaptcha integration |
| Embed Tokens | Secure embed authentication |
| Event Emission | postMessage to parent window |
| Preview Mode | Playground integration |
| Dynamic Safe Area | Fixed header detection on host page |
| New Session | Reset conversation with confirm modal |

### 4.10 Tests

| Type | Framework | Count |
|------|-----------|-------|
| Unit | Vitest + Testing Library | 8 test files (components, edge-cases, session, XSS) |
| E2E | Playwright | 6 spec files (chat, mobile, branding, config, secure embed) |

---

## 5. Database Schema

### 5.1 Core Tables

| Table | Description |
|-------|-------------|
| `users` | User accounts |
| `organizations` | Tenant organizations |
| `workspaces` | Workspace isolation |
| `chatbots` | Chatbot configurations |
| `sources` | Data sources (URL, PDF, text, sitemap) |
| `chunks` | Text chunks for RAG |
| `conversations` | Chat sessions |
| `messages` | Individual messages |
| `actions` | Smart tools/actions |
| `action_logs` | Action execution history |
| `handoffs` | Human handoff requests |
| `training_jobs` | Background ingestion jobs |
| `pending_urls` | URLs awaiting approval |
| `refresh_tokens` | JWT refresh tokens |
| `plans` | Subscription plans |
| `plan_configs` | Plan configuration |
| `analytics` | Daily analytics aggregates |
| `privacy_requests` | GDPR/KVKK data requests |
| `privacy_consents` | User privacy consents |
| `suggestion_jobs` | AI suggestion generation jobs |

### 5.2 Migration History

The system has evolved through **2 database migration versions** covering:

| Version | Features |
|---------|----------|
| **000001** | Initial schema: users, chatbots, conversations, messages, sources, chunks, actions, organizations, workspaces, plans, analytics, handoffs, training jobs, pending URLs, tokens |
| **000002** | Privacy requests table, `deleted_at` support |

---

## 6. API Reference

### 6.1 API Versioning

- **Base Path**: `/api/v1`
- **Format**: JSON
- **Authentication**: HttpOnly cookies (botla_token / botla_refresh_token)

### 6.2 Endpoint Categories

| Category | Base Path | Auth |
|----------|-----------|------|
| Auth | `/api/v1/auth/*` | Public |
| Chatbots | `/api/v1/chatbots/*` | Protected |
| Sources | `/api/v1/sources/*` | Protected |
| Chat | `/api/v1/chatbots/:id/chat` | Protected |
| Public | `/api/v1/public/*` | Public/Optional |
| User | `/api/v1/me/*` | Protected |
| Privacy | `/api/v1/me/privacy/*` | Protected |
| Organization | `/api/v1/organizations/*` | Protected |
| Plans | `/api/v1/plans/*` | Public |
| Analytics | `/api/v1/analytics` | Protected |
| Admin | `/api/v1/admin/*` | Platform Admin |
| Health | `/health` | Public |

### 6.3 Rate Limits

Rate limits per plan (configurable via env vars):

| Tier | Global | Chat | Sources | Auth (login) |
|------|--------|------|---------|--------------|
| **Free** | 60/min | 20/min | 10/min | 5/min |
| **Pro** | 120/min | 50/min | 30/min | 5/min |
| **Ultra** | 500/min | 200/min | 100/min | 5/min |

---

## 7. Security & Authentication

### 7.1 Authentication Methods

| Method | Type | Usage |
|--------|------|-------|
| **JWT Access Token** | HttpOnly cookie (botla_token) | API authentication |
| **JWT Refresh Token** | HttpOnly cookie (botla_refresh_token) | Token renewal with rotation |
| **Embed Token** | X-Embed-Token header | Widget authentication |
| **Platform Admin** | JWT claim | Admin routes |

### 7.2 Password Policy

| Requirement | Rule |
|-------------|------|
| **Minimum Length** | 8 characters |
| **Uppercase** | At least one A-Z |
| **Lowercase** | At least one a-z |
| **Digit** | At least one 0-9 |
| **Special Character** | At least one @$!%*?& |
| **Hashing** | bcrypt (not Argon2) |

### 7.3 Data Protection

| Feature | Description |
|---------|-------------|
| **bcrypt Hashing** | Password storage |
| **Token Hashing** | SHA-256 refresh token storage |
| **SSRF Protection** | Blocks private IP access |
| **Row-Level Security** | Tenant data isolation |
| **Soft Delete** | GDPR compliance |
| **Audit Logging** | Admin action tracking |
| **Privacy Requests** | Data export, correction, deletion workflows |
| **Consent Management** | User privacy consent tracking |

---

## 8. Plans & Entitlements

### 8.1 Plan Tiers

| Feature | Free | Starter | Pro | Ultra |
|---------|------|---------|-----|-------|
| **Max Chatbots** | 1 | 3 | 10 | 100 |
| **Monthly Tokens** | 100K | 500K | 1M | 5M |
| **Files per Bot** | 1 | 5 | 20 | 100 |
| **Total Files** | 5 | 25 | 100 | 1,000 |
| **File Size (MB)** | 5 | 10 | 20 | 50 |
| **Storage (MB)** | 10 | 100 | 500 | 2,000 |
| **OCR** | ❌ | ❌ | ✅ | ✅ |
| **JS Scraping** | ❌ | ✅ | ✅ | ✅ |
| **Guardrails** | ❌ | ❌ | ✅ | ✅ |
| **Smart Fallback** | ❌ | ❌ | ✅ | ✅ |
| **Human Handoff** | ❌ | ❌ | ❌ | ✅ |
| **Custom Branding** | ❌ | ❌ | ❌ | ✅ |
| **Suggested Questions** | 3 | 4 | 6 | 10 |
| **Chat RPM** | 30 | 60 | 100 | 500 |
| **Sources RPM** | 10 | 15 | 30 | 100 |

### 8.2 AI Models

| Plan | Default | Allowed Models |
|------|---------|----------------|
| **Free** | gpt-4o-mini | gpt-4o-mini |
| **Starter** | gpt-4o-mini | gpt-4o-mini, gpt-4o |
| **Pro** | gpt-4o | gpt-4o-mini, gpt-4o |
| **Ultra** | gpt-4o | gpt-4o-mini, gpt-4o |

### 8.3 Quota Enforcement

| Type | Mechanism |
|------|-----------|
| **Token Quota** | Reserve → Adjust pattern |
| **Chatbot Count** | Creation-time check |
| **File Limits** | Per-bot and total |
| **Rate Limits** | Redis sliding window |

---

## Summary

botla.app is a comprehensive AI chatbot platform featuring:

✅ **Multi-tenant architecture** with organizations, workspaces, and member management  
✅ **RAG pipeline** with semantic search, tiered confidence, and circuit-breaker protection  
✅ **Smart tools** (actions) for external API integration via function calling  
✅ **Human handoff** for escalation workflows with email collection  
✅ **Comprehensive analytics** and usage tracking with trends  
✅ **React 19 dashboard** with shadcn/ui, Apple-style design, full chatbot management  
✅ **Preact widget** with Shadow DOM isolation, markdown, suggestions, feedback  
✅ **Enterprise security** with HttpOnly cookies, rate limiting, SSRF protection  
✅ **Privacy compliance** (KVKK/GDPR) with consent management, data export, deletion workflows  
✅ **Background processing** for scalable ingestion with bounded worker pools  
✅ **Multi-language support** (Turkish UI, English error messages)  

---

*End of Comprehensive Feature Documentation*
