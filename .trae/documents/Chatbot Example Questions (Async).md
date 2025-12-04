**Approach**

* Generate example questions during ingestion via structured LLM output; store per‑source suggestions; aggregate to chatbot; expose in public config; show asynchronously in the widget.

**Backend Changes**

* Migrations:

  * `data_sources.suggested_questions JSONB NULL`.

  * `chatbots.suggested_questions JSONB NULL` (aggregated list).

* Ingestion:

  * New `rag.ExtractIngestionMetadata(ctx, client, content, lang) -> { capability_summary, suggested_questions }` replacing `ExtractTopics` calls in `internal/processing/sources_queue.go` (both URL/PDF/Text flows).

  * Prompt enforces JSON only output, localized via `pkg/langconfig` and requests 3–6 short, diverse, answerable questions.

  * Parse JSON; on failure derive templated questions from summary.

  * Store per‑source: `UpdateSourceCapability`, `UpdateSourceSuggestions`.

  * Aggregate per chatbot: simple query across its sources → unique, normalized, capped N suggestions; store in `chatbots.suggested_questions`.

* Public API:

  * Extend `publicChatbot` with `suggested_questions` and set it from chatbot row.

**Widget UI**

* New `components/Suggestions.tsx` chips.

* Read `suggested_questions` from public config in `widgetApp.tsx` and render suggestions in `ChatDrawer.tsx` above input until first user message.

* Clicking a suggestion fills `input` and calls `send()` immediately; disable while `loading`.

**100% Test Coverage Plan**

* Go unit tests (package‑level):

  * `internal/rag/topic_extractor.go` (new):

    * Valid JSON path: returns summary + questions.

    * Invalid/extra text path: JSON parse failure → fallback generation.

    * Language enforcement (TR/EN): localized prompt selection.

  * `internal/processing/sources_queue.go`:

    * For each source type (url/pdf/text): success path stores capability + suggestions, chunks, embeddings; failure paths: empty content, storage errors, PDF errors; branch coverage for each `case` block and early exits.

    * Aggregation function: dedupe, cap length/count, normalization, persistence to chatbot.

  * `internal/db/source.go` and `internal/db/chatbot.go`:

    * `UpdateSourceSuggestions` success/failure; CRUD includes new column.

  * `internal/api/handlers/public.go`:

    * `PublicChatbotConfig` includes `suggested_questions` field; 404/500 paths unchanged; success returns expected JSON.

* Go integration tests (`internal/integration`):

  * Ingest pipeline end‑to‑end with a mock LLMClient that returns controlled JSON; verify data\_sources and chatbot rows; public config includes suggestions.

  * Edge: malformed LLM output triggers fallback and still populates chatbot suggestions.

* Frontend widget unit tests (Vitest):

  * `widget/src/components/__tests__/Suggestions.test.tsx`:

    * Render with items; disabled state; click calls handler.

  * `widget/src/components/__tests__/ChatDrawer.test.tsx` (extend existing):

    * Shows suggestions when no user messages; hides after first user message; disabled during `loading`.

  * `widget/src/__tests__/widgetApp.test.tsx`:

    * Loads suggestions from config; clicking suggestion sets `input` and triggers `send()`; respects `loading`.

* Frontend e2e (Playwright):

  * Extend `frontend/e2e/widget-embed.spec.ts`:

    * Widget loads; suggestions appear; click sends; assistant response shows; suggestions disappear after first send.

  * Secure embed variant unchanged.

**Coverage Measurement & Definition of Done**

* Go:

  * Run `go test ./... -coverprofile=cover.out` and `go tool cover -func=cover.out`; ensure 100% for modified packages: `internal/rag`, `internal/processing`, `internal/db`, `internal/api/handlers`.

* Widget:

  * Vitest config: enable coverage and thresholds at 100% for `widget/src/**/*.{ts,tsx}`; run `npm test -- --coverage` (or `pnpm vitest --coverage`) inside `widget`.

* Frontend e2e: run Playwright suites to validate behavior; not counted in unit coverage but required for end‑to‑end validation.

* CI: add gates to fail if coverage < 100% for targeted paths.

**Performance & Safety**

* Ingestion‑time generation avoids runtime latency; widget loads suggestions asynchronously with no blocking.

* Enforce max items (≤6) and max length (≤120 chars); strip control characters; dedupe.

* Keep

