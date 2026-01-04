# Task: Implement Action Create/Edit Tests

> **Task ID**: 22-action-create-edit  
> **Source**: TEST_PATHS.md Sections 7.2-7.3  
> **Priority**: Medium (Smart Actions)  
> **Estimated Effort**: 8-10 hours  

---

## Detailed Prompt

Implement E2E tests for Action Create and Edit flows with HTTP and Function configurations.

### Reference Specifications (Sections 7.2-7.3)

**HTTP Action Configuration:**
- Name input, Description, Method select (GET/POST/PUT/DELETE)
- Endpoint input with URL validation
- Headers and Body (JSON) with validation
- Parameters array (name, type, required, default)

**Function Action Configuration:**
- Code editor with JavaScript function
- Syntax validation
- Test function execution

**Flow:**
- Open create/edit modal with pre-filled values
- Modify fields and save
- Toggle enabled status
- Test action before saving

### Implementation Requirements

1. `frontend/e2e/action-create.spec.ts`
2. `frontend/e2e/pages/action-form.page.ts`
3. `frontend/e2e/mocks/actions.mocks.ts`

---

## Implementation Plan

- Modal open/close tests
- HTTP action configuration tests
- Function action configuration tests
- Parameter configuration tests
- Test action execution tests
- Save and validation tests
- Edit pre-filled values tests

---

## Dependencies

- **Prerequisites**: 21-actions-list.md

---

## Related Tasks

- 21-actions-list.md - Actions list page
- 23-action-logs.md - Action execution logs

---

*Task created from: docs/frontend/TEST_PATHS.md Sections 7.2-7.3*
