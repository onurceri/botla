# botla.app Comprehensive Feature Documentation

> **Generated**: January 2026
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
| **Database** | PostgreSQL 15+ | Primary data storage |
| **Cache** | Redis | Rate limiting, session caching |
| **Vector DB** | Qdrant | Semantic search embeddings |
| **Storage** | Cloudflare R2 / AWS S3 | File storage |
| **Frontend Dashboard** | React 19 + Vite | Admin dashboard |
| **Widget** | Preact + Vite | Embeddable chat widget |
| **Infrastructure** | Docker Compose | Local development |

### 1.2 Project Structure

```
botla-app/
├── cmd/
│   ├── server/          # Backend entrypoint
│   └── cli/             # CLI utilities
├── internal/
│   ├── api/handlers/    # HTTP request handlers (77 files)
│   ├── auth/            # JWT & password utilities
│   ├── db/              # Database layer (sqlc-generated)
│   ├── integration/     # Integration tests (93 files)
│   ├── models/          # Domain models (22 files)
│   ├── processing/      # Background job processing
│   ├── rag/             # RAG pipeline (embeddings, LLM, vectors)
│   ├── repository/      # Data access layer (39 files)
│   ├── scraper/         # Web scraping (sitemap, selectors)
│   ├── services/        # Business logic (39 files)
│   └── workers/         # Background worker pool
├── pkg/
│   ├── config/          # Configuration (50+ env vars)
│   ├── langconfig/      # Language/localization (tr, en)
│   ├── logger/          # Structured JSON logging
│   ├── middleware/      # HTTP middleware stack
│   ├── storage/         # S3/R2 file storage
│   ├── ratelimit/       # Rate limiting (Redis/Memory)
│   └── ...              # Utilities
├── frontend/            # React dashboard
├── widget/              # Preact embeddable widget
└── db/migrations/       # Database migrations
```

---

## 2. Backend Features (Go)

### 2.1 Authentication & Authorization

#### 2.1.1 User Management

| Feature | Description |
|---------|-------------|
| **Registration** | Email/password with Argon2 hashing, password strength validation |
| **Login** | JWT token generation, refresh token rotation |
| **Password Validation** | 8+ chars, uppercase, lowercase, digit, special character |
| **Email Validation** | RFC 5322 format checking |
| **Platform Admin** | Superuser flag for system administration |

#### 2.1.2 JWT Authentication

```go
// Token Types
- Access Token:  1 hour expiry
- Refresh Token: 7 day expiry (rotated on use)

// Token Claims
- UserID: string (UUID)
- IsPlatformAdmin: bool
- TokenType: "access" | "refresh"
```

#### 2.1.3 Refresh Token Management

- SHA-256 hashed storage in database
- Token revocation support
- Automatic rotation on refresh
- Cookie-based and header-based delivery

### 2.2 Chatbot Management

#### 2.2.1 Chatbot Model (`internal/models/chatbot.go`)

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

#### 2.2.2 Chatbot Appearance (`internal/models/chatbot.go`)

| Field | Type | Description |
|-------|------|-------------|
| ThemeColor | string | Primary brand color (hex) |
| WelcomeMessage | string | Initial bot greeting |
| Position | string | Widget position (bottom-left, bottom-right) |
| BotMessageColor | string | Bot message bubble color |
| UserMessageColor | string | User message bubble color |
| BotMessageTextColor | string | Bot text color |
| UserMessageTextColor | string | User text color |
| ChatFontFamily | string | Font family for chat |
| ChatHeaderColor | string | Header background color |
| ChatHeaderTextColor | string | Header text color |
| ChatBackgroundColor | string | Chat background color |
| BubbleRadius | string | Border radius for bubbles |
| InputBackgroundColor | string | Input field background |
| InputTextColor | string | Input text color |
| SendButtonColor | string | Send button color |
| BotIcon | *string | Custom bot avatar URL |
| BotDisplayName | *string | Custom bot name |

#### 2.2.3 Chatbot Security

| Field | Type | Description |
|-------|------|-------------|
| SecureEmbedEnabled | bool | Require embed token for chat |
| AllowedDomains | *string | Comma-separated allowed domains |
| EmbedSecret | *string | Secret for signed embed tokens |

#### 2.2.4 Discovery & Scraping

| Field | Type | Description |
|-------|------|-------------|
| DiscoveryMode | string | auto, pending, disabled |
| IncludePaths | []string | URL patterns to include |
| ExcludePaths | []string | URL patterns to exclude |
| SelectorWhitelist | []string | CSS selectors for content |
| RefreshPolicy | string | manual, auto |
| RefreshFrequency | *string | daily, weekly, monthly |
| NextRefreshAt | *time.Time | Scheduled refresh time |
| LastRefreshAt | *time.Time | Last successful refresh |

#### 2.2.5 Suggestions & Engagement

