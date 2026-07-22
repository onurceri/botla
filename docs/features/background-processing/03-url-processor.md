> Extracted from: background-processing.md — Section: 2.2 Background Processing — URL Processor

# 2.2 Background Processing — URL Processor

- [ ] Fetches raw HTML for link discovery (sub-page crawling)
- [ ] Extracts text content via `ScrapeURLWithFallback` (static colly + dynamic go-rod)
- [ ] CSS selector whitelist from chatbot config is applied when present
- [ ] Dynamic JS rendering is used when plan allows (`ScrapingDynamicEnabled`)
- [ ] Content is normalized (Turkish character normalization via `text.NormalizeTR`)
- [ ] SHA-256 content hash is computed for change detection
- [ ] On refresh: unchanged content skips re-embedding (returns `Skipped: true`)
- [ ] On refresh: changed content deletes old vectors before re-embedding
- [ ] Sub-page discovery modes: `auto` (auto-create sources), `pending` (add to pending table), `disabled` (no discovery)
- [ ] Discovery is limited to 1-level depth (discovered sources don't crawl further)
- [ ] Path filters (include/exclude) from chatbot config are applied during discovery
- [ ] Plan limits (max URLs per bot, max pages per crawl) are enforced during discovery
- [ ] Ingestion metadata is extracted: capability summary, suggested questions
- [ ] Usage statistics are incremented (successful ingestion count, embedding tokens)
