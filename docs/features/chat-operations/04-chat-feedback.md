> Extracted from: chat-operations.md — Section: 1.4 Chat Feedback

## 1.4 Chat Feedback

- [ ] POST /api/v1/public/chatbots/{id}/feedback accepts `messageID` and `score` (boolean — thumbs up/down)
- [ ] Valid feedback updates the `feedback` field on the Message model
- [ ] Duplicate feedback for the same `messageID` overwrites the previous value
- [ ] Invalid `messageID` returns 404
- [ ] Missing or non-boolean `score` returns 400
- [ ] Feedback is stored in the database and visible in analytics
