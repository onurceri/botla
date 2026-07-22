> Extracted from: human-handoff.md — Section: 1.1 Human Handoff — Public Endpoints

## 1.1 Human Handoff — Public Endpoints

- [ ] POST /api/v1/public/chatbots/{id}/handoff creates a handoff request
- [ ] Request body includes the conversation context/sessionID
- [ ] Response returns `handoffRequestID` (UUID)
- [ ] A Message with role=system is created recording the handoff event
- [ ] The chat response includes `handoffRequestID` in subsequent messages
- [ ] POST /api/v1/public/chatbots/{id}/handoff/{requestId}/contact saves the user's contact email
- [ ] Request body includes `email` (valid email string)
- [ ] Invalid email returns 400
- [ ] Unknown requestId returns 404
- [ ] Already-resolved requests return 409 (conflict)
- [ ] Contact submission updates status to `contacted`
