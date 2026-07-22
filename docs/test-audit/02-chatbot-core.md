# Test Coverage Audit — Chatbot Core (chatbot-management, chat-operations, human-handoff)

**Scope:** All 30 spec docs in `docs/features/chatbot-management/`, `docs/features/chat-operations/`, and `docs/features/human-handoff/` read in full. Scenario groups mapped against Go unit tests (`internal/api/handlers`, `internal/services`, `internal/rag`, `internal/repository`), Go integration (`internal/integration`), frontend Vitest (`frontend/src/**/__tests__`), frontend Playwright (`frontend/e2e`), and widget tests (`widget/src`, `widget/e2e`). Evidence gathered by reading/grepping test files — no suites executed.

**Classification key:** COVERED = a test asserts the documented behavior (status codes, error codes, side effects) with file + test name cited. PARTIAL = some assertions exist but documented sub-behaviors are missing. MISSING = no test asserts the behavior.

---

## 1. Chatbot Management

### 1.1 List Chatbots (`02-chatbot-crud.md`, scenario 4.3 step 2)
**COVERED**
- `internal/integration/chatbot_test.go::TestChatbot_CRUD` — list-after-create, 200 (integration)
- `internal/integration/organization_constraints_test.go::TestWorkspaceScoping/WSC-002_ListChatbotsScoped` — `X-Workspace-ID` list filtering, WS1 bot not visible in WS2 list (integration)
- `frontend/src/pages/__tests__/ChatbotsPage.test.tsx` — list render, API error handling (frontend unit); `frontend/e2e/chatbot.spec.ts` — "dashboard shows chatbot list" (e2e)

*Missing detail:* empty-array response and non-member-of-workspace → `403` on list (WSC-003 explicitly noted as untestable in test comments — org-level membership means all org members see all workspaces).

### 1.2 Create Chatbot (`02-chatbot-crud.md`, scenarios 4.1/4.2/4.5)
**PARTIAL**
- Covered: `internal/integration/lifecycle_test.go::TestChatbot_Lifecycle` — BOT-001 (201 valid), BOT-003 (optional fields applied), BOT-004 (missing name → 400), BOT-005 (defaults); WSC-001 (workspace_id set from header); `TestChatbot_CRUD` (201); `frontend/src/features/chatbot/components/__tests__/NewChatbotForm.test.tsx` + e2e "can create a new chatbot" / "chatbot creation requires a name" (frontend)
- **Missing:** `ERR_MAX_CHATBOTS_EXCEEDED` → 403 at limit — code exists (`internal/api/handlers/chatbot_list.go` ~line 95) but **no test anywhere** hits it (only plan-config field tests in `internal/repository/plan_repo_test.go` / `internal/models/plan_limits_test.go`) → **P0 (plan/billing enforcement)**
- **Missing:** whitespace-only name → 400 (backend trims in `chatbot_list.go:73` but no test; frontend `frontend/src/domain/chatbot/__tests__/validation.test.ts` covers client-side only); name > 100 chars → 400 (no server-side validation or test found); language default `tr-TR` + locale normalization (`normalizeLocale` in `internal/services/chatbot_service.go:490` untested; `TestNormalizeLangCode` in `internal/services/chat_service_test.go` tests a *different* helper); org-context (`X-Organization-ID`) create + non-member 403 (scenario 4.4)

### 1.3 Get Chatbot by ID
**COVERED (mostly)**
- `TestChatbot_CRUD` — GET 200, 404 after soft-delete (integration); `internal/integration/multi_tenant_test.go::TestMultiTenant_ChatbotIsolation` — cross-user GET → 403/404 (integration); router table `internal/api/router/router_test.go` — 401 unauthenticated (unit)

*Missing:* non-UUID → 400 on the authenticated endpoint (only public-side UUID tests exist: `internal/api/handlers/public_uuid_test.go`).

