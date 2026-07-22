> Extracted from: plans-entitlements.md — Section: 1.1 Plan Management

## 1.1 Plan Management

- [ ] `PlanLimits` struct holds 72+ typed fields (replaces old JSONB `PlanConfig`), validated at startup
- [ ] Free plan defaults (`DefaultPlanLimits()`): 1 chatbot, 50 monthly ingestions, 250K embedding tokens, GPT-4o-mini only, 100K monthly chat tokens, 3 RAG top-K, no auto-refresh, no custom branding, 100 RPM global / 30 RPM chat / 10 RPM sources
- [ ] Pro plan defaults (`ProPlanLimits()`): 10 chatbots, 500 ingestions, 2.5M embedding tokens, GPT-4o + GPT-4o-mini, 1M chat tokens, 5 RAG top-K, auto-refresh enabled (5/mo), secure embed, custom thresholds, topic management, customizable guardrail messages, 500 RPM global / 100 RPM chat / 30 RPM sources
- [ ] Ultra plan defaults (`UltraPlanLimits()`): 100 chatbots, 10K ingestions, 100M embedding tokens, GPT-4o-mini + GPT-4o + GPT-5, 5M chat tokens, 10 RAG top-K, 8K response limit, 100/mo refresh, escalade fallback, custom branding, 2000 RPM global / 500 RPM chat / 100 RPM sources
- [ ] Plan validation at startup (`ValidateAllPlans()`) — fails boot if any plan limits are invalid
- [ ] Redis-cached plan lookup with 24-hour TTL (`planCacheTTL`)
- [ ] In-memory `sync.Map` fallback when Redis unavailable
- [ ] Cache invalidation: single-plan (`InvalidateCache`) and all-plans (`InvalidateAllCache`)
- [ ] Model resolution by plan via `model_registry.go` — `GetAvailableModels` filters `ai_models` table by `ChatAllowedModels`
- [ ] API model ID resolution: `GetAPIModelID` maps bare model name to OpenRouter API model ID
- [ ] Prompt language always English (output language controlled by directive, `chat_prompts.go:10`)

