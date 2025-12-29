# Frontend Task 008: Refresh Token Race Condition

## Background
In `src/api/client.ts`, the refresh token logic shares a promise. If the refresh fails, all waiting requests fail. Ideally, if it fails, they should fail, but we need to ensure the queueing logic is robust. The current "wait for promise then continue" might retry the request with the *old* token if not careful, or assume success.

**File:** `src/api/client.ts`
**Location:** Lines 97-120

## Integration Plan
1.  **Analyze Request Queueing**
    - Ensure that after `await refreshPromise`, the request is retried with the *new* token (header injection needs to happen *after* the wait).
    - `axios` interceptors usually run before request send. If we are in response interceptor (retry), we need to update the config's Authorization header explicitly.

2.  **Implement Failed Request Queue (Optional but Better)**
    - A specific queue `failedQueue` that replays requests with new token.

3.  **Verify**
    - Simulate 401. Fire 5 requests at once.
    - Ensure refresh happens once.
    - Ensure all 5 requests retry with new token.

## Checklist
- [x] Review refresh logic in `client.ts`
- [x] Ensure retried requests pick up new token
- [x] Verify with concurrent requests
