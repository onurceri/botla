# Task: Implement System Health Tests

> **Task ID**: 32-admin-health  
> **Source**: TEST_PATHS.md Section 9.4  
> **Priority**: Medium (Admin Panel)  
> **Estimated Effort**: 4-6 hours  

---

## Detailed Prompt

Implement E2E tests for System Health Monitoring including service status, health check execution, logs, and alert configuration.

### Reference Specifications (Section 9.4)

- Service status grid: PostgreSQL, Redis, Qdrant, Storage (R2), OpenAI API, Email service
- Each service shows status indicator, response time, last successful check, error message if failed
- Run health check button refreshes all services
- Service logs with filtering by level (error/warn/info) and search
- Alert configuration: CPU threshold, Memory threshold, Response time threshold, Email recipients
- Incident history with service, start/end time, severity, description

### Implementation Requirements

1. `frontend/e2e/admin-health.spec.ts`
2. `frontend/e2e/pages/admin-health.page.ts`
3. `frontend/e2e/mocks/admin.mocks.ts`

---

## Implementation Plan

- Service status grid display tests
- Status indicator tests
- Run health check tests
- Service logs display tests
- Log filtering tests
- Alert configuration tests
- Incident history tests

---

## Dependencies

- **Prerequisites**: 29-admin-dashboard.md (admin access)

---

## Related Tasks

- 29-admin-dashboard.md - Admin dashboard
- 33-admin-queues.md - Queue monitoring

---

*Task created from: docs/frontend/TEST_PATHS.md Section 9.4*
