# Task: Implement Privacy Settings Tests

> **Task ID**: 28-settings-privacy  
> **Source**: TEST_PATHS.md Section 8.5  
> **Priority**: Medium (Settings)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Privacy Settings including data export, account deletion, data retention, and activity log.

### Reference Specifications (Section 8.5)

- Privacy page shows current settings and data export options
- Data export: select data types (chatbots, conversations, sources, actions), generate export, download ZIP
- Delete account: password confirmation, type "DELETE", account deleted with KVKK compliance
- Data retention settings: 30 days, 90 days, 1 year, Forever
- Activity log: recent activity with action, timestamp, IP address, export option

### Implementation Requirements

1. `frontend/e2e/settings-privacy.spec.ts`
2. `frontend/e2e/pages/settings-privacy.page.ts`
3. `frontend/e2e/mocks/settings.mocks.ts`

---

## Implementation Plan

- Privacy settings display tests
- Data export flow tests
- Delete account flow tests
- Data retention settings tests
- Activity log display tests
- Activity log export tests

---

## Dependencies

- **Prerequisites**: None

---

## Related Tasks

- 24-settings-profile.md - Profile settings
- 27-settings-plan-billing.md - Plan and billing

---

*Task created from: docs/frontend/TEST_PATHS.md Section 8.5*
