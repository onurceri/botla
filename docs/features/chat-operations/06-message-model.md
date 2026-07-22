> Extracted from: chat-operations.md — Section: 1.6 Message Model

## 1.6 Message Model

- [ ] Messages store `ID` (UUID), `ConversationID`, `Role` (user/assistant/system)
- [ ] Content is stored as markdown in the `Content` field
- [ ] `Tokens` stores the token count for the message
- [ ] `Feedback` stores thumbs up/down (boolean, nullable)
- [ ] `SourcesJSON` stores source citations as JSON
- [ ] `ToolCallsJSON` stores action/tool invocation data as JSON
- [ ] `CreatedAt` stores the timestamp
