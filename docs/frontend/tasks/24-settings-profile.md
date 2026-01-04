# Task: Implement Profile Settings Tests

> **Task ID**: 24-settings-profile  
> **Source**: TEST_PATHS.md Section 8.1  
> **Priority**: Medium (Settings)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Profile Settings including avatar upload, name editing, language preference, and password change.

### Reference Specifications (Section 8.1)

- Profile page shows current values: Avatar, Email (read-only), Full name, Language preference
- Edit profile flow: Click edit, upload avatar, edit name, change language, cancel
- Change password modal with current, new, confirm password
- Validation for empty name, invalid avatar, weak password

### Implementation Requirements

1. `frontend/e2e/settings-profile.spec.ts`
2. `frontend/e2e/pages/settings-profile.page.ts`
3. `frontend/e2e/mocks/settings.mocks.ts`

---

## Implementation Plan

- Page load and display tests
- Edit profile flow tests
- Avatar upload tests
- Name edit tests
- Language change tests
- Password change modal tests
- Validation tests

---

## Dependencies

- **Prerequisites**: None

---

## Related Tasks

- 25-settings-organization.md - Organization settings
- 26-settings-workspace.md - Workspace settings
- 27-settings-plan-billing.md - Plan and billing

---

*Task created from: docs/frontend/TEST_PATHS.md Section 8.1*