### 1.4 Full PUT `/chatbots/{id}`
**COVERED**
- `internal/integration/chatbot_domain_test.go::TestChatbot_DomainUpdates` — update + GET-verify of name/desc/instruction/language/theme/temperature/max_tokens (integration)
- Plan gating: `internal/api/handlers/chatbot_unit_test.go::TestUpdateChatbot_AutoRefreshForbidden_ForFreePlan`, `TestUpdateChatbot_SecureEmbedForbidden_ForFreePlan`, `chatbot_handoff_enforcement_test.go::TestUpdateChatbot_HandoffForbidden_ForProPlan`, `chatbot_discovery_enforcement_test.go::TestChatbot_Update_DiscoveryMode_Forbidden_OnZeroCrawlLimit` (unit); `internal/integration/plan_limits_free_test.go::TestFreePlan_Guardrails_Restrictions` (thresholds/smart/topics → 403), `TestFreePlan_Branding_HideBrandingForbidden`, `TestFreePlan_RefreshPolicyAuto_Forbidden`, `TestFreePlan_SecureEmbed_Forbidden`; `internal/integration/plan_features_pro_test.go` — same fields 200 on Pro (integration)

*Missing:* empty body → 400, malformed JSON → 400; `FeatureError` body shape (`upgrade_required`/`feature`) rarely asserted (status code only).

### 1.5 Delete Chatbot
**PARTIAL**
- Covered: `TestChatbot_CRUD` — 204 then GET → 404 (integration); `lifecycle_test.go` deletion tests; `ChatbotsPage.test.tsx` "deletes bot and removes from list" (frontend)
- **Missing:** vector-store cleanup on delete (only admin-side `internal/repository/admin_chatbot_repo_test.go::TestMockAdminChatbotRepo_DeleteVectors` mock test); delete-already-deleted → 404

### 1.6 Domain-Specific PUT Endpoints (`03-domain-specific-put-endpoints.md`)
| Endpoint | Status | Evidence / gaps |
|---|---|---|
| basic-info | COVERED | `TestChatbot_DomainUpdates` (integration). *Gap:* empty name → 400, name trimming untested |
| appearance | PARTIAL | `TestChatbot_DomainUpdates` (theme/position), `internal/api/handlers/chatbot_appearance_test.go::TestUpdateAppearance_NewFields` (bubble radius, new colors, persistence) (unit+testdb); frontend `ColorsSection`/`AppearanceSection` tests; widget `widget/e2e/widget-config.spec.ts` (e2e). **Missing: invalid color → 403 `feature:"color"`** (P1 — documented behavior, zero tests); `SuggestedQuestions` dedupe/cap covered by `internal/services` `TestNormalizeSuggestions` + `internal/rag/topic_extractor_test.go::TestNormalizeSuggestions_DedupeAndCap` (unit); `ManualQuestions` plan limit untested; `CustomBranding` Enterprise gate untested |
| model | COVERED | `internal/integration/chatbot_test.go::TestFreePlan_ModelRestriction_AllowsGpt4oMini` / `TestFreePlan_ModelRestriction_BlocksDisallowedModels` (403 for disallowed) (integration); `internal/integration/temperature_model_test.go::TestModelRestrictions`, `TestTemperatureParameter`, `TestMaxTokensParameter` (integration). *Gap:* max_tokens below-min/above-max → 403 with distinct `upgrade_required` flags not asserted |
| security | COVERED | `internal/integration/chatbot_secure_embed_test.go::TestChatbot_SecureEmbed_UpdateAndGet`, `plan_limits_free_test.go::TestFreePlan_SecureEmbed_Forbidden`, `secure_embed_comprehensive_test.go` (integration); `TestUpdateChatbot_SecureEmbedForbidden_ForFreePlan` (unit); `internal/services/chatbot_service_test.go::TestFlexibleStringSlice_UnmarshalJSON`, `TestSecuritySettingsRequest_UnmarshalJSON` (JSON-array + comma-string + empty + null) |
| guardrails | COVERED (config) | `plan_limits_free_test.go::TestFreePlan_Guardrails_Restrictions` (integration); `internal/services/guardrail_service_test.go::TestGuardrailService_EnforceFallbackMode`, `_DoesNotMutateOriginal`, `_PermissionChecks`, `_InitializeThresholdConfig`, fallback message getters (unit). *Gap:* `topic_restrictions: null` clears existing — untested |
| handoff | COVERED | `chatbot_handoff_enforcement_test.go::TestUpdateChatbot_HandoffForbidden_ForProPlan` (unit); enable-via-PUT exercised in `internal/integration/handoff_test.go::TestHandoff_Flow` (integration) |
| refresh | PARTIAL | `plan_limits_free_test.go::TestFreePlan_RefreshPolicyAuto_Forbidden` (integration), `TestUpdateChatbot_AutoRefreshForbidden_ForFreePlan` (unit). **Missing: scenario 4.9 `NextRefreshAt` transitions** (auto→future date, manual→null, frequency-change-while-manual) — zero tests reference `NextRefreshAt` (P2) |
| scraping | PARTIAL | `TestChatbot_Update_DiscoveryMode_Forbidden_OnZeroCrawlLimit` (unit), `plan_limits_free_test.go::TestFreePlan_DiscoveryMode_Disabled_OnZeroCrawlLimit` (integration). *Gap:* IncludePaths/ExcludePaths normalization + dedupe, SelectorWhitelist normalization untested (P2) |

