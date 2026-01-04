# Task: Implement Edge Cases Tests

> **Task ID**: 37-edge-cases  
> **Source**: TEST_PATHS.md Section 11  
> **Priority**: Medium-Low (Quality Assurance)  
> **Estimated Effort**: 8-10 hours  

---

## Detailed Prompt

Implement E2E tests for Edge Cases and Error States including network errors, form validation, file upload errors, and modal handling.

### Reference Specifications (Section 11)

**Network Error Handling:**
- API timeout (30s) shows loading spinner timeout, toast, retry button
- 401 Unauthorized shows session expired modal, relogin redirect
- 403 Forbidden shows access denied message
- 404 Not Found shows 404 page with go home button
- 429 Rate Limited shows rate limit toast, retry timer
- 500 Server Error shows error page with error ID
- Network offline shows offline indicator, disabled API calls
- WebSocket disconnect shows reconnecting indicator

**Form Validation Errors:**
- Required field empty shows error message, icon, prevents submission
- Invalid format (email, URL, JSON) shows specific error and hints
- Length validation (min/max) shows appropriate messages
- Number validation (min/max) shows range errors
- Match validation (passwords, emails) shows mismatch message
- Custom validation (username taken, chatbot limit, file size)

**File Upload Errors:**
- Wrong file type shows "Only X allowed"
- File too large shows "Max X MB"
- Corrupted file shows parsing error with suggestion
- Network interruption shows pause with resume/cancel
- Virus detected shows security rejection
- Storage quota exceeded shows upgrade option

**Modal/Confirmation Dialogs:**
- Open modal shows correct content, focus trap, body scroll lock
- Close modal via X, Cancel, Overlay, Escape
- Confirmation dialog with warning, destructive action color, text confirmation
- Unsaved changes shows discard/save/keep editing options
- Loading state disables buttons, prevents close

### Implementation Requirements

1. `frontend/e2e/edge-cases.spec.ts`
2. `frontend/e2e/utils/error-injection.ts`
3. `frontend/e2e/mocks/errors.mocks.ts`

---

## Implementation Plan

- Network error handling tests (all HTTP status codes)
- Form validation tests
- File upload error tests
- Modal behavior tests
- Offline handling tests
- WebSocket disconnect tests

---

## Dependencies

- **Prerequisites**: All authentication and core feature tests

---

## Related Tasks

- 38-accessibility.md - Accessibility tests
- 39-performance.md - Performance tests

---

*Task created from: docs/frontend/TEST_PATHS.md Section 11*