| Field | Type | Description |
|-------|------|-------------|
| SuggestedQuestions | []string | AI-generated questions |
| ManualQuestions | []string | Admin-defined questions |
| SuggestionsEnabled | bool | Show suggestion carousel |

#### 2.2.6 Branding & Customization

| Field | Type | Description |
|-------|------|-------------|
| HideBranding | bool | Hide "Powered by Botla" |
| CustomBranding | *CustomBranding | Logo, text, link |

#### 2.2.7 Guardrails & Thresholds

| Field | Type | Description |
|-------|------|-------------|
| ConfidenceThreshold | float64 | Default confidence threshold |
| ThresholdConfig | *ThresholdConfig | Tiered threshold configuration |
| FallbackMessages | *FallbackMessages | Custom fallback messages |
| TopicRestrictions | *TopicConfig | Allowed/blocked topics |

#### 2.2.8 Human Handoff

| Field | Type | Description |
|-------|------|-------------|
| HandoffEnabled | bool | Enable human handoff |
| HandoffType | string | Type of handoff |
| HandoffConfig | *HandoffConfig | Handoff configuration |

#### 2.2.9 Chatbot CRUD Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List Chatbots | `ChatbotListHandler` | `GET /api/v1/chatbots` |
| Get Chatbot | `ChatbotItemHandler` | `GET /api/v1/chatbots/:id` |
| Create Chatbot | `ChatbotHandlers.Create` | `POST /api/v1/chatbots` |
| Update Chatbot | `ChatbotHandlers.Update` | `PUT /api/v1/chatbots/:id` |
| Delete Chatbot | `ChatbotHandlers.Delete` | `DELETE /api/v1/chatbots/:id` |
| Get Appearance | `ChatbotAppearanceHandler` | `GET /api/v1/chatbots/:id/appearance` |
| Update Appearance | `ChatbotAppearanceHandler` | `PUT /api/v1/chatbots/:id/appearance` |

### 2.3 Source Management (Data Ingestion)

#### 2.3.1 Source Types

| Type | Description |
|------|-------------|
| URL | Web page scraping |
| PDF | PDF document processing (with OCR support) |
| Text | Raw text content |
| Sitemap | XML sitemap crawling |

#### 2.3.2 Source Model (`internal/models/source.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| ChatbotID | string | Parent chatbot ID |
| SourceType | string | url, pdf, text, sitemap |
| SourceURL | *string | Original URL (for URL/sitemap types) |
| Title | *string | Source title |
| ContentHash | *string | SHA-256 of content (duplicate detection) |
| Status | string | pending, processing, completed, failed |
| ChunkCount | int | Number of text chunks |
| ErrorMessage | *string | Error details if failed |
| DiscoveredURLs | int | URLs found via discovery |
| PendingURLs | int | URLs awaiting processing |

#### 2.3.3 Source Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List Sources | `SourceHandlers.List` | `GET /api/v1/chatbots/:id/sources` |
| Create Source | `SourceHandlers.Create` | `POST /api/v1/chatbots/:id/sources` |
| Create PDF | `SourceCreateHandler` | `POST /api/v1/chatbots/:id/sources/pdf` |
| Create URL | `SourceSingleHandler` | `POST /api/v1/chatbots/:id/sources/url` |
| Create Sitemap | `SitemapHandler` | `POST /api/v1/chatbots/:id/sources/sitemap` |
| Bulk Create | `BulkSourceHandler` | `POST /api/v1/chatbots/:id/sources/bulk` |
| Refresh Source | `SourceRefreshHandler` | `POST /api/v1/sources/:id/refresh` |
| Get Chunks | `SourceChunksHandler` | `GET /api/v1/sources/:id/chunks` |
| Get Job Status | `TrainingJobHandler` | `GET /api/v1/sources/:id/job` |
| Retry Job | `TrainingJobHandler` | `POST /api/v1/sources/:id/job/retry` |
| Delete Source | `SourceHandlers.Delete` | `DELETE /api/v1/sources/:id` |
| List Pending | `PendingURLsHandler` | `GET /api/v1/chatbots/:id/pending-urls` |
| Approve Pending | `PendingURLsHandler` | `POST /api/v1/chatbots/:id/pending-urls/:url_id/approve` |
| Reject Pending | `PendingURLsHandler` | `POST /api/v1/chatbots/:id/pending-urls/:url_id/reject` |

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
| Get Config | `PublicHandlers.Config` | `GET /api/v1/public/chatbots/:id` |
| Submit Feedback | `ChatHandlers.Feedback` | `POST /api/v1/messages/:id/feedback` |
| Human Handoff | `HandoffHandlers.Request` | `POST /api/v1/public/chatbots/:id/handoff` |
| Submit Handoff Email | `HandoffHandlers.SubmitContact` | `POST /api/v1/public/chatbots/:id/handoff/:request_id/contact` |

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
| Get Action | `ActionHandlers.Get` | `GET /api/v1/chatbots/:id/actions/:action_id` |
| Create Action | `ActionHandlers.Create` | `POST /api/v1/chatbots/:id/actions` |
| Update Action | `ActionHandlers.Update` | `PUT /api/v1/chatbots/:id/actions/:action_id` |
| Delete Action | `ActionHandlers.Delete` | `DELETE /api/v1/chatbots/:id/actions/:action_id` |
| Execute Action | `ActionHandlers.Execute` | `POST /api/v1/chatbots/:id/actions/:action_id/execute` |
| Get Logs | `ActionLogsHandler` | `GET /api/v1/chatbots/:id/action-logs` |

