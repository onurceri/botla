> Extracted from: chat-operations.md — Section: 1.1 Authenticated Chat

## 1.1 Authenticated Chat

- [ ] POST /api/v1/chatbots/{id}/chat accepts a valid authenticated request and returns a streamed or JSON response
- [ ] Request body supports `message` (string, max 4000 chars) and `sessionID` (string)
- [ ] A new conversation is created when `sessionID` is not previously seen
- [ ] An existing conversation is resumed when `sessionID` matches a prior session
- [ ] Response includes `messageID` (UUID), `conversationID` (UUID), `response` (markdown string)
- [ ] Response includes `sources` array of `SourceCitation` objects when sources are found
- [ ] Response includes `tokensUsed` (int) with the total token count
- [ ] Response includes `handoffRequestID` (*string, null if no handoff is active)
- [ ] Messages exceeding 4000 chars are rejected with a 422 or 400 status
- [ ] Missing `message` field returns a 400 validation error
- [ ] Invalid chatbot ID returns 404
- [ ] Chatbot not belonging to user's org returns 403
- [ ] Disabled/paused chatbot returns 403 or 423
- [ ] Rate limiting is applied per-session and per-IP
