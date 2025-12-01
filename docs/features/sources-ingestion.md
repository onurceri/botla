# Sources Ingestion

## Purpose
Attach knowledge sources (PDF, URL, Text) to chatbots for RAG.

## User Flow
1. View sources list for a chatbot.
2. Add source (PDF upload, URL, or pasted text).
3. Processing queue ingests and indexes content.
4. Check status; delete source when needed.

## Backend Interfaces
- List/Add under chatbot: `cmd/server/main.go:67-78`, `internal/api/handlers/source.go:23-163`.
- Status/Delete: `cmd/server/main.go:84`, `internal/api/handlers/source.go:165-227`.
- Queue: `internal/processing/sources_queue.go` (enqueue, worker).
- Storage: `pkg/storage/r2.go`.

## Frontend Interfaces
- Component: `frontend/src/components/chatbot/SourceUploader.tsx:14-172`.
- API: `frontend/src/api/source.ts`.

## Constraints
- PDF max size 50MB; content-type check; storage required for file/text.

## Error Handling
- Clear status codes; frontend toasts for success/failure; standardized messages.

## Testing
- Backend ingestion tests for PDF/URL/Text.
- Frontend uploader interactions and error display.