### 2.7 Organization & Workspace Management

#### 2.7.1 Organization Model (`internal/models/organization.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| Name | string | Organization name |
| Slug | string | URL-safe identifier |
| OwnerID | string | Owner user ID |
| CreatedAt | time.Time | Creation timestamp |

#### 2.7.2 Workspace Model (`internal/models/workspace.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| OrganizationID | string | Parent organization ID |
| Name | string | Workspace name |
| Slug | string | URL-safe identifier |
| CreatedAt | time.Time | Creation timestamp |

#### 2.7.3 Organization/Workspace Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List Organizations | `OrganizationHandlers.List` | `GET /api/v1/organizations` |
| Get Organization | `OrganizationHandlers.Get` | `GET /api/v1/organizations/:id` |
| Create Organization | `OrganizationHandlers.Create` | `POST /api/v1/organizations` |
| Update Organization | `OrganizationHandlers.Update` | `PUT /api/v1/organizations/:id` |
| List Workspaces | `WorkspaceHandlers.List` | `GET /api/v1/organizations/:id/workspaces` |
| Create Workspace | `WorkspaceHandlers.Create` | `POST /api/v1/organizations/:id/workspaces` |
| Update Workspace | `WorkspaceHandlers.Update` | `PUT /api/v1/workspaces/:id` |
| Delete Workspace | `WorkspaceHandlers.Delete` | `DELETE /api/v1/workspaces/:id` |
| Get Settings | `WorkspaceSettingsHandler` | `GET /api/v1/workspaces/:id/settings` |
| Update Settings | `WorkspaceSettingsHandler` | `PUT /api/v1/workspaces/:id/settings` |

### 2.8 User & Profile Management

#### 2.8.1 User Model (`internal/models/user.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| Email | string | User email (unique) |
| PasswordHash | string | Argon2 hashed password |
| FullName | string | User's full name |
| AvatarURL | *string | Profile picture URL |
| PlanID | string | Subscription plan ID |
| IsPlatformAdmin | bool | Admin flag |
| LanguageCode | string | Preferred language |
| CreatedAt | time.Time | Registration timestamp |
| UpdatedAt | time.Time | Last update |
| DeletedAt | *time.Time | Soft delete |

#### 2.8.2 User Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| Get Profile | `MeHandlers.Get` | `GET /api/v1/me` |
| Update Profile | `MeHandlers.Update` | `PUT /api/v1/me` |
| Update Password | `MeHandlers.UpdatePassword` | `PUT /api/v1/me/password` |
| Get Usage | `UsageHandler` | `GET /api/v1/me/usage` |
| Delete Account | `PrivacyHandlers.DeleteAccount` | `DELETE /api/v1/me` |
| Export Data | `PrivacyHandlers.ExportData` | `GET /api/v1/me/export` |

### 2.9 Plans & Entitlements

#### 2.9.1 Plan Model (`internal/models/plan.go`)

| Field | Type | Description |
|-------|------|-------------|
| ID | string | UUID primary key |
| Name | string | Plan display name |
| Code | string | Plan code (free, starter, pro, enterprise) |
| MonthlyTokenLimit | int | Token quota |
| MaxChatbots | int | Chatbot limit |
| FilesPerBot | int | Max files per chatbot |
| MaxFileSizeMB | int | Max file upload size |
| TotalStorageMB | int | Total storage limit |
| OCREnabled | bool | OCR feature flag |
| DynamicScrapingEnabled | bool | JavaScript scraping |
| GuardrailsEnabled | bool | Custom guardrails |
| SmartFallbackEnabled | bool | AI-powered fallback |
| EscalateEnabled | bool | Human handoff |
| CustomBrandingEnabled | bool | White-label branding |
| MaxSuggestedQuestions | int | Suggestion carousel limit |
| RateLimitGlobal | int | Global RPM |
| RateLimitChat | int | Chat RPM |
| RateLimitSources | int | Sources API RPM |

#### 2.9.2 Plan Operations

| Operation | Handler | Endpoint |
|-----------|---------|----------|
| List Plans | `PlansHandler.List` | `GET /api/v1/plans` |
| Get Current Plan | `PlanHandlers.Get` | `GET /api/v1/me/plan` |
| Get Limits | `PlanHandlers.Limits` | `GET /api/v1/me/limits` |

