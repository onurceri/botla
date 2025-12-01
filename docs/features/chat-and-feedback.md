# Chat & Feedback

## Purpose
RAG chat using indexed sources; user messages and assistant responses; thumbs feedback.

## User Flow
1. Start a session; send message.
2. Embed message; retrieve context (Qdrant); generate completion (OpenAI).
3. Store messages; update analytics.
4. Submit feedback for a message.

## Backend Interfaces
- Chat: `cmd/server/main.go:73-75`, `internal/api/handlers/chat.go:37-154`.
- Feedback: `cmd/server/main.go:81-82` (wrapped with auth), `internal/api/handlers/chat.go:160-194`.
- Retrieval/Completion: `internal/rag/*`.

## Frontend Interfaces
- Chat components: `frontend/src/components/chatbot/*`.
- API: `frontend/src/api/chatbot.ts`.

## Error Handling
- Timeouts for embeddings/completions; fallback answer when context empty.
- Feedback requires auth and validates message existence.

## Testing
- Backend chat and feedback tests.
- Frontend chat UI interaction and feedback toggles.

