> Extracted from: chat-operations.md — Section: 1.2 Public Chat

## 1.2 Public Chat

- [ ] POST /api/v1/public/chatbots/{id}/chat accepts requests without authentication
- [ ] Session management works identically to authenticated chat (create/resume)
- [ ] CORS headers are set appropriately for cross-origin requests (embed scenarios)
- [ ] Same validation rules apply: message max 4000 chars, missing field rejection
- [ ] Public chatbot endpoint only works for chatbots with `isPublic: true`
- [ ] Non-public chatbot returns 404 (information hiding)
- [ ] Response format matches authenticated chat exactly