### 2.10 Analytics

#### 2.10.1 Analytics Model (`internal/models/analytics.go`)

| Field | Type | Description |
|-------|------|-------------|
| ChatbotID | string | Parent chatbot ID |
| Date | time.Date | Aggregation date |
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
| Get Chatbot Analytics | `AnalyticsHandler.Chatbot` | `GET /api/v1/chatbots/:id/analytics` |
| Get Usage Stats | `UsageHandler` | `GET /api/v1/chatbots/:id/usage` |
| Get Source Stats | `SourceUsageStatsHandler` | `GET /api/v1/chatbots/:id/source-usage` |
| Get Insights | `InsightsHandler` | `GET /api/v1/chatbots/:id/insights` |

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
| Request Handoff | `HandoffHandlers.Request` | `POST /api/v1/chatbots/:id/handoff` |
| Submit Contact | `HandoffHandlers.SubmitContact` | `POST /api/v1/chatbots/:id/handoff/:request_id/contact` |
| List Requests | `HandoffHandlers.List` | `GET /api/v1/chatbots/:id/handoff-requests` |
| Update Status | `HandoffHandlers.UpdateStatus` | `PUT /api/v1/chatbots/:id/handoff-requests/:request_id` |

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

#### 2.14.1 Source Queue (`internal/processing/source_queue.go`)

- Async processing for web scraping, PDF parsing, embedding generation
- Persistent job queue with retry support
- Worker pool for parallel processing

#### 2.14.2 Refresh Scheduler (`internal/services/refresh_scheduler.go`)

- Scheduled content refresh for sources
- Daily, weekly, monthly refresh policies
- Automatic cooldown enforcement

#### 2.14.3 Retention Job (`internal/services/retention_job.go`)

- KVKK/GDPR compliance data cleanup
- Automatic deletion of old data
- Configurable retention periods

### 2.15 Web Scraping

#### 2.15.1 Scraper Features

| Feature | Description |
|---------|-------------|
| **Static Scraping** | HTML content extraction |
| **Dynamic Scraping** | JavaScript-rendered content (Puppeteer) |
| **Sitemap Parsing** | XML sitemap crawling |
| **CSS Selector Targeting** | Specific element extraction |
| **Path Filtering** | Include/exclude URL patterns |
| **Discovery Mode** | Auto-discover subpages |
| **Rate Limiting** | Respect robots.txt, crawl delay |
| **ETag Support** | Conditional requests for updates |

#### 2.15.2 SSRF Protection (`pkg/urlutil/ssrf.go`)

- Blocks private IP ranges (10.x, 172.16.x, 192.168.x)
- Blocks localhost and metadata endpoints
- Supports allowlist for internal services
- Configurable strict/relaxed mode

### 2.16 PDF Processing

| Feature | Description |
|---------|-------------|
| **Text Extraction** | Fitz-based PDF text extraction |
| **OCR Support** | Tesseract for image-based PDFs (Pro+) |
| **Multi-page** | Handles multi-page documents |
| **Metadata Extraction** | Title, author, page count |

---

## 3. Frontend Features (React)

### 3.1 Routing Structure

#### 3.1.1 Public Routes

| Path | Component | Description |
|------|-----------|-------------|
| `/` | `LandingPage` | Marketing landing page |
| `/login` | `LoginPage` | User login |
| `/register` | `RegisterPage` | User registration |

#### 3.1.2 Protected Routes (Dashboard)

| Path | Component | Description |
|------|-----------|-------------|
| `/dashboard` | `DashboardLayout` | Dashboard with sidebar |
| `/dashboard/chatbots` | `ChatbotsPage` | List all chatbots |
| `/dashboard/chatbots/:id` | `ChatbotDetailPage` | Chatbot detail with tabs |
| `/dashboard/chatbots/:id/settings` | `SettingsTab` | Chatbot settings |
| `/dashboard/chatbots/:id/security` | `SecurityTab` | Security configuration |
| `/dashboard/chatbots/:id/sources` | `SourcesTab` | Source management |
| `/dashboard/chatbots/:id/actions` | `ActionsTab` | Smart actions |
| `/dashboard/chatbots/:id/playground` | `PlaygroundTab` | Chat playground |
| `/dashboard/chatbots/:id/deploy` | `DeployTab` | Deployment/embed code |
| `/dashboard/chatbots/:id/insights` | `InsightsTab` | Analytics & insights |
| `/settings/profile` | `ProfilePage` | User profile |
| `/settings/organization` | `OrganizationSettingsPage` | Organization settings |
| `/settings/workspace` | `WorkspaceSettingsPage` | Workspace settings |
| `/settings/plan` | `PlanPage` | Subscription management |
| `/settings/privacy` | `PrivacySettingsPage` | Privacy controls |
| `/onboarding` | `OnboardingPage` | First-time setup |

