> Extracted from: chat-operations.md — Section: 1.5 Session & Conversation Management

## 1.5 Session & Conversation Management

- [ ] Conversations are identified by `sessionID` provided by the client
- [ ] A session maps one-to-one with a Conversation (ID, ChatbotID, SessionID, UserID, StartedAt, LastMessageAt)
- [ ] `LastMessageAt` is updated on every message in the conversation
- [ ] Public conversations have `UserID = null`
- [ ] Authenticated conversations are linked to the requesting user
- [ ] Conversation history is loaded and included in the RAG context window
- [ ] Old messages beyond the token limit are pruned via summarization or truncation
- [ ] Session timeout / inactivity is handled (configurable)
