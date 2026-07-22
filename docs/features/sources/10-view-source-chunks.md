> Extracted from: sources.md — Section: 2.9 Source Management — View Source Chunks

# 2.9 Source Management — View Source Chunks

- [ ] `GET /api/v1/sources/{id}/chunks` returns paginated chunks from Qdrant
- [ ] Supports `offset` query parameter for cursor-based pagination
- [ ] Default page size is 20 chunks
- [ ] Response includes `chunks` array and `next_cursor` for pagination
- [ ] Each chunk includes: id, payload (chatbot_id, source_id, chunk_index, original_text, source_type, created_at)
- [ ] Cross-tenant access returns 404
- [ ] Deleted source returns 404