### 1.7 Suggestions (`04-suggestions.md`, scenario 4.6)
**COVERED**
- Unit: `internal/api/handlers/chatbot_suggestions_test.go` — `TestRegenerateSuggestions_Success/InvalidChatbotID/Unauthorized/MethodNotAllowed/ReturnsJSON`, `TestGetSuggestionJobStatus_Success/NotFound/WithSuggestions/WithError/WithEmptySuggestions/InvalidChatbotID/MethodNotAllowed/ReturnsJSON` (handler)
- Integration: `internal/integration/suggestion_job_test.go::TestSuggestionRegenerationPolling_FullFlow` (202+job_id → poll → completed), `_ConcurrentRequests` (multi-regenerate, no dedup), `_GetLatestJob`, `_JobCompletion`, `_JobFailure`, `_JobWithRunningStatus`
- Frontend: `frontend/src/hooks/mutations/__tests__/useChatbotMutations.test.tsx` (regenerate + poll until completed, job failure, polling timeout); `SuggestionsPanel.test.tsx` (add/remove, 6-cap + dedupe)

*Gap:* "no sources → completes with empty questions" edge not explicitly asserted; regenerate-404 for non-existent chatbot only covered via shared `getChatbotContext` indirectly.

### 1.8 Sitemap Discovery (`05-sitemap-discovery.md`, scenario 4.7)
**COVERED**
- `internal/integration/url_discovery_test.go::TestURLDiscovery_AutoMode`, `_PendingMode` (integration); `internal/api/handlers/sitemap_test.go::TestDiscoverSitemap_Unauthorized/MethodNotAllowed/InvalidPath/NoBody/ResponseFormat` (unit)

*Gap:* invalid-sitemap-URL failure capture (202 with failure in source status) and 0-URL completion not asserted.

### 1.9 Auth/State Edge Cases (`08-edge-cases.md`) & Expected Behaviors (`09-expected-behavior.md`)
**PARTIAL**
- Covered: 401 without/with tampered/expired JWT (`internal/integration/auth_edges_test.go`, `auth_expired_access_test.go`, router_test 401 table — integration+unit); 405 on dispatch (`internal/integration/methods_test.go::TestMethods_NotAllowed`, router_test method table); soft-deleted → 404 (`TestChatbot_CRUD`)
- Missing: `DeletedAccountMiddleware` → 403 (no test found); malformed-JSON → 400 on chatbot PUTs; `null`-field-no-update semantics; reserved path `"new"` → 400 (code exists in `internal/api/handlers/chatbot_context.go:28`, untested); plan-missing/plan-zero-limits/downgrade edge cases; concurrent create-at-plan-limit race (no atomic count+create); concurrent delete+update; DB-failure → 500

---

## 2. Chat Operations

### 2.1 Scenario 1 — Basic Authenticated Chat Flow (`12-test-scenarios.md`)
**PARTIAL**
- Covered: `internal/integration/chat_test.go::TestChat_StubbedContext` (200, non-empty response, tokens>0), `internal/integration/mock_chat_test.go::TestMockChatFlow` (response content + `tokens_used` with mocked RAG), `internal/integration/analytics_feedback_test.go::TestAnalytics_ThumbsUpAfterFeedback` (asserts Conversation row by `session_id` + Message row in DB), `internal/integration/turkish_test.go` (user message content persisted) (integration)
- **Missing:** response-field contract — `messageID`/`conversationID` UUID format, `sources` array shape, `handoffRequestID` field — asserted nowhere in Go tests (struct exists in test but fields unchecked); no explicit assertion that *both* role=user and role=assistant messages are persisted per exchange (P1)

