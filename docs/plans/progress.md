# Implementation Progress

## Completed Features

### Phase 1: Core Product Improvements

#### ✅ 1.2 Path-Based Include/Exclude Filtering (2025-12-07)
- **Goal**: Allow filtering scraped URLs via glob patterns (e.g., `/blog/*`).
- **Implementation**:
  - **DB**: Added `include_paths` and `exclude_paths` (TEXT[]) to `chatbots` table.
  - **Backend**: Created `PathFilter` engine in `internal/scraper`.
  - **Frontend**: Added `PathFilterSection` to `SourceUploader`.
- **Status**: Verified (All tests passed).

#### ✅ 1.3 CSS Selector Scraping (2025-12-07)
- **Goal**: Extract specific content using CSS selectors to reduce noise.
- **Implementation**:
  - **DB**: Added `selector_whitelist` (TEXT[]) to `chatbots` table.
  - **Backend**: Created `SelectorExtractor` in `internal/scraper` using `goquery`.
  - **Frontend**: Added "Advanced Scraping Settings" UI.
- **Status**: Verified (All tests passed).

#### ✅ 1.4 Sitemap Parser (2025-12-07)
- **Goal**: Auto-discover and bulk import URLs from sitemaps.
- **Implementation**:
  - **Backend**: Created `SitemapParser` (supports recursive indexes) and new endpoints (`/sitemap/discover`, `/sources/bulk`).
  - **Frontend**: Created `SitemapImport` component with bulk selection tools.
- **Status**: Verified (Parser & API tests passed).

#### ✅ 1.5 URL Checkbox UI (2025-12-08)
- **Goal**: Allow users to review and select discovered URLs before adding them as sources.
- **Implementation**:
  - **DB**: 
    - Migration `000011_pending_discovered_urls`: Created `pending_discovered_urls` table.
    - Added `discovery_mode` column to `chatbots` table (auto/pending/disabled).
  - **Backend**:
    - Created `internal/db/pending_url.go` with CRUD operations for pending URLs.
    - Updated `internal/processing/url_processor.go` with discovery mode logic.
    - Created `internal/api/handlers/pending_urls.go` with endpoints:
      - `GET /pending-urls` - List pending URLs with pagination
      - `POST /pending-urls/approve` - Approve and create sources
      - `POST /pending-urls/reject` - Reject pending URLs
      - `POST /pending-urls/clear` - Clear all pending URLs
    - Added routes to `cmd/server/main.go`.
  - **Frontend**:
    - Added pending URL API functions to `api/source.ts`.
    - Created `PendingURLsPanel.tsx` - Checkbox selection UI with approve/reject actions.
    - Created `DiscoveryModeSection.tsx` - Radio button UI for discovery mode selection.
    - Updated `useChatbotForm.ts` with `discoveryMode` state.
    - Integrated components into `ChatbotDetailPage.tsx` Sources tab.
- **Status**: Verified (Backend builds, frontend builds, unit tests passed).

