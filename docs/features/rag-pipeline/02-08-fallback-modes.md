> Extracted from: rag-pipeline.md — Section: 2.8 Fallback Modes

## 2.8 Fallback Modes

- [ ] **Static fallback:** Returns generic "I don't know" message (configurable)
- [ ] **Smart fallback:** AI generates a redirect response ("I can't answer that, but I can help with X")
- [ ] **Escalate fallback:** Offers human handoff ("Would you like to speak with a human?")
- [ ] Fallback mode is configurable per chatbot
- [ ] Fallback message is rendered in the UI (not an error)
- [ ] Fallback is triggered when all search results are in the low tier (< 0.30)
