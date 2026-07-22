> Extracted from: sources.md — Section: 2.1 Source Management — List Sources

# 2.1 Source Management — List Sources

- [ ] `GET /api/v1/chatbots/{id}/sources` returns all sources for a chatbot
- [ ] Response includes all DataSource fields: id, chatbot_id, source_type, source_url, file_path, original_filename, status, error_message, chunk_count, processed_at, created_at, hash, size_bytes, last_refreshed_at, is_discovered, capability_summary
- [ ] Sources are scoped to the chatbot — only sources belonging to the authenticated user's chatbot are returned
- [ ] Soft-deleted sources (deleted_at IS NOT NULL) are excluded from results
- [ ] Empty list returns `[]` (HTTP 200) when no sources exist
- [ ] Unauthenticated requests return 401
- [ ] Cross-tenant access returns 404 (source belongs to another user's chatbot)