### 2.2 Scenarios 2–3 — Session Continuity / New Session
**PARTIAL (repo-level only)**
- Covered: `internal/repository/conversation_repo_integration_test.go::TestPostgresConversationRepo_GetOrCreateBySessionID` (same session → same conversation; new session → new conversation), `_CreateMessage`, `_GetMessages` (integration w/ DB)
- **Missing:** end-to-end two-message-same-session test asserting same `conversationID`, new `messageID`, 4-message history; **`LastMessageAt` update has zero test references** (P2→P1 for analytics correctness)

### 2.3 Scenarios 4–5 — Public Chat & Non-Public 404
**PARTIAL / doc-mismatch**
- Covered: `internal/integration/public_endpoints_test.go::TestPublicEndpoints_BasicFlow` (public config 200 + public chat 200, no auth) (integration); widget `widget/e2e/chat.spec.ts` + `chat-flows.spec.ts` (real chat UX); CORS: `internal/integration/cors_public_test.go::TestCORS_PublicEndpoints`, `cors_allow_origin_test.go::TestCORS_AllowConfiguredOrigin/_DisallowUnconfiguredOrigin/_MultipleOrigins` (integration)
- **MISSING: `isPublic: false` → 404 information-hiding** — **no `is_public` flag exists in models/migrations at all** (grep-verified); public endpoints serve any valid chatbot ID. Doc-vs-implementation mismatch — spec requires either implementation or doc correction (P1, potential data-exposure concern)

### 2.4 Scenario 6 — Public Config
**COVERED (mostly)**
- `internal/api/handlers/chatbot_appearance_test.go::TestPublicChatbotConfig_NewFields` (unit), `internal/api/handlers/public_suggestions_cache_test.go::TestPublicChatbotConfig_SuggestionsCacheKeyedByUpdatedAt` (cache key), `internal/integration/public_suggestions_test.go::TestPublicChatbotConfig_IncludesSuggestions` (integration), `internal/integration/handoff_test.go::TestHandoff_Widget` (handoff_enabled flag in config), `widget/e2e/widget-config.spec.ts` (theme color, position, welcome message, fonts, colors — e2e)

*Missing:* `Cache-Control` header assertion (zero test references); "no sensitive fields leaked" (embed secret, instructions) not asserted.

### 2.5 Scenario 7 — Chat Feedback
**PARTIAL**
- Covered: `internal/api/handlers/feedback_test.go::TestFeedbackHandler_FirstThumbsUp`, `_ChangeThumbsDownToUp` (overwrite semantics + analytics increment, unit); `internal/integration/analytics_feedback_test.go` / `analytics_full_coverage_test.go` (feedback → analytics counts, integration); `internal/integration/feedback_test.go::TestFeedback_Protected_Unauthorized/_Authorized` (401/auth on protected route); widget `widget/src/components/__tests__/Message.test.tsx` (thumbs up/down clicks), `widget/e2e/chat.spec.ts` (feedback buttons) — **but all backend coverage is on the protected `/api/v1/messages/{id}/feedback` route**
- **Missing:** the documented **public** endpoint `POST /api/v1/public/chatbots/{id}/feedback` (`PublicHandlers.SubmitFeedback`, `internal/api/handlers/public.go:362`) has only `public_uuid_test.go::TestPublicSubmitFeedback_InvalidUUID_Returns400`; no happy-path, no 404-unknown-message, no overwrite test (P1 — the widget uses this route). Also doc says field `score`, impl uses `thumbs_up`; missing/non-boolean score → 400 cannot be distinguished (bool zero-value) and is untested

### 2.6 Scenario 8 — Message Validation
**PARTIAL (mostly gaps)**
- Covered: widget-side char limit + truncation (`widget/e2e/chat.spec.ts` "shows character limit counter", "truncates messages exceeding max length"; "does not send empty messages")
- **Missing (backend):** empty/whitespace message → 400 and missing `session_id` → 400 (implemented in `internal/api/handlers/chat.go:67-69`, untested); **4000-char server-side limit does not exist in code** (grep-verified — doc-vs-impl mismatch); non-string message → 400; unknown-fields-ignored → 200 (P1)

