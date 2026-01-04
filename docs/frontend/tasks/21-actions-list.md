# Task: Implement Actions List Tests

> **Task ID**: 21-actions-list  
> **Source**: TEST_PATHS.md Section 7.1  
> **Priority**: Medium (Smart Actions)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for the Actions List Page showing HTTP and Function-based smart actions.

### Reference Specifications (Section 7.1)

- Empty state when no actions exist
- Action cards grid with name, description, type badge, endpoint/method, status toggle
- Hover shows Edit and Delete buttons
- Toggle action status on/off
- View action logs showing execution history with timestamps, status, duration, request/response preview

### Implementation Requirements

1. `frontend/e2e/actions-list.spec.ts`
2. `frontend/e2e/pages/actions-list.page.ts`
3. `frontend/e2e/mocks/actions.mocks.ts`

---

## Implementation Plan

- Empty state tests
- Action cards display tests
- Status toggle tests
- Hover states tests
- Action logs view tests
- Search and filter tests

---

## Dependencies

- **Prerequisites**: None

---

## Related Tasks

- 22-action-create.md - Create action
- 23-action-logs.md - Action logs
- 10-chatbots-create.md - Creates chatbot with actions

---

*Task created from: docs/frontend/TEST_PATHS.md Section 7.1*
