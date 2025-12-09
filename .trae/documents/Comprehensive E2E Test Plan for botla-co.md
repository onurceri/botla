## Goals
- Validate all user-facing flows, backend APIs, and the embeddable widget end-to-end.
- Prove RBAC correctness for org/workspace operations under all role combinations.
- Exercise external systems with stubs and minimal data volume to control costs.
- Cover edge cases: timeouts, partial failures, rate limits, retries, last-owner protection, and multi-tenant scoping.

## Tooling & Modes
- Playwright for dashboard and widget E2E: base at `http://localhost:5173` with `VITE_E2E=1` and stubbed API calls.
- Backend integration tests (Go) using `httptest.NewServer` via `internal/integration/testserver.go` with seeded DB schema and in-memory storage.
- Two execution modes:
  - Stubbed mode: LLM (`OPENAI_API_BASE`) and Qdrant stubs; minimal data sizes; zero external cost.
  - Real backend mode: Playwright drives the dashboard against local backend (`E2E_API_BASE=http://localhost:8080`) for selected smoke paths.

## Environment & Fixtures
- DB test schema via `internal/integration/testutils.go` seeds `languages` and `plans` and constrains connections.
- Auth bypass for UI via `storageState` and `VITE_E2E`; backend auth uses real JWT from `authToken()` when needed.
- Storage: memory storage for tests; R2 only in a targeted sanity test with small files.
- Stubs:
  - OpenAI: `startOpenAIStub`, `startOpenAIErrorStub`, `startOpenAITimeoutStub`.
  - Qdrant: `startQdrantStub`, error variants for search and upsert.
  - HTML: small pages for URL ingest.

## RBAC Matrix Coverage
- Roles: `owner`, `admin`, `member` from `internal/services/organization_service.go`.
- Middleware `RequireOrganizationAccess` enforces minimum roles via `hasMinRole` `pkg/middleware/organization.go:71-78`.
- Route-level requirements (examples):
  - Update/Delete org → `owner` `cmd/server/main.go:166-167`.
  - List workspaces → `member` `cmd/server/main.go:170`; create/update/delete → `admin` `cmd/server/main.go:171-173`.
  - Members list → `member` `cmd/server/main.go:176`; add/remove/update → `admin` `cmd/server/main.go:177-179`.
- Service-level constraints:
  - Prevent self-promotion `internal/services/organization_service.go:362-365`.
  - Only `owner` can assign `owner` `internal/services/organization_service.go:378-381`.
  - Prevent removing/demoting last owner `internal/services/organization_service.go:367-376`, `321-329`.
- Tests:
  - Verify access allowed/denied for each route per role (UI and API).
  - Validate service constraints on role changes and membership removals.
  - Confirm `caller_role` visibility in responses used by UI gating `internal/api/handlers/organization.go:226-265`.

## Feature Suites
### Authentication & Sessions
- UI: Register/Login, invalid inputs, server errors, logout, redirect to dashboard.
- Token refresh: expired access token triggers refresh; invalid/expired refresh results in forced logout.
- API: Invalid/missing `Authorization` rejected `pkg/middleware/auth.go`; refresh rotation and invalidation.

### Organizations & Workspaces (Multi-tenant)
- Create/update/delete org; switch current org; workspace CRUD; switcher UI reflects changes.
- RBAC: exercise `member/admin/owner` across org/workspace routes.
- Path scoping extracts `orgID` `pkg/middleware/organization.go:59-69`; header scoping honor `X-Organization-ID` where applicable.
- Edge: invalid `orgID`, non-member access, workspace of another org.

### Chatbot Creation & Management
- Create bot: required fields, duplicate names slug handling, edit settings (appearance/colors/branding/guardrails/handoff/discovery).
- Delete bot: confirm modal, cascading removal of sources and analytics.
- Edge: invalid payloads, non-existent bot IDs.

