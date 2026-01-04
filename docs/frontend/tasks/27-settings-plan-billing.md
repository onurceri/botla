# Task: Implement Plan & Billing Tests

> **Task ID**: 27-settings-plan-billing  
> **Source**: TEST_PATHS.md Section 8.4  
> **Priority**: Medium (Settings)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Plan & Billing including plan details, usage statistics, upgrade flow, and invoice management.

### Reference Specifications (Section 8.4)

- Plan page shows current plan, monthly cost, renewal date, payment method
- Usage statistics: tokens used/limit, chatbots used/limit, storage used/limit, files uploaded/limit
- Available plans display with features list and upgrade buttons
- Upgrade flow: select plan, billing period, confirm upgrade
- Cancel subscription with consequences warning
- Update payment method via Stripe Elements
- Invoice history with download PDF

### Implementation Requirements

1. `frontend/e2e/settings-plan.spec.ts`
2. `frontend/e2e/pages/settings-plan.page.ts`
3. `frontend/e2e/mocks/billing.mocks.ts`

---

## Implementation Plan

- Current plan display tests
- Usage statistics tests
- Available plans display tests
- Upgrade flow tests
- Cancel subscription tests
- Payment method update tests
- Invoice download tests

---

## Dependencies

- **Prerequisites**: None

---

## Related Tasks

- 24-settings-profile.md - Profile settings
- 28-settings-privacy.md - Privacy settings

---

*Task created from: docs/frontend/TEST_PATHS.md Section 8.4*
