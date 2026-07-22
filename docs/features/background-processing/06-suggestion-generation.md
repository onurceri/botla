> Extracted from: background-processing.md — Section: 2.5 Background Processing — Suggestion Generation

# 2.5 Background Processing — Suggestion Generation

- [ ] During ingestion, `ExtractIngestionMetadata` is called for each source
- [ ] Metadata includes: `capability_summary` (string) and `suggested_questions` (string array)
- [ ] After source completion, `ReAggregateSuggestionsForChatbot` is triggered
- [ ] Suggestions are aggregated with weighted selection (sources with more chunks have priority)
- [ ] Existing chatbot suggestions are preserved and deduplicated
- [ ] Max 6 suggestions are kept (configurable via plan limit `ChatMaxSuggestedQuestions`)
- [ ] Questions are normalized (trimmed, truncated to 120 chars)
- [ ] Suggestion re-aggregation runs in a background goroutine with panic recovery
