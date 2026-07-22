> Extracted from: human-handoff.md — Section: 1.2 Human Handoff — Authenticated Management Endpoints

## 1.2 Human Handoff — Authenticated Management Endpoints

- [ ] GET /api/v1/chatbots/{id}/handoff-requests lists all handoff requests for a chatbot
- [ ] List supports pagination (page, limit query params)
- [ ] List supports filtering by status
- [ ] GET /api/v1/chatbots/{id}/handoff-requests/{requestId} returns full handoff detail
- [ ] Detail includes: ID, ChatbotID, ConversationID, MessageID, Status, ContactEmail, ContactedAt, ResolvedAt, Notes
- [ ] PATCH /api/v1/chatbots/{id}/handoff-requests/{requestId} updates status
- [ ] Status transitions: pending → contacted, contacted → resolved
- [ ] Body can include `notes` (string) for agent notes
- [ ] Invalid status transitions are rejected (e.g., pending → resolved without contacting)
- [ ] Authorization: only users with access to the chatbot's org can manage requests
