# Widget Task 005: Embed Token Error Handling

## Background
If an embed token is required but fails to fetch, the widget currently logs a warning and proceeds, leading to a likely 401 Authorization error later. This should stop the flow or notify the user/developer.

**File:** `widget/src/widgetApp.tsx`
**Location:** Lines 221-222

## Integration Plan
1.  **Catch Token Failure**
    - In the `catch` block for token fetch, instead of just logging, set a specific error state.

2.  **Prevent Chat**
    - If token missing (and `secure_embed_enabled` is true), prevent sending messages.
    - Show an error in the UI: "Unable to initialize secure session."

3.  **Verify**
    - Configure chatbot with secure embed.
    - Simulate token fetch failure.
    - Verify helpful error message appears.

## Checklist
- [ ] Update token fetch error handling
- [ ] Implement UI feedback for token failure
- [ ] Prevent message sending if security requirements not met
