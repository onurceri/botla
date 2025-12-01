# Architecture Overview

## Backend
- Language: Go (`net/http`) with layered handlers, middleware, services, and DB access.
- Entry: `cmd/server/main.go:22-38` launches server and wraps router with CORS and request logging.
- Routing: `cmd/server/main.go:40-90` registers all endpoints using `http.ServeMux`.
- Auth: `pkg/middleware/auth.go:15-37` verifies `Authorization: Bearer` JWT access tokens; `internal/auth/jwt.go:15-45` issues and verifies tokens.
- Handlers: `internal/api/handlers/*` implement feature endpoints (auth, chatbots, sources, chat, analytics, health).
- Processing: `internal/processing/*` queues and ingests sources; `internal/rag/*` handles embeddings and retrieval with OpenAI and Qdrant.
- Storage: `pkg/storage/*` integrates Cloudflare R2 for PDF/text content.
- Database: `internal/db/*` CRUD for users, chatbots, sources, conversations, messages, analytics.
- Tests: `internal/integration/*` and unit tests across modules.

## Frontend
- Framework: React + TypeScript + Vite; routing with `react-router-dom`.
- Entry: `frontend/src/main.tsx:7-15` mounts app inside `QueryClientProvider`.
- Router: `frontend/src/App.tsx:20-48` defines public and protected routes with `PrivateRoute`.
- UI: Tailwind-based components under `frontend/src/components/ui/*` and shared layout under `frontend/src/components/layout/*`.
- State/Networking: `@tanstack/react-query` for server state; `frontend/src/api/client.ts:3-6,8-46` central axios client with auth/refresh interceptors.

## Features
- Authentication (register, login, refresh, logout, protected ping)
- Chatbot management (list, create, get, update, delete)
- Sources ingestion (PDF/URL/Text upload and processing, status, delete)
- Chat & feedback (RAG retrieval, completion, message feedback)
- Analytics (7-day aggregates per user’s chatbots)
- Dashboard & settings UI

## Integrations
- OpenAI (embeddings, chat completions)
- Qdrant (vector storage and search)
- Cloudflare R2 (file storage)

## Security Notes
- Protected routes enforce JWT via `AuthMiddleware`.
- Refresh tokens are rotated and stored in DB.
- Feedback endpoint requires protection (see feature docs).

