# Widget Task 008: Font Loading Isolation

## Background
The widget loads fonts by appending `<link>` tags to the `targetDocument.head`. This affects the parent page ("polluting" the global style scope).

**File:** `widget/src/widgetApp.tsx`
**Location:** Lines 128-138

## Integration Plan
1.  **Assess Shadow DOM**
    - If the widget uses Shadow DOM (it likely should), append styles/fonts *inside* the shadow root.
    - If not using Shadow DOM, consider adopting it for full isolation.

2.  **Alternative (Scoped)**
    - Only load fonts if not already present? (Hard to detect perfectly).
    - Accept a prop to *disable* font loading if the host site wants to manage it.

3.  **Verify**
    - Check that widget fonts don't override host site fonts (Same name collision).

## Checklist
- [ ] Evaluate Shadow DOM implementation
- [ ] Move font loading into Shadow Root if possible
- [ ] Or add configuration to disable auto-font loading
