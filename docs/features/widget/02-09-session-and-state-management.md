> Extracted from: widget.md — Section: 2.9 Session & State Management

## 2.9 Session & State Management

- [ ] Session ID is generated via `crypto.randomUUID()`
- [ ] Session is persisted to localStorage with prefix `chatbot_session_`
- [ ] `getSession()` retrieves session from localStorage
- [ ] `saveSession()` writes session to localStorage
- [ ] `clearSession()` removes session from localStorage
- [ ] `ensureSession()` creates a new session if none exists
- [ ] Messages are restored on page reload
- [ ] Session expires after configured TTL
- [ ] `data-reset-session` clears existing session on mount
- [ ] `data-session-id` uses explicit ID instead of generated one
- [ ] Session data includes: messages, sessionId, chatbotId, createdAt
- [ ] Multiple chatbots on the same page have independent sessions