### 2.7 Scenario 9 — Chat Auth & Authz
**COVERED**
- router_test table: `/api/v1/chatbots/x/chat` → 401 unauthenticated (unit); `internal/integration/multi_tenant_test.go::TestMultiTenant_ChatIsolation` (cross-tenant chat blocked, integration); middleware-level expired/tampered JWT 401 (`auth_edges_test.go`, `auth_expired_access_test.go`)

*Missing:* disabled/paused chatbot → 403/423 (no disabled-state test found).

### 2.8 Scenario 10 — Source Citations
**PARTIAL**
- Covered: `TestMockChatFlow` (mocked search result flows into context; response asserted) (integration); RAG tiering unit tests (see 2.10)
- **Missing:** `sources` array non-empty with required fields (title/url/snippet), sources rendered as markdown citations in the response text — untested end-to-end (P1)

### 2.9 Scenario 11 — Markdown Rendering
**PARTIAL**
- Covered: widget `Message.test.tsx` renders markdown (react-markdown mocked); frontend `HandoffRequestsTab.test.tsx` renders markdown messages
- Missing: no backend/e2e assertion of markdown structure (fencing, tables, no double-escaping) in chat responses (P2 — inherently LLM-dependent)

### 2.10 Scenario 12 — RAG Tiered Search (`07-chat-pipeline-rag-integration.md`)
**COVERED (unit)**
- `internal/rag/search_test.go::TestSearchContextTiered_Unit`, `TestSearchContext_ThresholdAndMaxTokens`, `_AllBelowThreshold`, `_ThresholdZero/_ThresholdOne`, `_Separator`, `_MissingQdrant`, `TestSearchContext_TurkishScoring` (unit); `internal/services/chat_context_builder_test.go::TestChatContextBuilder_Build_ThresholdConfig`, `_PlanEnforcedThresholds` (unit)

*Layer flag:* tier thresholds/`RAG_TOPK`/`RAG_MAX_CONTEXT_TOKENS` verified only at unit level with stubbed vector clients; no real-Qdrant tiered test (`internal/integration/qdrant/` requires TEST_INTEGRATION=1); `ConfidenceTier` in chat response unasserted.

### 2.11 Scenario 13 — Fallback Strategies (`08-chat-pipeline-fallback-handling.md`)
**COVERED (mostly)**
- `internal/integration/chat_fallback_test.go::TestChat_Fallback_NoContext` (integration); `internal/integration/chat_empty_state_test.go::TestChat_EmptyState_GreetingResponse/_FactualRefusal/_IdentityQuestion` (integration); `internal/services/chat_service_test.go::TestGetStaticFallbackMessage/TestGetErrorMessage/TestGetHandoffMessage/TestGetEmptyStateMessage/TestBuildRestrictedFallbackPrompt` (unit); `internal/services/guardrail_service_test.go::TestGuardrailService_GetFallbackMessage/_GetStaticFallbackMessage/_GetErrorMessage/_GetHandoffMessage/_GetEmptyStateMessage` (unit); escalate→handoff: `internal/integration/agent_handoff_test.go::TestAutomatedHandoff` (LLM tool-call triggers handoff creation, integration)

*Missing:* "fallback type logged for analytics" not asserted; smart-fallback LLM-general-knowledge path only prompt-unit-tested.

### 2.12 Scenarios 14–15 — Guardrails Runtime (`09-chat-pipeline-guardrails.md`)
**PARTIAL**
- Covered: threshold-config init/plan enforcement (unit, 2.10); low-context → static fallback (integration, 2.11)
- **Missing:** low-confidence response *blocked/flagged* with refusal message + incident logging (no runtime test); **topic-restriction runtime refusal** — restrictions are injected into the system prompt (`internal/services/chat_pipeline.go:108` → `BuildSystemPrompt`) but no test asserts the prompt contains them nor that a restricted-topic query yields a refusal without leakage (P1); guardrail-violation logging untested

