## Backend Test Enhancements
- OpenAI stubs: expand `internal/integration/*chat*_test.go` to cover embeddings decode errors, chat 4xx/5xx, latency/timeout using `httptest.Server` via `OPENAI_API_BASE` and `internal/rag/openai.go` retry paths.
- Qdrant stubs: extend `internal/integration/sources_test.go` and `chat_qdrant_error_test.go` to simulate collection init failures, upsert/delete/search 4xx/5xx, empty hits vs errors, and network timeouts against `internal/rag/qdrant.go`.
- Analytics validation: add integration tests asserting per‑day increments for messages/conversations after chat flow in `internal/api/handlers/chat.go` and thumbs up/down after feedback; verify DB upserts in `internal/db/analytics.go` and API series from `internal/api/handlers/analytics.go` align. If thumbs need charting, extend API to include them.
- Auth edges: add tests for invalid Bearer formats, expired tokens in `pkg/middleware/auth.go`, refresh rotation race conditions and revoked refresh handling in `internal/api/handlers/auth.go`.
- Rate‑limit: add per‑IP vs per‑user isolation tests; implement `X‑RateLimit‑Limit`, `X‑RateLimit‑Remaining`, `Retry‑After` headers in `pkg/middleware/ratelimit.go`; wire middleware in `cmd/server/main.go` (currently only test mux uses it) and validate 429 behavior.
- Scraper: preserve conditional skips for dynamic rendering (`internal/scraper/browser_test.go`); add deterministic fixtures for static/dynamic pages; assert `visibleText` parsing correctness in `internal/scraper/worker.go`.

## Frontend Test Enhancements
- Sources polling: unit test `frontend/src/pages/ChatbotDetailPage.tsx` to verify pending → processing → completed transitions; stop after terminal/timeout; ensure `refreshSources()` is called. Use jest fake timers and MSW.
- URL validation: add client‑side format validation in `SourceUploader.tsx`; test invalid format and unreachable URL; assert toast message contents and `error` severity.
- PDF constraints: test backend 413 and 400 responses and confirm consistent error messaging (`SourceUploader.tsx`); optionally add client‑side size/type guards.
- Chat UI edge‑cases: add loading guards to prevent rapid sends in `ChatbotDetailPage.tsx` and widget; confirm empty‑message guard and disabled state correctness with tests.
- Analytics: test date‑range filtering correctness, totals recalculation, and loading/empty/error visual states in `frontend/src/pages/AnalyticsPage.tsx`.
- Auth/refresh: test concurrent 401s ensure single retry using `_retry` flag in `frontend/src/api/client.ts`; verify redirect to `/login` on refresh failure.

## End‑to‑End (E2E) Smoke (Optional)
- Flow: Login → Create Chatbot → Add Source (Text) → Chat → Thumbs Up → View Analytics using Playwright/Cypress.
- Data seeding: add fixtures and cleanup routines; isolate accounts per run.
- CI execution: mark as smoke, parallelizable; skip heavy scraper scenarios.

## CI and Coverage Gates (Optional)
- Coverage thresholds: frontend/backend ≥ 80%, critical paths ≥ 90%.
- Jobs: backend (`go test ./...`), frontend (`npm run lint`, `tsc --noEmit`, `npm test`); cache modules.
- Flaky mitigation: conditionally skip dynamic scraper if environment missing; stabilize timeouts.

## Security and Config
- Secrets: scan and ensure keys are never logged; verify env presence (`OPENAI_API_KEY`, `QDRANT_URL/API_KEY`, `R2_*`).
- CORS and rate‑limit: confirm allowed origins and sensible defaults in pre‑prod; enable rate limiter in `cmd/server/main.go`.
- Migrations: ensure latest schema applied; document rollback plan.

## Pre‑Prod Checklist
- All tests green (unit/integration/frontend, optional E2E).
- Coverage thresholds met; reports stored.
- Smoke run of critical flows on staging.
- Error monitoring hooked; request logs sampled.
- Feature docs reviewed and updated (`docs/features/*`).

## References
- Feedback protection: `cmd/server/main.go`, `internal/integration/testserver.go`.
- Memory storage for tests: `pkg/storage/memory.go`.
- Analytics page: `frontend/src/pages/AnalyticsPage.tsx` (filter depends on `data`).
- Key tests to extend: `internal/integration/*_test.go`, `frontend/src/pages/__tests__/*`, `frontend/src/components/chatbot/__tests__/*`. 

Confirm and I will implement tests, headers, and UI guards, verify locally, and share results.