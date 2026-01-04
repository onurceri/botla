# Task: Implement Organization Settings Tests

> **Task ID**: 25-settings-organization  
> **Source**: TEST_PATHS.md Section 8.2  
> **Priority**: Medium (Settings)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Organization Settings including name edit, logo upload, member management, and invite flow.

### Reference Specifications (Section 8.2)

- Organization page shows name, slug, member count, created date
- Edit organization: name edit, logo upload
- Members table shows name, email, role (owner/admin/member), status (active/invited)
- Invite member flow: email input, role select, send invite
- Update member role, remove member, resend invitation
- Delete organization with confirmation

### Implementation Requirements

1. `frontend/e2e/settings-organization.spec.ts`
2. `frontend/e2e/pages/settings-organization.page.ts`
3. `frontend/e2e/mocks/settings.mocks.ts`

---

## Implementation Plan

- Organization info display tests
- Edit organization tests
- Members table display tests
- Invite member flow tests
- Role update tests
- Remove member tests
- Delete organization tests

---

## Dependencies

- **Prerequisites**: None

---

## Related Tasks

- 24-settings-profile.md - Profile settings
- 26-settings-workspace.md - Workspace settings
- 28-settings-privacy.md - Privacy settings

---

*Task created from: docs/frontend/TEST_PATHS.md Section 8.2*
