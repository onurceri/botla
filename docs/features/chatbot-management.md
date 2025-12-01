# Chatbot Management

## Purpose
Create, list, view, update, and soft-delete chatbots per user.

## User Flow
1. Navigate to Chatbots list.
2. Create new chatbot with styling and model options.
3. View and update chatbot details.
4. Delete chatbot (soft-delete).

## Backend Interfaces
- List/Create: `cmd/server/main.go:51`, `internal/api/handlers/chatbot.go:39-103`.
- By ID: `cmd/server/main.go:77-78`, `internal/api/handlers/chatbot.go:105-227`.
- DB: `internal/db/chatbot.go` (CRUD), `internal/models/chatbot.go`.

## Frontend Interfaces
- Pages: `frontend/src/pages/ChatbotsPage.tsx`, `frontend/src/pages/ChatbotDetailPage.tsx`.
- Components: `frontend/src/components/chatbot/*` (form, card, etc.).
- API: `frontend/src/api/chatbot.ts`.

## Error Handling
- 400 for bad names; 403 for ownership; 404 for missing; 405 for wrong method.

## Testing
- Backend integration tests for CRUD.
- Frontend form validation and update flows.

