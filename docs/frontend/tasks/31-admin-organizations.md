# Task: Implement Admin Organization Management Tests

> **Task ID**: 31-admin-organizations  
> **Source**: TEST_PATHS.md Section 9.3  
> **Priority**: Medium (Admin Panel)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Admin Organization Management including listing, viewing, editing, suspending, merging, and deleting organizations.

### Reference Specifications (Section 9.3)

- Organization table showing name, slug, owner, member count, chatbot count, plan, created date, status
- View organization detail panel with members list, chatbots list, usage stats, billing history
- Edit organization: name, plan
- Suspend organization (locks all users)
- Merge organizations (source to target)
- Delete organization with confirmation

### Implementation Requirements

1. `frontend/e2e/admin-organizations.spec.ts`
2. `frontend/e2e/pages/admin-organizations.page.ts`
3. `frontend/e2e/mocks/admin.mocks.ts`

---

## Implementation Plan

- Organization table display tests
- Organization detail view tests
- Edit organization tests
- Suspend organization tests
- Merge organizations tests
- Delete organization tests

---

## Dependencies

- **Prerequisites**: 29-admin-dashboard.md (admin access)

---

## Related Tasks

- 29-admin-dashboard.md - Admin dashboard
- 30-admin-users.md - User management

---

*Task created from: docs/frontend/TEST_PATHS.md Section 9.3*
