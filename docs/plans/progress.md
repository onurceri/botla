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

---

## Pending Roadmap

### Phase 1: Core Product Improvements
- [ ] 1.1 LLM Client Abstraction
- [ ] 1.6 Auto-Refresh Scheduler
- [ ] 1.7 White-Label Branding

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