#### ✅ 1.6 Auto-Refresh Scheduler (2025-12-08)
- **Goal**: Automatically refresh URL sources at scheduled intervals.
- **Implementation**:
  - **DB**:
    - Migration `000012_auto_refresh_config`: Added columns to `chatbots`:
      - `refresh_policy` (TEXT) - 'manual' or 'auto'
      - `refresh_frequency` (TEXT) - 'daily', 'weekly', 'monthly'
      - `next_refresh_at` (TIMESTAMPTZ) - Next scheduled refresh time
      - `last_refresh_at` (TIMESTAMPTZ) - Last refresh execution time
    - Added `auto_refresh_count` to `usage_ingestions` for tracking.
    - Created index `idx_chatbots_next_refresh` for efficient scheduler queries.
  - **Backend**:
    - Created `internal/services/refresh_scheduler.go`:
      - `RefreshScheduler` service with 5-minute polling loop
      - `CalculateNextRefresh()` for daily/weekly/monthly scheduling
      - Plan-based limit checking via `GetAutoRefreshCountForMonth()`
    - Added scheduler functions in `internal/db/chatbot.go`:
      - `GetChatbotsDueForRefresh()` - Find bots needing refresh
      - `UpdateChatbotRefreshTimes()` - Update next/last refresh timestamps
      - `GetAutoRefreshCountForMonth()` / `IncrementAutoRefreshCount()` - Usage tracking
      - `GetURLSourcesForChatbot()` - Get URL sources for a chatbot
    - Updated `internal/models/chatbot.go` with RefreshPolicy, RefreshFrequency, NextRefreshAt, LastRefreshAt
    - Updated `internal/api/handlers/chatbot.go` and `chatbot_item.go` for API support
    - Integrated scheduler startup/shutdown in `cmd/server/main.go`
  - **Frontend**:
    - Created `RefreshSettings.tsx` component with:
      - Policy selection (Manual/Auto) with visual cards
      - Frequency selection (Daily/Weekly/Monthly) buttons
      - Status display showing last/next refresh times
    - Updated `useChatbotForm.ts` with refreshPolicy, refreshFrequency, nextRefreshAt, lastRefreshAt
    - Integrated RefreshSettings into `extraUrlSettings` prop (only visible when "Web Sitesi" URL tab is active)
- **Note**: Auto-refresh only applies to URL sources (not PDFs/text) since they are dynamic content.
- **Status**: Verified (Backend builds, frontend builds, all tests passed).

#### ✅ 1.7 White-Label Branding (2025-12-08)
- **Goal**: Plan-based "Powered by Botla" removal and customizable branding options.
- **Implementation**:
  - **DB**:
    - Migration `000013_branding_options`: Added columns to `chatbots`:
      - `hide_branding` (BOOLEAN) - Hide "Powered by Botla" (Pro+ feature)
      - `custom_branding` (JSONB) - Custom branding config: {logo_url, text, link} (Enterprise feature)
  - **Backend**:
    - Updated `internal/models/chatbot.go`: Added `HideBranding` and `CustomBranding` fields
    - Updated `internal/models/plan.go`: Added `BrandingConfig` with `CanHideBranding` and `CanCustomBranding`
    - Updated `internal/db/chatbot.go`: All CRUD queries now include branding columns
    - Updated `internal/api/handlers/chatbot.go`: Added branding fields to request struct
    - Updated `internal/api/handlers/chatbot_item.go`: 
      - Plan-based validation for branding changes (403 Forbidden if plan doesn't allow)
      - Added branding fields to `applyChatbotUpdates()`
    - Updated `internal/api/handlers/public.go`: Added branding fields to public chatbot config response
  - **Widget**:
    - Updated `widget/src/components/ChatDrawer.tsx`:
      - Added `hideBranding` and `customBranding` props
      - Conditional rendering: custom branding > hide branding > default Botla branding
    - Updated `widget/src/widgetApp.tsx`: Passes branding props to ChatDrawer from config
  - **Frontend**:
    - Created `BrandingSettings.tsx` component with:
      - Toggle to hide "Powered by Botla" (locked for Free plan)
      - Custom branding inputs: logo URL, text, link (locked for non-Enterprise)
      - Live preview of branding changes
      - Plan feature indicators (Pro+/Enterprise badges)
    - Updated `useChatbotForm.ts` with `hideBranding` and `customBranding` state
- **Status**: Verified (Backend builds, frontend builds, widget builds, unit tests passed).

---

## Pending Roadmap

### Phase 1: Core Product Improvements
- [ ] 1.1 LLM Client Abstraction

### Phase 2: Integrations
- [ ] 2.1 Function Calling
- [ ] 2.2 Zapier Integration
- [ ] 2.3 Operator Handoff
- [ ] 2.4 Guardrails UI
- [ ] 2.5 Temperature/MaxTokens UI

### Phase 3: Agency and White-Label
- [ ] 3.1 Multi-Tenant Architecture
- [ ] 3.2 Custom Domain Routing
- [ ] 3.3 Advanced Analytics

