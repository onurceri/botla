> Extracted from: chatbot-management.md — Section 2.2

# 2.2 Domain-Specific PUT Endpoints

## Basic Info — `PUT /api/v1/chatbots/{id}/basic-info`
- [ ] Updates `Name`, `Description`, `CustomInstruction`, `Language`
- [ ] `name` is required; empty string returns `400 Bad Request`
- [ ] `name` trimmed of leading/trailing whitespace
- [ ] `language` normalized via `normalizeLocale`
- [ ] `200 OK` with updated chatbot on success
- [ ] Plan validation run on all changed fields

## Appearance — `PUT /api/v1/chatbots/{id}/appearance`
- [ ] Updates: `ThemeColor`, `WelcomeMessage`, `Position`, `BotMessageColor`, `UserMessageColor`, `BotMessageTextColor`, `UserMessageTextColor`, `ChatFontFamily`, `ChatHeaderColor`, `ChatHeaderTextColor`, `ChatBackgroundColor`, `BubbleRadius`, `InputBackgroundColor`, `InputTextColor`, `SendButtonColor`, `BotIcon`, `BotDisplayName`
- [ ] Also updates branding/suggestions sub-fields: `HideBranding`, `CustomBranding`, `SuggestedQuestions`, `ManualQuestions`, `SuggestionsEnabled`
- [ ] Color fields validated: must be HEX (`#RRGGBB`, `#RGB`, `#RRGGBBAA`) or RGBA (`rgba(...)`)
- [ ] Invalid color format returns `403` with feature error `"color"`, `upgrade_required: false`
- [ ] `HideBranding` validated against plan — gated plans return `upgrade_required: true`
- [ ] `CustomBranding` requires Enterprise plan
- [ ] `Position` accepts: `bottom-right`, `bottom-left` (no server-side enum validation currently, but frontend constraints apply)
- [ ] `BubbleRadius` stored as CSS value string (e.g. `"22px"`)
- [ ] `SuggestedQuestions` deduplicated and truncated to 6 items of max 120 chars each
- [ ] `ManualQuestions` validated against plan `ChatMaxManualQuestions` limit
- [ ] `200 OK` with updated chatbot on success

## Model — `PUT /api/v1/chatbots/{id}/model`
- [ ] Updates: `Model`, `Temperature`, `MaxTokens`, `CustomInstruction`
- [ ] `model` validated against plan's `ChatAllowedModels` list
- [ ] Disallowed model returns `403` with `"model_selection"` feature error
- [ ] `temperature` range: min 0, max 2 (validated by float64 pointer)
- [ ] `max_tokens` validated against plan's `ChatMinResponseTokenLimit` / `ChatMaxResponseTokenLimit`
- [ ] Below min returns `403` with `upgrade_required: false`
- [ ] Above max returns `403` with `upgrade_required: true`
- [ ] `200 OK` with updated chatbot on success

## Security — `PUT /api/v1/chatbots/{id}/security`
- [ ] Updates: `SecureEmbedEnabled`, `AllowedDomains`, `EmbedSecret`
- [ ] `SecureEmbedEnabled` validated against plan — gated plans return `upgrade_required: true`
- [ ] `AllowedDomains` accepts both JSON array `["example.com"]` and comma-separated string `"example.com,test.com"` (via `FlexibleStringSlice`)
- [ ] Setting `AllowedDomains` without `SecureEmbedEnabled` feature returns feature error
- [ ] `EmbedSecret` stored as-is (plaintext — should encode/tokenize in practice)
- [ ] `200 OK` with updated chatbot on success

## Guardrails — `PUT /api/v1/chatbots/{id}/guardrails`
- [ ] Updates: `ConfidenceThreshold`, `ThresholdConfig`, `FallbackMessages`, `TopicRestrictions`
- [ ] `ThresholdConfig` contains: `high_threshold`, `medium_threshold`, `fallback_mode`, `show_confidence_warning`
- [ ] `fallback_mode` values: `"smart"` (requires `GuardrailsCanUseSmartFallback`), `"static"` (always allowed), `"escalate"` (requires `GuardrailsCanUseEscalateFallback`)
- [ ] Threshold customization (`high_threshold`/`medium_threshold` non-zero or `show_confidence_warning`) requires `GuardrailsCanCustomizeThresholds`
- [ ] `TopicRestrictions` managed via `TopicConfig`: `allowed_topics`, `blocked_topics`, `blocked_message`
- [ ] Setting `TopicRestrictions` requires `GuardrailsCanManageTopics`
- [ ] Setting `TopicRestrictions` to `null` in request clears existing topic restrictions
- [ ] `FallbackMessages` object: `no_info_found`, `error_message`, `handoff_message`
- [ ] `200 OK` with updated chatbot on success

## Handoff — `PUT /api/v1/chatbots/{id}/handoff`
- [ ] Updates: `HandoffEnabled`, `HandoffType`, `HandoffConfig`
- [ ] `HandoffEnabled: true` validates against `GuardrailsCanUseEscalateFallback` plan limit
- [ ] `HandoffType`: `"email"` is currently supported
- [ ] `HandoffConfig`: `email_to`, `email_subject`
- [ ] `200 OK` with updated chatbot on success

## Refresh — `PUT /api/v1/chatbots/{id}/refresh`
- [ ] Updates: `RefreshPolicy`, `RefreshFrequency`
- [ ] `RefreshPolicy`: `"manual"` or `"auto"`
- [ ] `"auto"` requires `RefreshEnabled` plan limit — returns feature error if not allowed
- [ ] When set to `"auto"`, `NextRefreshAt` calculated based on frequency and current time
- [ ] When set to `"manual"`, `NextRefreshAt` set to `nil`
- [ ] `RefreshFrequency`: `"daily"`, `"weekly"`, `"monthly"` (for auto policy)
- [ ] `200 OK` with updated chatbot on success

## Scraping — `PUT /api/v1/chatbots/{id}/scraping`
- [ ] Updates: `DiscoveryMode`, `IncludePaths`, `ExcludePaths`, `SelectorWhitelist`
- [ ] `DiscoveryMode`: `"auto"`, `"pending"`, `"disabled"`
- [ ] Non-disabled mode requires `ScrapingMaxPagesPerCrawl > 0` in plan limits
- [ ] `IncludePaths`/`ExcludePaths`: normalized (leading `/` added), deduplicated
- [ ] `SelectorWhitelist`: CSS selectors, whitespace-normalized, deduplicated
- [ ] `200 OK` with updated chatbot on success
