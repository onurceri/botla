# Widget Task 007: Debounce Send Function

## Background
Rapid clicking on the send button could trigger multiple `send()` calls before the `loading` state updates.

**File:** `widget/src/widgetApp.tsx`
**Location:** Lines 204-278

## Integration Plan
1.  **Implement Debounce**
    - Wrap the `send` function (or the button click handler) with a debounce utility (e.g. 300ms).
    - Or ensure `loading` state is set *synchronously* and immediately (which React state setter usually is inside event handlers, but batching can vary).
    - A simple boolean ref `isSendingRef` can guard against this reliably.

2.  **Verify**
    - Rapidly click send. Ensure only one message goes out.

## Checklist
- [ ] Add `isSending` ref or debounce
- [ ] Guard `send` function
- [ ] Verify no duplicate messages
