# Frontend Task 001: Improve Token Validation

## Background
In `src/App.tsx`, the `isValidToken` function only checks for non-null/non-empty strings. Malformed tokens (e.g., "abc") pass this check but fail API calls, leading to potential app crashes or loops.

**File:** `src/App.tsx`
**Location:** Lines 45-47

## Integration Plan
1.  **Refactor `isValidToken`**
    - Add a check to ensure the token has 3 parts separated by dots (standard JWT format).
    - Optionally decode the header/payload to check for expiry (though backend does this, rapid frontend check provides better UX).

2.  **Verify**
    - Test with "invalid_token" string -> Should return false.
    - Test with valid JWT -> Should return true.

## Checklist
- [ ] Update `isValidToken` in `src/App.tsx`
- [ ] Ensure it returns false for non-JWT strings
- [ ] Verify functionality with invalid local storage token
