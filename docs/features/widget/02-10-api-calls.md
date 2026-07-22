> Extracted from: widget.md — Section: 2.10 API Calls

## 2.10 API Calls

### 2.10.1 Chat (`sendMessage`)

- [ ] POST to `/api/v1/public/chatbots/{chatbotId}/chat`
- [ ] Request body includes: `message`, `sessionId`, `history` (message array)
- [ ] Response includes: `reply` (text), `suggestions` (array), `sources` (array, optional)
- [ ] Loading state is shown while waiting for response
- [ ] Network errors show a retry-friendly error message
- [ ] 4xx errors show appropriate user-facing messages
- [ ] 5xx errors show "Service unavailable, please try again"
- [ ] Rate limit (429) shows a "Too many messages, please slow down" message
- [ ] Invalid chatbot-id (404) shows a "Chatbot not found" or misconfigured message

### 2.10.2 Config (`GET /api/v1/public/chatbots/{id}/config`)

- [ ] Fetches chatbot configuration on mount
- [ ] Merges server-side config with `data-*` attributes (server takes precedence for defined values)
- [ ] Caches config in session storage to avoid repeated requests
- [ ] Falls back to `data-*` attributes if fetch fails

### 2.10.3 Feedback

- [ ] POST to `/api/v1/public/chatbots/{chatbotId}/feedback`
- [ ] Request body includes: `sessionId`, `messageId`, `rating` (up/down)
- [ ] Visual feedback (color change) confirms submission
- [ ] Duplicate feedback requests are prevented (idempotent)

### 2.10.4 Handoff

- [ ] POST to `/api/v1/public/chatbots/{chatbotId}/handoff/{requestId}/contact`
- [ ] Request body includes: `email`, `name` (optional), `message`
- [ ] Email field is validated before submission
- [ ] Success state is displayed after submission
- [ ] Error state with retry is displayed on failure

### 2.10.5 Embed Token

- [ ] GET request to `data-embed-token-url` on mount
- [ ] Token is appended to all subsequent API requests as `X-Embed-Token` header
- [ ] If token fetch fails (4xx/5xx/network), widget shows auth error
- [ ] Token expiry causes widget to re-fetch or show expiry message
