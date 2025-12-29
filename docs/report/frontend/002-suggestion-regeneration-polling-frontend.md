# Frontend Task 002: Suggestion Regeneration Polling

## Background
`useRegenerateSuggestions` in `src/hooks/mutations/useChatbotMutations.ts` waits for a hardcoded 2000ms. This is flaky.

**File:** `src/hooks/mutations/useChatbotMutations.ts`
**Location:** Lines 81-83

## Integration Plan
1.  **Remove Hardcoded Wait**
    - Remove `await new Promise((resolve) => setTimeout(resolve, 2000))`.

2.  **Implement Polling (or Wait for Response)**
    - *Dependent on Backend Task 007.*
    - If backend returns a Job ID: Implement a polling loop checking status every 1s until completed/failed.
    - If backend is synchronous: Just awaiting the API call is sufficient (remove explicit sleep).

3.  **Update UI**
    - Ensure the mutation uses `isPending` state to show a loading spinner during this process.

## Checklist
- [x] Sync with Backend Task 007 outcome
- [x] Remove hardcoded sleep
- [x] Implement polling if async, or rely on promise resolution if sync
- [x] Verify UI feedback during regeneration

