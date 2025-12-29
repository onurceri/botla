# Backend Task 007: Suggestion Regeneration Polling Support

## Background
The frontend previously waited a hardcoded 2 seconds for suggestion regeneration to complete (`useChatbotMutations.ts`). This was fragile and has been removed.

**Current State:** The backend endpoint `POST /suggestions/regenerate` is already async - it returns `202 Accepted` immediately and runs `ReAggregateSuggestionsForChatbot` in a background goroutine.

**Problem:** The backend doesn't return a job ID, so the frontend cannot poll for completion status. The frontend currently just invalidates queries after the 202 response, but there's no guarantee suggestions are ready.

**Goal:** Return a job ID from the regeneration endpoint and provide a status endpoint so the frontend can poll until suggestions are ready.

## Current Implementation
**File:** `internal/api/handlers/chatbot_suggestions.go`

```go
// Current: Returns 202 immediately, no job tracking
go processing.ReAggregateSuggestionsForChatbot(context.Background(), h.DB, chatbotID, h.Log)
w.WriteHeader(http.StatusAccepted)
```

## Integration Plan
1.  **Create Suggestion Regeneration Job**
    - Modify `RegenerateSuggestions` handler to create a job record in the database
    - Return `202 Accepted` with JSON body containing `job_id`

2.  **Add Status Endpoint**
    - Create `GET /api/v1/chatbots/{id}/suggestions/status` endpoint
    - Return job status: `pending`, `processing`, `completed`, `failed`
    - Include `suggested_questions` in response when `completed`

3.  **Update Processing Function**
    - Modify `ReAggregateSuggestionsForChatbot` to accept a job ID
    - Update job status to `processing` when starting
    - Update job status to `completed` or `failed` when done

4.  **Frontend Integration**
    - Once this is implemented, update `useRegenerateSuggestions` in frontend:
      - Parse job ID from 202 response
      - Poll status endpoint every 1s until completed/failed
      - See TODO in `frontend/src/hooks/mutations/useChatbotMutations.ts`

## Checklist
- [x] Analyze `suggestions/regenerate` handler (confirmed async with 202)
- [x] Create job tracking table/model for suggestion regeneration (or reuse existing job system)
- [x] Modify `RegenerateSuggestions` to return job ID in response body
- [x] Create `GET /chatbots/{id}/suggestions/status` endpoint
- [x] Update `ReAggregateSuggestionsForChatbot` to track job progress
- [x] Document polling mechanism for frontend
- [x] Update frontend `useRegenerateSuggestions` to implement polling (see TODO)

