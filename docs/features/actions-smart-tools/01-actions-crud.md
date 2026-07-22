> Extracted from: actions-smart-tools.md — Section: 1.1 Actions — CRUD

## 1.1 Actions — CRUD

- [ ] GET /api/v1/chatbots/{id}/actions lists all actions for a chatbot
- [ ] POST /api/v1/chatbots/{id}/actions creates a new action with name, description, type, endpoint, method, headers, body, parameters
- [ ] GET /api/v1/chatbots/{id}/actions/{actionId} returns a single action
- [ ] PUT /api/v1/chatbots/{id}/actions/{actionId} updates an action
- [ ] DELETE /api/v1/chatbots/{id}/actions/{actionId} deletes an action
- [ ] GET /api/v1/chatbots/{id}/actions/logs returns execution logs for actions
- [ ] Invalid chatbot ID returns 404 on all action endpoints
- [ ] Action not found returns 404 on single-action endpoints
- [ ] Validation: name and endpoint are required
- [ ] Validation: type must be `http` or `function`
- [ ] Validation: invalid URL format is rejected
- [ ] Authorization: user must have access to the chatbot's org
