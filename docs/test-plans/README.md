# Test Plans Index

This directory contains comprehensive test plans for improving the Botla test suite.

## Overview

| Plan ID | Name | Priority | Duration | Status |
|---------|------|----------|----------|--------|
| TP-GO-COVERAGE-001 | Go Backend Coverage Improvement | CRITICAL | 2-3 weeks | Draft |
| TP-GO-PARALLEL-001 | Go Backend Parallelization Fix | CRITICAL | 1-2 weeks | Draft |
| TP-FRONTEND-MOCK-001 | Frontend Mock Cleanup | HIGH | 2-3 weeks | Draft |
| TP-WIDGET-COVERAGE-001 | Widget Test Coverage | HIGH | 2-3 weeks | Draft |
| TP-BACKEND-INTEGRATION-001 | Backend Integration Tests | HIGH | 2-3 weeks | Draft |

## Execution Order

These plans should be executed in the following order:

1. **TP-GO-PARALLEL-001** - Fix parallelization first (enables faster test execution)
2. **TP-GO-COVERAGE-001** - Increase coverage after parallelization
3. **TP-FRONTEND-MOCK-001** - Stabilize frontend tests
4. **TP-WIDGET-COVERAGE-001** - Expand widget tests
5. **TP-BACKEND-INTEGRATION-001** - Add real integration tests

## Total Timeline

| Phase | Duration | Focus |
|-------|----------|-------|
| Phase 1-2 | Weeks 1-3 | Go Backend (Parallelization + Coverage) |
| Phase 3 | Weeks 4-5 | Frontend Mock Cleanup |
| Phase 4 | Weeks 6-7 | Widget Coverage |
| Phase 5 | Weeks 8-9 | Backend Integration Tests |
| **Total** | **~9 weeks** | Complete test improvement |

## Quick Reference

### Go Backend Plans

- **[01-go-backend-coverage-improvement.md](01-go-backend-coverage-improvement.md)** - Increase coverage from 63% to 90%
- **[02-go-backend-parallelization-fix.md](02-go-backend-parallelization-fix.md)** - Remove t.Setenv() anti-pattern

### Frontend Plans

- **[03-frontend-mock-cleanup.md](03-frontend-mock-cleanup.md)** - Stabilize selectors, add data-testid, remove waitForTimeout

### Widget Plans

- **[04-widget-test-coverage.md](04-widget-test-coverage.md)** - Comprehensive widget E2E and unit tests

### Backend Integration Plans

- **[05-backend-integration-tests.md](05-backend-integration-tests.md)** - Real service integration tests

## Key Metrics

### Current State

| Metric | Value | Target |
|--------|-------|--------|
| Go Backend Coverage | 63.3% | 90% |
| Frontend Coverage | 80% | 85% |
| Widget Coverage | ~10% | 80% |
| Parallel Tests | Partial | Full |
| Real Integration Tests | 0 | Comprehensive |

### Target State

| Metric | Value | Notes |
|--------|-------|-------|
| Go Backend Coverage | 90%+ | After TP-GO-COVERAGE-001 |
| Frontend Coverage | 85%+ | After TP-FRONTEND-MOCK-001 |
| Widget Coverage | 80%+ | After TP-WIDGET-COVERAGE-001 |
| Test Execution Time | 50% faster | After TP-GO-PARALLEL-001 |
| Integration Tests | 100+ | After TP-BACKEND-INTEGRATION-001 |

## Agent Prompts

Each plan includes a detailed Sisyphus agent prompt at the beginning. When executing a plan:

1. Copy the agent prompt
2. Provide it to Sisyphus
3. Sisyphus will execute the plan step by step

## Progress Tracking

Each plan includes:
- Daily checklist
- Milestone reviews
- Success criteria
- Risk mitigation

## Dependencies

### Prerequisites

- Docker and Docker Compose
- PostgreSQL, Redis, Qdrant (for integration tests)
- Node.js and npm (for frontend tests)
- Go 1.25+ (for backend tests)

### Test Data

- `internal/testdb/` - Database test utilities
- `internal/testutils/` - Configuration utilities
- `internal/integration/fixtures/` - Test fixtures
- `frontend/e2e/helpers.ts` - E2E helpers
- `frontend/src/__tests__/factories.ts` - Data factories

## Verification

After each plan is executed:

```bash
# Go Backend
make test-all
make cover-gate

# Frontend
npm run test
npm run e2e

# Widget
cd widget && npm run test
cd widget && npm run e2e
```

## Contact

For questions about these plans, refer to:
- [AGENTS.md](../AGENTS.md) - Project conventions
- [CLAUDE.md](../CLAUDE.md) - Development guidelines
- [docs/test-plans/](./) - Test plans

## Revision History

| Version | Date | Author | Summary |
|---------|------|--------|---------|
| 1.0 | 2026-01-03 | Sisyphus | Initial test plans |
