# Widget Task 001: Fix Stale Input Closure

## Background
In `widget/src/widgetApp.tsx`, `pickSuggestion` sets state `input` and then immediately calls `send()` inside `setTimeout`. Due to React's closure and batching, `send()` might use the old `input` value.

**File:** `widget/src/widgetApp.tsx`
**Location:** Lines 280-286

## Integration Plan
1.  **Modify Send Function**
    - Update `send` to accept an optional `overrideContent` argument.
    - If provided, use `overrideContent` instead of `input` state.

2.  **Update PickSuggestion**
    - Call `send(q)` directly instead of setting state and waiting.
    - (Optional) Still set state to update UI if needed, but send logic shouldn't depend on it.

3.  **Verify**
    - Click a suggestion. Verify correct text is sent immediately.

## Checklist
- [ ] Refactor `send` function signature
- [ ] Update `pickSuggestion` to pass text directly
- [ ] Verify functionality
