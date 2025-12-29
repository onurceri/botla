# Backend Task 007: Suggestion Regeneration Polling Support

## Background
The frontend currently waits a hardcoded 2 seconds for suggestion regeneration to complete (`useChatbotMutations.ts`). This is fragile. The backend endpoint `POST /suggestions/regenerate` likely triggers a background job or runs synchronously but potentially slowly.

**Goal:** Ensure the backend provides a way to track the status of this operation, or returns immediately if it's async, allowing the frontend to poll or wait for a specific "completed" signal/job status.

## Integration Plan
1.  **Analyze Endpoint**
    - Check `internal/api/handlers/chatbot_suggestions.go` (inferred path).
    - Determine if it runs synchronously or asynchronously.

2.  **Improve Response (If Async)**
    - If async, return a `job_id`.
    - Provide a status endpoint `GET /chatbots/{id}/suggestions/status`.

3.  **Improve Response (If Sync but Slow)**
    - If it's intended to be synchronous, ensure the timeout is sufficient and consider moving to async if >5s.
    - *Assuming Async is better:* Refactor to return 202 Accepted with Job ID.

4.  **Verify**
    - Test the endpoint response.

## Checklist
- [ ] Analyze `suggestions/regenerate` handler
- [ ] Return Job ID if async, or ensure consistent behavior
- [ ] Document appropriate polling mechanism for frontend
