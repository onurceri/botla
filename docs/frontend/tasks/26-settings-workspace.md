# Task: Implement Workspace Settings Tests

> **Task ID**: 26-settings-workspace  
> **Source**: TEST_PATHS.md Section 8.3  
> **Priority**: Medium (Settings)  
> **Estimated Effort**: 4-6 hours  

---

## Detailed Prompt

Implement E2E tests for Workspace Settings including workspace creation, editing, deletion, and default workspace selection.

### Reference Specifications (Section 8.3)

- Workspace page lists all workspaces with name, slug, chatbot count
- Create workspace: name input, auto-generated slug, create button
- Edit workspace: name change, save
- Delete workspace with confirmation
- Set default workspace

### Implementation Requirements

1. `frontend/e2e/settings-workspace.spec.ts`
2. `frontend/e2e/pages/settings-workspace.page.ts`
3. `frontend/e2e/mocks/settings.mocks.ts`

---

## Implementation Plan

- Workspace list display tests
- Create workspace tests
- Edit workspace tests
- Delete workspace tests
- Set default workspace tests
- Empty state tests

---

## Dependencies

- **Prerequisites**: None

---

## Related Tasks

- 25-settings-organization.md - Organization settings
- 24-settings-profile.md - Profile settings

---

*Task created from: docs/frontend/TEST_PATHS.md Section 8.3*