#### 3.1.3 Admin Routes

| Path | Component | Description |
|------|-----------|-------------|
| `/admin` | `AdminLayout` | Admin dashboard layout |
| `/admin/users` | `AdminUsersPage` | User management |
| `/admin/organizations` | `AdminOrganizationsPage` | Organization management |
| `/admin/chatbots` | `AdminChatbotsPage` | Chatbot management |
| `/admin/sources` | `AdminSourcesPage` | Source monitoring |
| `/admin/system` | `AdminSystemPage` | System health |
| `/admin/queues` | `AdminQueuesPage` | Job queue status |
| `/admin/errors` | `AdminErrorsPage` | Error log viewer |
| `/admin/audit` | `AdminAuditPage` | Audit logs |
| `/admin/privacy` | `AdminPrivacyPage` | Privacy compliance |

### 3.2 Authentication & Session

| Feature | Description |
|---------|-------------|
| **JWT Validation** | Token format and expiry checking |
| **Session Persistence** | localStorage token storage |
| **Session Expiry** | Event-based session expired detection |
| **Protected Routes** | Automatic redirect to login |
| **Platform Admin** | Admin route protection |

### 3.3 Chatbot Management Components

#### 3.3.1 Chatbot List Page (`ChatbotsPage.tsx`)

| Feature | Description |
|---------|-------------|
| **Chatbot Cards** | Display with name, model, status |
| **Create Dialog** | New chatbot creation form |
| **Search/Filter** | Find chatbots by name |
| **Pagination** | Paginated list view |

#### 3.3.2 Chatbot Detail Page (`ChatbotDetailPage.tsx`)

| Tab | Features |
|-----|----------|
| **Settings** | Name, description, language, model, instructions |
| **Security** | Embed security, domain restrictions |
| **Sources** | Add/manage URLs, PDFs, text, sitemaps |
| **Actions** | Create/edit smart tools |
| **Playground** | Test chat interface |
| **Deploy** | Embed code, widget settings |
| **Insights** | Analytics, usage charts |

#### 3.3.3 Settings Tab Components

| Component | Features |
|-----------|----------|
| `IdentitySection` | Bot name, description, avatar |
| `AppearanceSection` | Colors, fonts, positioning |
| `ParameterBuilder` | Temperature, max tokens |
| `ChatbotSidebar` | Navigation sidebar |

#### 3.3.4 Sources Tab Components

| Component | Features |
|-----------|----------|
| `SourcesTab` | Source list and management |
| `SourceCard` | Source display with status |
| `PathFilterSection` | URL path filters |
| `RefreshSettings` | Auto-refresh configuration |
| `ChunkInspector` | View chunk content |

#### 3.3.5 Actions Tab Components

| Component | Features |
|-----------|----------|
| `ActionsTab` | Action list |
| `ActionLogs` | Execution history |
| `ParameterBuilder` | Tool parameter configuration |

#### 3.3.6 Playground Tab Components

| Component | Features |
|-----------|----------|
| `PlaygroundTab` | Chat interface |
| `PlaygroundConsole` | Chat messages display |
| `EmbedCodePanel` | Widget embed code |
| `SuggestionsPanel` | Suggested questions |

### 3.4 Analytics Features

#### 3.4.1 Analytics Dashboard

| Component | Features |
|-----------|----------|
| `ChatbotAnalytics` | Message volume, token usage |
| `SourceUsageStats` | Source performance metrics |

### 3.5 Organization Features

| Component | Features |
|-----------|----------|
| `OrganizationContext` | Current org state management |
| `OrganizationSwitcher` | Switch between organizations |
| `CreateOrganizationDialog` | Create new organization |
| `CreateWorkspaceDialog` | Create new workspace |

### 3.6 Admin Features

| Component | Features |
|-----------|----------|
| `HealthPanel` | System health indicators |
| `StatsCard` | Metric display cards |
| `AdminRoute` | Admin access guard |

### 3.7 API Layer (`frontend/src/api/`)

| Module | Features |
|--------|----------|
| `auth.ts` | Login, register, logout, token refresh |
| `chatbot.ts` | CRUD operations, appearance, domain |
| `source.ts` | Source CRUD, upload, refresh |
| `chat.ts` | Chat messages, feedback |
| `action.ts` | Action CRUD, execution, logs |
| `organization.ts` | Organization management |
| `workspace.ts` | Workspace management |
| `user.ts` | Profile, password, usage |
| `analytics.ts` | Usage stats, insights |
| `handoff.ts` | Handoff requests |
| `plan.ts` | Plans, limits |
| `admin.ts` | Admin operations |

### 3.8 State Management

| Approach | Library | Usage |
|----------|---------|-------|
| **Server State** | TanStack Query | API data caching |
| **Global State** | React Context | Organization, theme |
| **Local State** | useState, useReducer | Component state |
| **URL State** | React Router | Route parameters |