### 2.13 Scenario 16 — Circuit Breaker (`07`)
**COVERED (unit only — flagged)**
- `internal/rag/circuit_breaker_test.go` — `TestCircuitBreaker_ClosedState/_OpensOnFailures/_OpenStateRejectsQuickly/_HalfOpenAllowsProbe/_ClosesOnSuccess/_MetricsRecorded`, `TestDefaultCircuitBreakerSettings`, `TestCircuitBreakerManager_RegisterAndGet/_GetAllStates/_GetAllProviderStatus` (unit, thorough)

*Layer flag:* no integration test of breaker-open behavior at the chat endpoint; note impl degrades gracefully to a 200 fallback message (`internal/integration/chat_error_test.go::TestChat_OpenAIError_GracefulMessage`, `chat_timeout_test.go::TestChat_OpenAIEmbeddingTimeout_Fallback`, `chat_qdrant_error_test.go::TestChat_QdrantSearchError_Fallback`, `chat_embed_error_test.go::TestChat_EmbeddingError_Fallback`) — docs' "503 on breaker open" doesn't match implemented behavior (P2 doc fix).

### 2.14 Scenario 32 — Rate Limiting
**COVERED (mostly)**
- `internal/integration/rate_limit_test.go::TestRateLimit_ChatEndpoint`, `ratelimit_test.go::TestRateLimit_HeadersPresent/TestRateLimit_RetryAfterOnBlock/TestRateLimit_Recovery/TestRateLimit_PerUserIsolationExtended`, `ratelimit_ip_test.go::TestRateLimit_IPIsolationOnHealth`, `ratelimit_plan_test.go`, `ratelimit_isolation_test.go` (integration)

*Missing:* per-**session** limiting (limiter is per-user/IP — doc mismatch); handoff endpoint 429 not specifically tested.

### 2.15 Scenarios 33–34 — Performance / Concurrent Sessions
**MISSING** — no latency (p50/p95/p99) tests exist anywhere; no concurrent-sessions-isolation chat test (nearest: `internal/integration/quota_enforcement_test.go::TestQuota_RaceCondition_DoubleSpend`, `TestSuggestionRegeneration_ConcurrentRequests`) (P2)

### 2.16 Edge Cases (`13-edge-cases.md`)
- Covered: Unicode/Turkish (`internal/integration/turkish_test.go` backend; widget e2e Unicode/special-chars/long/multiline); embedding/Qdrant/OpenAI failure fallbacks (integration, 2.13); large-context truncation (`TestSearchContext_ThresholdAndMaxTokens`); XSS/session persistence client-side (`widget/src/__tests__/xss-prevention.test.ts`, `session.test.ts`); history pruning (`internal/services/chat_service_test.go::TestCalculateHistoryLimit`, unit)
- **Missing:** 10K-char sessionID; SQL-injection-in-sessionID sanitization; 100+-message conversation e2e; empty-LLM-content response; simultaneous-feedback race; DB-failure-during-chat → 500; cross-feature action+handoff priority (P2 mostly)

### 2.17 Prompts (`10-prompts.md`) & Message Model (`06-message-model.md`)
**PARTIAL** — `internal/integration/chat_prompt_default_tr_test.go::TestChat_DefaultTurkishPrompt` (integration), `chat_context_builder_test.go` (unit) cover prompt assembly/language directive; "all prompts English" not asserted. Message model fields covered via `conversation_repo_integration_test.go` except `SourcesJSON`/`ToolCallsJSON` persistence assertions.

---

## 3. Human Handoff

> **Major doc-vs-impl mismatches found:** (1) impl status model is `pending → assigned → resolved` (`internal/models/handoff.go:23,34`), docs say `pending → contacted → resolved`; (2) POST handoff returns **200**, docs say 201; (3) PATCH returns **204**, docs say 200; (4) duplicate handoff returns **409**, docs say "returns existing requestId"; (5) docs say disabled-chatbot handoff → 404, impl returns 400 (asserted in test).

### 3.1 Scenario 26 — Handoff Request Flow
**PARTIAL**
- Covered: `internal/integration/handoff_test.go::TestHandoff_Flow` (public POST → 200, request listed), `internal/integration/agent_handoff_test.go::TestAutomatedHandoff` (tool-call path), widget `widget/src/__tests__/edge-cases.test.ts` "handles successful response with handoff" (`handoff_request_id` parsing)
- **Missing:** `role=system` message recording the handoff event (unasserted); subsequent-chat-responses-include-`handoffRequestID` not asserted end-to-end (P1)

