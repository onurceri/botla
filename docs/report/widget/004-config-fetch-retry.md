# Widget Task 004: Config Fetch Retry

## Background
The widget fetches configuration once on load. If this fails (network blip), the widget remains broken.

**File:** `widget/src/widgetApp.tsx`
**Location:** Lines 56-74

## Integration Plan
1.  **Implement Retry Logic**
    - Wrap the fetch in a retry loop (e.g., 3 attempts with exponential backoff).
    - Or use a utility like `fetchWithRetry`.

2.  **Verify**
    - Simulate network failure for first 2 requests, succeed on 3rd.
    - Verify widget loads successfully.

## Checklist
- [ ] Create `fetchWithRetry` utility or inline logic
- [ ] Apply to config fetch
- [ ] Verify behavior on temporary network failure