### 3.9 UI Components

| Component | Description |
|-----------|-------------|
| `Button` | Reusable button with variants |
| `Input` | Form input with validation |
| `Card` | Container card component |
| `Dialog` | Modal dialog |
| `Tabs` | Tab navigation |
| `Toast` | Notification toasts |
| `Tooltip` | Hover tooltips |
| `Select` | Dropdown select |
| `Switch` | Toggle switch |
| `Badge` | Status badge |
| `Avatar` | User/bot avatar |
| `Spinner` | Loading spinner |

### 3.10 Form Handling

| Feature | Library | Description |
|---------|---------|-------------|
| **Forms** | React Hook Form | Form state management |
| **Validation** | Zod | Schema validation |
| **UI Library** | Radix UI | Accessible primitives |

### 3.11 Charts & Analytics

| Library | Usage |
|---------|-------|
| **Recharts** | Usage charts, feedback trends |

### 3.12 Internationalization

| Feature | Status |
|---------|--------|
| **Turkish (tr)** | Full support |
| **English (en)** | Full support |
| **API Messages** | Localized error messages |

---

## 4. Widget Features (Preact)

### 4.1 Widget Architecture

```
widget/src/
├── components/     # UI components
├── hooks/          # Custom hooks
├── services/       # Business logic
├── stores/         # State management
├── lib/            # Utilities
├── i18n/           # Translations
├── config/         # Configuration
└── api/            # API client
```

### 4.2 Components

#### 4.2.1 Main Components

| Component | File | Description |
|-----------|------|-------------|
| `WidgetApp` | `widgetApp.tsx` | Main widget container |
| `ChatBubble` | `ChatBubble.tsx` | Launcher button with unread badge |
| `ChatDrawer` | `ChatDrawer.tsx` | Chat conversation panel |
| `Message` | `Message.tsx` | Message bubble component |
| `Suggestions` | `Suggestions.tsx` | Suggestion carousel |

### 4.3 Widget Configuration

| Parameter | Type | Description |
|-----------|------|-------------|
| `chatbot-id` | string | Chatbot UUID (required) |
| `api-base` | string | API base URL override |
| `color` | string | Theme color (hex) |
| `welcome` | string | Welcome message override |
| `position` | string | bottom-left or bottom-right |
| `auto-open` | boolean | Auto-open on load |
| `header-color` | string | Header background color |
| `header-text-color` | string | Header text color |
| `bot-message-color` | string | Bot bubble color |
| `user-message-color` | string | User bubble color |
| `bot-message-text-color` | string | Bot text color |
| `user-message-text-color` | string | User text color |
| `font-family` | string | Chat font family |
| `panel-height` | string | Panel height (CSS) |
| `panel-width` | string | Panel width (CSS) |
| `panel-bg-color` | string | Panel background |
| `input-bg-color` | string | Input background |
| `input-text-color` | string | Input text color |
| `chat-bg-color` | string | Chat area background |
| `bubble-radius` | string | Bubble border radius |
| `send-button-color` | string | Send button color |
| `bot-name` | string | Bot display name |
| `bot-icon` | string | Bot avatar URL |
| `hide-branding` | boolean | Hide "Powered by Botla" |
| `custom-branding` | JSON | Custom branding object |
| `preview-mode` | boolean | Preview/embed mode |
| `position-strategy` | string | fixed or absolute |
| `suggestions` | JSON | Override suggestions |
| `embed-token-url` | string | URL to fetch embed token |
| `captcha-site-key` | string | hCaptcha site key |
| `reset-session` | boolean | Reset session on load |
| `session-id` | string | Override session ID |
| `use-url-overrides` | boolean | Use URL param overrides |

### 4.4 State Management

#### 4.4.1 Chat Store (`stores/chatStore.ts`)

| State | Type | Description |
|-------|------|-------------|
| `messages` | ChatMessage[] | Conversation history |
| `currentSession` | ChatSession | Active session data |
| `typingUsers` | string[] | Currently typing users |
| `readState` | Record<string, boolean> | Read/unread tracking |

#### 4.4.2 Session Store (`stores/sessionStore.ts`)

| State | Type | Description |
|-------|------|-------------|
| `sessionId` | string | Current session ID |
| `userToken` | string | Authentication token |
| `metadata` | Record<string, unknown> | Session metadata |
| `createdAt` | Date | Session creation time |

### 4.5 Custom Hooks

| Hook | File | Features |
|------|------|----------|
| `useChat` | `hooks/useChat.ts` | Send messages, manage conversation |
| `useSession` | `hooks/useSession.ts` | Session lifecycle, persistence |
| `useWidget` | `hooks/useWidget.ts` | Visibility, configuration |

### 4.6 Services

#### 4.6.1 API Service (`services/api.ts`)