### 3.2 Scenario 27 — Submit Contact Email ⚠️
**MISSING — and the endpoint is broken**
- `PublicSubmitEmail` (`internal/api/handlers/handoff.go:190`) has **zero tests** (grep-verified across all Go test dirs). Worse, by inspection: it validates the email via `mail.ParseAddress` but **never persists it** — calls `UpdateHandoffRequestStatus(ctx, requestID, "pending", nil)`, dropping the email and *resetting* status to pending; no 404 for unknown requestId; no 409 for already-resolved; `ContactedAt` never set. Every doc assertion (200, status=contacted, ContactEmail stored, 400/404/409) is unimplemented AND untested.
- Only widget UI coverage exists: `widget/src/components/__tests__/Message.test.tsx` "renders handoff card and handles email submission" (component-level; verifies `onSubmitEmail` callback, not the API).
- **P0: user contact data is silently lost in production; tests would have caught it.**

### 3.3 Scenario 28 — Management List
**PARTIAL**
- Covered: `TestHandoff_Flow` (list 200 non-empty), `internal/integration/multi_tenant_test.go::TestMultiTenant_HandoffIsolation` (cross-tenant), router_test 401 for `/api/v1/chatbots/x/handoff-requests` (integration+unit); `frontend/src/features/chatbot/pages/tabs/__tests__/HandoffRequestsTab.test.tsx` (list render, detail, upgrade message)
- **Missing:** **pagination (page/limit) and status filtering are not implemented in `ListHandoffRequests`** (`internal/api/handlers/handoff.go:120-134` ignores query params) — doc mismatch; sort order unasserted (P1 doc/impl reconciliation)

### 3.4 Scenario 29 — Get Detail
**COVERED (mostly)** — `internal/integration/handoff_test.go::TestHandoff_RequestDetail_NotFoundReturns404` (integration); `HandoffRequestsTab.test.tsx` "opens detail and renders markdown messages" (frontend). *Gap:* full-field contract assertion (ID, ChatbotID, ConversationID, MessageID, Status, ContactEmail, ContactedAt, ResolvedAt, Notes).

### 3.5 Scenario 30 — Status Transitions
**PARTIAL**
- Covered: `internal/integration/handoff_test.go::TestHandoff_Status_Lifecycle` (pending → assigned → resolved, verified via list), `TestHandoff_UpdateStatus_NotFoundReturns404`, `TestHandoff_DuplicateRequest` (409 duplicate; new request allowed after resolution) (integration)
- **Missing:** invalid-transition rejection (`pending → resolved` → 422, `resolved → contacted` → 422), invalid status value → 400, `ResolvedAt` timestamp assertion, notes persistence — none tested (P1). Auth 401 via router table; wrong-org via `TestMultiTenant_HandoffIsolation`

### 3.6 Scenario 31 — Full Lifecycle
**PARTIAL** — composite of `TestHandoff_Flow` + `TestHandoff_Status_Lifecycle` + `internal/integration/handoff_test.go::TestHandoff_Analytics` (handoff counted in analytics), but the contact-email middle step is missing (see 3.2), and timeline timestamps (`ContactedAt`/`ResolvedAt`) are never asserted.

### 3.7 Handoff Edge Cases (`06-edge-cases.md`)
- Covered: duplicate → 409 (`TestHandoff_DuplicateRequest`); handoff-disabled → 400 (`TestHandoff_EdgeCases`); misconfigured email → 200 partial success (`TestHandoff_EdgeCases`); request-without-prior-conversation (implicitly via `TestHandoff_Status_Lifecycle` creating request directly with session_id)
- **Missing:** invalid sessionID → 400; exotic/injection emails sanitization; concurrent PATCH; long-notes validation; same-status PATCH; deleted-conversation orphan handling; authenticated-vs-public UserID population; 10K+ list scale (P2)

### 3.8 Handoff Config on Chatbot (`chatbot-management/03` Handoff PUT)
**COVERED** — see §1.6 (handoff row): plan-gate unit test (`TestUpdateChatbot_HandoffForbidden_ForProPlan`) + integration enable-flow (`TestHandoff_Flow`).

<!-- APPEND -->
