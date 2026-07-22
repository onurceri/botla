> Extracted from: chat-operations.md — Section: 4. Edge Cases

# 4. Edge Cases

## Chat Operations

- [ ] **Empty message**: Returns 400, system does not crash
- [ ] **Message with only whitespace**: Should be treated as empty or trimmed, returns 400
- [ ] **Message at max length (4000 chars)**: Processed correctly, not truncated
- [ ] **Unicode/emoji in message**: Properly encoded and handled
- [ ] **Extremely long sessionID (e.g., 10K chars)**: Truncated or validated, no crash
- [ ] **SessionID with special characters / SQL injection attempts**: Properly sanitized
- [ ] **SessionID reuse across different chatbots**: Creates separate conversations per chatbot
- [ ] **Conversation with 100+ messages**: Context window management works, oldest messages pruned/summarized
- [ ] **Empty response from LLM**: Gracefully handled, returns fallback or error message
- [ ] **LLM returns non-markdown**: Stored as-is, no processing crash
- [ ] **Very large source context (over context limit)**: Truncated to fit `RAG_MAX_CONTEXT_TOKENS`
- [ ] **Feedback on non-existent message**: 404, not created
- [ ] **Duplicate feedback submission**: Overwrites, does not create duplicate records
- [ ] **Public config for deleted chatbot**: 404
- [ ] **Simultaneous feedback requests for same message**: Race condition handled (last write wins or transaction)
- [ ] **Database connection failure during chat**: Returns 500, error logged
- [ ] **Embedding service unavailable**: Circuit breaker triggers, error returned

## Cross-Feature Edge Cases

- [ ] **Action execution triggers handoff**: Action result triggers escalation logic → handoff request created
- [ ] **Handoff during action execution**: Handoff takes priority, action may be cancelled
- [ ] **RAG content contains PII/confidential data**: Guardrails or filters prevent exposure
- [ ] **Chat message that triggers both action and handoff**: Both processed, action result used in final response before handoff
- [ ] **Rate limit reached during handoff request**: 429, handoff not created
- [ ] **All external services down (Qdrant, OpenAI, Postgres)**: Graceful degradation, clear error messages