| Method | Description |
|--------|-------------|
| `sendMessage()` | Send chat message |
| `getChatHistory()` | Retrieve conversation |
| `createSession()` | Initialize session |
| `endSession()` | Close session |
| `getWidgetConfig()` | Fetch widget config |
| `trackEvent()` | Send analytics |

#### 4.6.2 Session Service (`services/session.ts`)

| Method | Description |
|--------|-------------|
| `initialize()` | Set up new session |
| `resume()` | Restore previous session |
| `save()` | Persist session state |
| `clear()` | Clear session data |

### 4.7 Internationalization

#### 4.7.1 Supported Languages

| Language | File | Status |
|----------|------|--------|
| **English** | `i18n/locales/en.ts` | Complete |
| **Turkish** | `i18n/locales/tr.ts` | Complete |

#### 4.7.2 Translation Keys

| Key | Description |
|-----|-------------|
| `welcomeMessage` | Initial greeting |
| `placeholder` | Input placeholder |
| `send` | Send button |
| `loading` | Loading text |
| `error` | Error message |
| `attachments` | Attachment button |
| `minimize` | Minimize label |
| `close` | Close label |

### 4.8 Widget Features

| Feature | Description |
|---------|-------------|
| **Real-time Chat** | SSE streaming for responses |
| **Markdown Rendering** | Rich text support |
| **Suggestions** | AI-generated question carousel |
| **Feedback** | Thumbs up/down on messages |
| **Human Handoff** | Email capture for support |
| **Session Persistence** | localStorage for conversation |
| **Typing Indicator** | Animated typing dots |
| **Unread Badge** | Message count on launcher |
| **Custom Styling** | CSS variables for theming |
| **Shadow DOM** | Style isolation |
| **Google Fonts** | Dynamic font loading |
| **Captcha Support** | hCaptcha integration |
| **Embed Tokens** | Secure embed authentication |
| **Event Emission** | postMessage to parent window |
| **Preview Mode** | Playground integration |

### 4.9 Utility Functions

| Function | File | Description |
|----------|------|-------------|
| `formatTimestamp()` | `lib/utils.ts` | Format message time |
| `truncateText()` | `lib/utils.ts` | Truncate long text |
| `generateId()` | `lib/utils.ts` | Generate unique ID |
| `debounce()` | `lib/utils.ts` | Debounce function |
| `throttle()` | `lib/utils.ts` | Throttle function |
| `isMobile()` | `lib/utils.ts` | Mobile detection |
| `escapeHtml()` | `lib/utils.ts` | HTML sanitization |
| `sanitizeUrl()` | `utils/sanitize.ts` | URL validation |

---

## 5. Database Schema

### 5.1 Core Tables

| Table | Description |
|-------|-------------|
| `users` | User accounts |
| `organizations` | Tenant organizations |
| `workspaces` | Workspace isolation |
| `chatbots` | Chatbot configurations |
| `sources` | Data sources (URL, PDF, text) |
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

### 5.2 Migration History

The system has evolved through **52 database migrations** covering:

| Category | Features |
|----------|----------|
| **Core** | Users, chatbots, conversations, messages |
| **Multi-tenancy** | Organizations, workspaces |
| **Sources** | URL, PDF, text, sitemap types |
| **RAG** | Chunks, embeddings, search |
| **Actions** | Tool definitions, execution logs |
| **Guardrails** | Topic restrictions, thresholds |
| **Handoff** | Human escalation workflow |
| **Analytics** | Usage tracking, feedback |
| **Branding** | Custom appearance options |
| **Privacy** | KVKK/GDPR compliance |

---

## 6. API Reference

### 6.1 API Versioning

- **Base Path**: `/api/v1`
- **Format**: JSON
- **Authentication**: Bearer JWT (header or cookie)

### 6.2 Endpoint Categories

| Category | Base Path | Auth |
|----------|-----------|------|
| Auth | `/api/v1/auth/*` | Public |
| Chatbots | `/api/v1/chatbots/*` | Protected |
| Sources | `/api/v1/sources/*` | Protected |
| Chat | `/api/v1/chatbots/:id/chat` | Protected |
| Public | `/api/v1/public/*` | Public/Optional |
| User | `/api/v1/me/*` | Protected |
| Organization | `/api/v1/organizations/*` | Protected |
| Admin | `/api/v1/admin/*` | Platform Admin |

### 6.3 Rate Limits

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
| **JWT Access Token** | Bearer header | API authentication |
| **JWT Refresh Token** | Header/Cookie | Token renewal |
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

### 7.3 Security Middleware

| Middleware | Purpose |
|------------|---------|
| `AuthMiddleware` | JWT verification |
| `CORSMiddleware` | CORS policy enforcement |
| `RateLimiter` | DoS protection |
| `RecoveryMiddleware` | Panic recovery |
| `SecurityHeaders` | HTTP security headers |
| `MaxBytes` | Request size limit |
| `SSRFValidator` | URL attack prevention |

