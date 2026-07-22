> Extracted from: actions-smart-tools.md — Section: 4. Edge Cases

# 4. Edge Cases

## Actions

- [ ] **Action name with special characters**: Validated and sanitized
- [ ] **Action endpoint with no protocol**: Rejected (e.g., `api.weather.com` without `https://`)
- [ ] **Action endpoint with file:// protocol**: Rejected (SSRF)
- [ ] **Action headers containing sensitive values**: Stored encrypted or masked in logs
- [ ] **Action body with invalid JSON**: Rejected at creation
- [ ] **Action execution timeout (slow endpoint)**: Terminated after timeout, error returned to LLM
- [ ] **Action execution returns very large response (>1MB)**: Truncated or rejected
- [ ] **Action parameters don't match LLM arguments**: Graceful error, LLM retries or responds without
- [ ] **Maximum number of actions per chatbot exceeded**: Configurable limit enforced
- [ ] **Action with same name as another (within chatbot)**: Rejected (duplicate name)
- [ ] **Delete action that is currently executing**: Graceful handling, in-flight execution completes or is abandoned
- [ ] **Update action endpoint while chat is in progress**: No impact on in-flight requests, new requests use new config
- [ ] **Action execution during circuit breaker open state**: Queued or rejected with 503

## Cross-Feature Edge Cases

- [ ] **Action execution triggers handoff**: Action result triggers escalation logic → handoff request created
- [ ] **Handoff during action execution**: Handoff takes priority, action may be cancelled
- [ ] **RAG content contains PII/confidential data**: Guardrails or filters prevent exposure
- [ ] **Chat message that triggers both action and handoff**: Both processed, action result used in final response before handoff
- [ ] **Rate limit reached during handoff request**: 429, handoff not created
- [ ] **All external services down (Qdrant, OpenAI, Postgres)**: Graceful degradation, clear error messages
