# Task: Implement Queue Monitoring Tests

> **Task ID**: 33-admin-queues  
> **Source**: TEST_PATHS.md Section 9.5  
> **Priority**: Medium (Admin Panel)  
> **Estimated Effort**: 6-8 hours  

---

## Detailed Prompt

Implement E2E tests for Queue Monitoring including queue overview, job list, job actions, and bulk operations.

### Reference Specifications (Section 9.5)

- Queue overview showing all job queues: Source processing, Embedding generation, Refresh jobs, Retention jobs, Analytics aggregation
- Each queue displays: Pending jobs, Processing jobs, Failed jobs, Average wait time, Success rate
- Job list with ID, status, progress, created at, started at
- Filter jobs by status, date range, search by ID
- Job actions: Retry failed job, Cancel pending job, View job log
- Bulk actions: Select multiple jobs, Retry selected
- Pause/Resume queue, Clear queue with confirmation

### Implementation Requirements

1. `frontend/e2e/admin-queues.spec.ts`
2. `frontend/e2e/pages/admin-queues.page.ts`
3. `frontend/e2e/mocks/admin.mocks.ts`

---

## Implementation Plan

- Queue overview display tests
- Queue details view tests
- Job list display tests
- Job filter tests
- Job action tests (retry, cancel, view log)
- Bulk action tests
- Pause/Resume queue tests
- Clear queue tests

---

## Dependencies

- **Prerequisites**: 29-admin-dashboard.md (admin access)

---

## Related Tasks

- 29-admin-dashboard.md - Admin dashboard
- 32-admin-health.md - System health
- 34-admin-errors.md - Error logs

---

*Task created from: docs/frontend/TEST_PATHS.md Section 9.5*