### 7.4 Data Protection

| Feature | Description |
|---------|-------------|
| **Argon2 Hashing** | Password storage |
| **Token Hashing** | SHA-256 refresh token storage |
| **SSRF Protection** | Blocks private IP access |
| **Row-Level Security** | Tenant data isolation |
| **Soft Delete** | GDPR compliance |
| **Audit Logging** | Admin action tracking |

---

## 8. RAG Pipeline

### 8.1 Pipeline Stages

```
[Source Input]
      ↓
[Content Fetch] → [HTML Parsing / PDF Text Extraction]
      ↓
[Chunking] → [Sentence-aware splitting with 15% overlap]
      ↓
[Embedding] → [OpenAI text-embedding-3-small]
      ↓
[Vector Storage] → [Qdrant collection]
      ↓
[Query] → [Embed query vector]
      ↓
[Search] → [Tiered similarity search (High/Medium/Low)]
      ↓
[LLM Generation] → [OpenAI GPT-4o / GPT-4o-mini]
      ↓
[Response] → [With citations and sources]
```

### 8.2 Chunking Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Target Tokens** | 512 | Desired chunk size |
| **Overlap** | ~15% | Tail overlap for context |
| **Boundaries** | Paragraph/Sentence | Preserves semantic units |
| **Abbreviation Handling** | Configurable | Prevents sentence splitting |

### 8.3 Search Tiers

| Tier | Threshold | Description |
|------|-----------|-------------|
| **High** | ≥ 0.50 | Strong match |
| **Medium** | 0.30 - 0.49 | Weak match (with warning) |
| **Low** | < 0.30 | No good match (fallback) |

### 8.4 Fallback Modes

| Mode | Trigger | Behavior |
|------|---------|----------|
| **Static** | Low confidence | Generic "I don't know" |
| **Smart** | Low confidence | AI redirect based on capabilities |
| **Escalate** | Low confidence | Human handoff workflow |

### 8.5 Topic Extraction

During ingestion, each source undergoes:
- **Capability Summary** - What the source covers
- **Suggested Questions** - Auto-generated Q&A
- **Language Detection** - Content language

### 8.6 Circuit Breaker

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Failure Ratio** | 0.50 | Trip on 50% failure |
| **Timeout** | 30s | Half-open state duration |
| **Requests** | 3 | Requests in half-open |
| **Interval** | 60s | Sample window |

---

## 9. Plans & Entitlements

### 9.1 Plan Tiers

| Feature | Free | Pro | Ultra |
|---------|------|-----|-------|
| **Max Chatbots** | 1 | 10 | 100 |
| **Monthly Tokens** | 100K | 1M | 5M |
| **Files per Bot** | 1 | 20 | 100 |
| **Total Files** | 5 | 100 | 1,000 |
| **File Size (MB)** | 5 | 20 | 50 |
| **Storage (MB)** | 10 | 500 | 2,000 |
| **OCR** | ❌ | ✅ | ✅ |
| **JS Scraping** | ❌ | ✅ | ✅ |
| **Guardrails** | ❌ | ✅ | ✅ |
| **Smart Fallback** | ❌ | ✅ | ✅ |
| **Human Handoff** | ❌ | ❌ | ✅ |
| **Custom Branding** | ❌ | ❌ | ✅ |
| **Suggested Questions** | 3 | 6 | 10 |
| **Chat RPM** | 30 | 100 | 500 |
| **Sources RPM** | 10 | 30 | 100 |

### 9.2 AI Models

| Plan | Default | Allowed Models |
|------|---------|----------------|
| **Free** | gpt-4o-mini | gpt-4o-mini |
| **Pro** | gpt-4o | gpt-4o-mini, gpt-4o |
| **Ultra** | gpt-4o | gpt-4o-mini, gpt-4o, gpt-5 |

### 9.3 Quota Enforcement

| Type | Mechanism |
|------|-----------|
| **Token Quota** | Reserve → Adjust pattern |
| **Chatbot Count** | Creation-time check |
| **File Limits** | Per-bot and total |
| **Rate Limits** | Redis sliding window |

---

## Summary

botla.app is a comprehensive AI chatbot platform featuring:

✅ **Multi-tenant architecture** with organizations and workspaces  
✅ **RAG pipeline** with semantic search and tiered confidence  
✅ **Smart tools** (actions) for external API integration  
✅ **Human handoff** for escalation workflows  
✅ **Comprehensive analytics** and usage tracking  
✅ **React dashboard** with full chatbot management  
✅ **Preact widget** with extensive customization  
✅ **Enterprise security** with JWT, rate limiting, SSRF protection  
✅ **Multi-language support** (Turkish, English)  
✅ **Background processing** for scalable ingestion  
✅ **KVKK/GDPR compliance** with data retention policies  

---

*End of Comprehensive Feature Documentation*
