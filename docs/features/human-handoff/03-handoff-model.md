> Extracted from: human-handoff.md — Section: 1.3 Handoff Model

## 1.3 Handoff Model

- [ ] Model fields: ID, ChatbotID, ConversationID, MessageID, Status, ContactEmail, ContactedAt, ResolvedAt, Notes
- [ ] Status enum: `pending`, `contacted`, `resolved`
- [ ] `ContactedAt` is set when status changes to `contacted`
- [ ] `ResolvedAt` is set when status changes to `resolved`
- [ ] Handoff requests are linked to a specific message in the conversation
