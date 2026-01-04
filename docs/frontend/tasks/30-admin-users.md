# Task: Implement Admin User Management Tests

> **Task ID**: 30-admin-users  
> **Source**: TEST_PATHS.md Section 9.2  
> **Priority**: Medium (Admin Panel)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Admin User Management including user listing, search, filter, edit, suspend, and delete.

### Reference Specifications (Section 9.2)

- User table with pagination showing avatar, name, email, plan, status, created date, last login
- Search users by name or email
- Filter by plan (free/pro/ultra), status (active/suspended), date range
- View user detail panel with chatbots, usage stats, activity log, login history
- Edit user: name, plan, admin status
- Suspend/unsuspend user with reason
- Delete user with confirmation
- Export users in CSV/JSON

### Implementation Requirements

1. `frontend/e2e/admin-users.spec.ts`
2. `frontend/e2e/pages/admin-users.page.ts`
3. `frontend/e2e/mocks/admin.mocks.ts`

---

## Implementation Plan

- User table display tests
- Search functionality tests
- Filter functionality tests
- User detail view tests
- Edit user tests
- Suspend/unsuspend tests
- Delete user tests
- Export tests

---

## Dependencies

- **Prerequisites**: 29-admin-dashboard.md (admin access)

---

## Related Tasks

- 29-admin-dashboard.md - Admin dashboard
- 31-admin-organizations.md - Organization management

---

*Task created from: docs/frontend/TEST_PATHS.md Section 9.2*
