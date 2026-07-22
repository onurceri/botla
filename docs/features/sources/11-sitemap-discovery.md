> Extracted from: sources.md — Section: 2.10 Source Management — Sitemap Discovery

# 2.10 Source Management — Sitemap Discovery

- [ ] `POST /api/v1/chatbots/{id}/sitemap/discover` accepts `sitemap_url` in JSON body
- [ ] Sitemap URL is validated before parsing
- [ ] XML sitemap is parsed and URLs are extracted
- [ ] Chatbot's `include_paths` and `exclude_paths` filters are applied
- [ ] Duplicate URLs are deduplicated
- [ ] Response includes `urls` array and `total_count`
- [ ] Invalid sitemap URL returns 400 Bad Request
- [ ] Unparseable sitemap XML returns 400 Bad Request with details