### Sources Ingestion (URLs, PDFs, Sitemaps, Filters)
- Text upload: tiny text stored via storage service; download during processing.
- PDF upload: tiny PDF processed; OCR fallback only under `fitz` tag.
- URL ingest: stub HTML; respect path-based filtering and CSS selector extraction.
- Sitemap import: small sitemap, invalid sitemap URL.
- Refresh & pending URLs: schedule refresh, mark pending, cancel.
- Edge: unsupported content-types, large file rejection, duplicate sources, removal cleanly deletes storage.

### Embedding & Widget
- Embed code panel renders correct snippet; copy action.
- Widget loads from build; bubble renders; drawer opens; config fetched; chat works with stubbed public endpoints.
- Secure embed: token issuance flow; shadow DOM assertions.
- Edge: cross-origin config failures, missing chatbot ID, network timeout.

### Chat & Feedback
- Chat success: minimal prompt; message list updates; suggestions shown.
- Failures: LLM 500, embeddings empty, Qdrant search errors; UI shows retries or error notice.
- Feedback: positive/negative feedback posts for a message ID.
- Edge: long prompts truncated, rate limit responses, idempotent retries.

### Analytics
- Usage charts render from API series; empty-state when no data.
- Filters: date ranges and per-chatbot view.
- Edge: large numbers formatting; missing series; backend error handling.

### Plans, Quotas, Model Providers
- Plan page loads; quota warnings trigger gating on ingest/chat.
- Switch default model provider where UI exists; verify correct provider headers are sent.
- Edge: over-quota rejection, model unavailable error mapping.

### Refresh Scheduling & Discovery
- Configure auto-refresh cadence; verify queue receives jobs.
- Discovery mode toggles; pending URLs panel updates.
- Edge: invalid schedules, overlapping refresh windows.

### Internationalization
- Language config fetch; UI language switch (if present).
- Edge: missing translations; RTL layout sanity.

### Health, Errors, Middleware
- `/health` responds OK.
- CORS allows tenant headers `pkg/middleware/cors.go:14-18`; preflight passes.
- Rate limit surfaces 429; request logging does not leak secrets.
- Error mapping returns structured codes/messages.

### Storage & Files (R2)
- Upload/download/delete using storage service for tiny files.
- Generate keys via `storage.GenerateKey`; ensure isolation by org/bot.
- Edge: missing key download returns 404; delete non-existent key handled.

## External Integrations Strategy
- LLMs: default to stubs; when needed, use minimal prompts and cheapest models; cap tokens via config.
- Qdrant: use stub with minimal vectors; verify collection creation and search filters.
- Browser: rod-based dynamic pages only in targeted tests; short timeouts; small pool.
- R2: run only a sanity test with a tiny file; skip by default in CI unless secrets present.

## Edge Cases & Reliability
- Timeouts and retries: LLM and Qdrant; ensure UI/APIs surface errors clearly.
- Idempotency: repeated source refresh and message feedback.
- Concurrency: parallel ingest jobs do not corrupt state.
- Pagination: members, chatbots, sources lists.
- Security: token tampering, cross-org access, last-owner protections.

## CI & Cost Controls
- Suites:
  - Smoke (PR): login, create bot, text ingest, chat, analytics load, RBAC basic.
  - Full (daily): all feature suites with stubs; include widget secure embed.
  - External sanity (manual/nightly): R2 tiny file; real backend mode; optional real LLM tiny prompt.
- Controls:
  - Always use stubs unless `E2E_USE_REAL` flags set.
  - Minimal content sizes; low parallelism for heavy tests; trace on-first-retry.
  - Token/ops budget guard to abort if thresholds exceeded.

## Deliverables
- Playwright specs per feature suite with shared helpers and fixtures.
- Go integration tests expanding current coverage for RBAC, ingest, chat, analytics, storage.
- CI pipeline stages and flags for stubbed vs real modes.
- Documentation for running locally and in CI, with env matrix and data budgets.

## Milestones
1. Baseline smoke suite + RBAC route matrix (UI/API) stubbed.
2. Sources ingestion + widget embed + chat failure scenarios.
3. Analytics + refresh scheduling + discovery mode coverage.
4. Storage sanity + optional real-provider smoke with strict budgets.
