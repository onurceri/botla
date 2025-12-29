# Frontend Task 006: Session Event Cleanup

## Background
In `App.tsx`, the `session-expired` event listener might be re-registered unnecessarily or not cleaned up correctly in edge cases due to the `toast` dependency in `useEffect`.

**File:** `src/App.tsx`
**Location:** Lines 66-72

## Integration Plan
1.  **Refactor Effect**
    - Use `useCallback` for the handler or move it inside `useEffect`.
    - Ensure `removeEventListener` is always called with the exact same function reference.
    - Check if `toast` dependency is truly needed or if a ref can be used.

2.  **Verify**
    - Add console logs on mount/unmount of the effect.
    - Trigger session expiry. Ensure only one toast/redirect happens.

## Checklist
- [x] Refactor `session-expired` useEffect
- [x] Ensure robust cleanup
- [x] Verify single event handling
