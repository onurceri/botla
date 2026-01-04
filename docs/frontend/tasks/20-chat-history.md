# Task: Implement Chat History Tests

> **Task ID**: 20-chat-history  
> **Source**: TEST_PATHS.md Section 6.2  
> **Priority**: Medium (Chat & Actions)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Chat History functionality including conversation list, renaming, deleting, and exporting.

### Reference Specifications (Section 6.2)

- Conversation list sidebar with all conversations
- Each conversation shows title, date, message count
- Click conversation loads messages
- Hover shows options menu (rename, delete, export)
- Export chat in JSON, Markdown, PDF formats
- Search in chat highlights matching messages

### Implementation Requirements

1. `frontend/e2e/chat-history.spec.ts`
2. `frontend/e2e/pages/chat-history.page.ts`
3. `frontend/e2e/mocks/chat-history.mocks.ts`

---

## Implementation Plan

- Page load and conversation list tests
- Conversation selection tests
- Rename conversation tests
- Delete conversation tests
- Export chat tests
- Search in chat tests

---

## Dependencies

- **Prerequisites**: 19-playground.md (playground page)

---

## Related Tasks

- 19-playground.md - Chat playground
- 21-actions-list.md - Smart actions

---

*Task created from: docs/frontend/TEST_PATHS.md Section 6.2*
