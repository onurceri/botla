> Extracted from: plans-entitlements.md — Section: 1.3 PlanLimits Categories (from `internal/models/plan_limits.go`)

## 1.3 PlanLimits Categories (from `internal/models/plan_limits.go`)

**Chatbots**
- [ ] `MaxChatbots` — maximum chatbots allowed

**Scraping**
- [ ] `ScrapingDynamicEnabled` — dynamic JS rendering enabled
- [ ] `ScrapingMaxURLsPerBot` — max URLs per chatbot
- [ ] `ScrapingMaxPagesPerCrawl` — max sub-pages per crawl

**Files**
- [ ] `FilesMaxSizeMB` — max file size in MB
- [ ] `FilesMaxFilesPerBot` — max files per chatbot
- [ ] `FilesMaxFilesTotal` — total files across all chatbots
- [ ] `FilesTotalStorageMB` — total storage quota in MB
- [ ] `FilesMaxTextLength` — max text length for text sources

**Chat**
- [ ] `ChatDefaultModel` — default LLM model ID
- [ ] `ChatAllowedModels` — list of allowed model IDs
- [ ] `ChatMaxMonthlyTokens` — monthly token limit
- [ ] `ChatRAGTopK` — RAG context retrieval count
- [ ] `ChatRAGMaxContextTokens` — max context window tokens
- [ ] `ChatMaxSuggestedQuestions` — max auto-suggested questions
- [ ] `ChatMaxManualQuestions` — max manual questions
- [ ] `ChatMinResponseTokenLimit` — minimum max_tokens value
- [ ] `ChatMaxResponseTokenLimit` — maximum max_tokens value

**Refresh**
- [ ] `RefreshEnabled` — auto-refresh sources enabled
- [ ] `RefreshMaxMonthly` — max monthly refreshes

**Security**
- [ ] `SecuritySecureEmbedEnabled` — secure embed (domain whitelist)

**Guardrails**
- [ ] `GuardrailsCanCustomizeThresholds` — adjust sensitivity thresholds
- [ ] `GuardrailsCanUseSmartFallback` — AI-powered fallback responses
- [ ] `GuardrailsCanUseEscalateFallback` — escalate-to-human
- [ ] `GuardrailsCanManageTopics` — topic whitelist/blacklist
- [ ] `GuardrailsCanCustomizeMessages` — edit fallback messages

**Branding**
- [ ] `BrandingCanHideBranding` — hide "Powered by Botla"
- [ ] `BrandingCanCustomBranding` — custom brand/logo

**Rate Limits**
- [ ] `RateLimitsRequestsPerMinute` — global API rate limit
- [ ] `RateLimitsWindowSeconds` — global rate limit window
- [ ] `RateLimitsChatRPM` — chat endpoint rate limit
- [ ] `RateLimitsChatWindow` — chat window seconds
- [ ] `RateLimitsSourcesRPM` — sources endpoint rate limit
- [ ] `RateLimitsSourcesWindow` — sources window seconds

