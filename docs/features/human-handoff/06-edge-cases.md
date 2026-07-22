> Extracted from: human-handoff.md — Section: 4. Edge Cases

# 4. Edge Cases

## Human Handoff

- [ ] **Handoff request without prior conversation**: Creates conversation if needed
- [ ] **Handoff request with invalid sessionID**: 400
- [ ] **Multiple handoff requests for same session**: Second request returns existing requestId (idempotent) or updates it
- [ ] **Contact email with exotic characters (RFC-compliant)**: Accepted and stored
- [ ] **Contact email with newlines or injection attempts**: Sanitized
- [ ] **Handoff request for disabled chatbot**: Returns 404
- [ ] **Concurrent PATCH on same handoff request**: Last write wins or transaction isolation
- [ ] **Handoff request linked to deleted conversation**: Orphaned gracefully, still accessible via management
- [ ] **Extremely long notes field**: Truncated or validated (max length)
- [ ] **PATCH with no status change (same status)**: Accepted and ignored or returns 200 with current state
- [ ] **Handoff request list for chatbot with 10K+ requests**: Pagination works, no timeout
- [ ] **User submits contact before handoff request is created**: Returns 404 (requestId doesn't exist)
- [ ] **Agent updates handoff to contacted but email is empty**: Allowed (contact might be via other channel)
- [ ] **Handoff request from authenticated vs. public chat**: Both work, UserID may be populated for authenticated

## Cross-Feature Edge Cases

- [ ] **Action execution triggers handoff**: Action result triggers escalation logic → handoff request created
- [ ] **Handoff during action execution**: Handoff takes priority, action may be cancelled
- [ ] **RAG content contains PII/confidential data**: Guardrails or filters prevent exposure
- [ ] **Chat message that triggers both action and handoff**: Both processed, action result used in final response before handoff
- [ ] **Rate limit reached during handoff request**: 429, handoff not created
- [ ] **All external services down (Qdrant, OpenAI, Postgres)**: Graceful degradation, clear error messages
