> Extracted from: chat-operations.md — Section: 1.8 Chat Pipeline — Fallback Handling

## 1.8 Chat Pipeline — Fallback Handling

- [ ] No relevant chunks found → fallback strategy is invoked
- [ ] **Static fallback**: returns a configured default message
- [ ] **Smart fallback**: LLM generates a response from general knowledge (if allowed)
- [ ] **Escalate fallback**: returns a handoff prompt/suggestion
- [ ] Fallback behavior is configurable per chatbot
- [ ] Fallback type is logged for analytics
