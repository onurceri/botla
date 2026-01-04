# Task: Implement Admin Dashboard Tests

> **Task ID**: 29-admin-dashboard  
> **Source**: TEST_PATHS.md Section 9.1  
> **Priority**: Medium (Admin Panel)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Admin Dashboard including system overview, stats cards, health indicators, and quick actions.

### Reference Specifications (Section 9.1)

- Admin dashboard accessible only to admins
- System stats cards: Total Users, Total Organizations, Total Chatbots, Active Sessions, Queue Jobs, API Response Time
- Recent activity: Latest signups, Latest chatbots, System errors
- Health indicators: Database, Redis, Qdrant, Storage with status
- Quick actions: Flush cache, Run migrations, Send announcement, Export stats
- Charts: Daily active users, Chatbot creations, Token usage, API requests

### Implementation Requirements

1. `frontend/e2e/admin-dashboard.spec.ts`
2. `frontend/e2e/pages/admin-dashboard.page.ts`
3. `frontend/e2e/mocks/admin.mocks.ts`

---

## Implementation Plan

- Admin access tests (non-admin blocked)
- System stats display tests
- Recent activity tests
- Health indicator tests
- Quick action tests
- Chart display tests

---

## Dependencies

- **Prerequisites**: None (requires admin user)

---

## Related Tasks

- 30-admin-users.md - User management
- 31-admin-organizations.md - Organization management
- 32-admin-health.md - System health
- 33-admin-queues.md - Queue monitoring
- 34-admin-errors.md - Error logs

---

*Task created from: docs/frontend/TEST_PATHS.md Section 9.1*
