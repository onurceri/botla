# Task: Implement Test Naming Conventions

> **Task ID**: 01-test-naming-conventions  
> **Source**: TEST_PATHS.md Section 1  
> **Priority**: Highest (Foundation)  
> **Estimated Effort**: 2-4 hours

---

## Detailed Prompt

Implement comprehensive test naming conventions for all E2E and integration tests in the Botla-Co frontend test suite. This foundational task establishes the naming standards that all subsequent test implementations must follow.

### Context

The Botla-Co project uses Playwright for E2E testing. Currently, the test files exist in `frontend/e2e/` directory but lack consistent naming conventions and element identifiers. This task creates the foundational naming standards that ensure:
- Consistent element identification across all test files
- Maintainable and readable test code
- Clear mapping between test IDs and UI components
- Improved debugging experience with descriptive selectors

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 1:

#### 1.1 File Naming Pattern
```
{page-or-feature}.spec.ts
```

#### 1.2 Test Naming Pattern
```typescript
test.describe('Feature Area', () => {
  test('should perform action when user does X', async () => { ... });
  test('should show error when Y condition', async () => { ... });
  test('should handle hover state on element', async () => { ... });
});
```

#### 1.3 Element Naming Convention

| Element Type | Prefix | Example |
|--------------|--------|---------|
| Button | `btn` | `btn-create-chatbot` |
| Input | `input` | `input-email` |
| Select | `select` | `select-plan` |
| Link | `link` | `link-login` |
| Tab | `tab` | `tab-settings` |
| Modal | `modal` | `modal-confirm-delete` |
| Toast | `toast` | `toast-success` |
| Dropdown | `dropdown` | `dropdown-menu` |
| Checkbox | `checkbox` | `checkbox-terms` |
| Radio | `radio` | `radio-mode` |
| Toggle | `toggle` | `toggle-visibility` |

### Implementation Requirements

1. **Create a Test Naming Standards Document** (`frontend/e2e/TESTING_STANDARDS.md`)
   - Document all naming conventions
   - Provide examples for each element type
   - Include anti-patterns to avoid

2. **Create a Shared Test Helpers Module** (`frontend/e2e/utils/test-helpers.ts`)
   - Export element locator functions
   - Provide consistent selector generation
   - Include validation helpers

3. **Create Element ID Constants** (`frontend/e2e/utils/selectors.ts`)
   - Define constants for all common elements
   - Use consistent naming across all test files

4. **Update Existing Test Files** to follow conventions:
   - `frontend/e2e/auth.spec.ts`
   - `frontend/e2e/chatbot.spec.ts`
   - `frontend/e2e/smoke.spec.ts`
   - `frontend/e2e/widget-embed.spec.ts`
   - `frontend/e2e/widget-embed-secure.spec.ts`
   - `frontend/e2e/widget-branding.spec.ts`
   - `frontend/e2e/mobile-responsiveness.spec.ts`
   - `frontend/e2e/chunk-inspector.spec.ts`

5. **Create a Naming Linter Rule** (optional, for future implementation)
   - ESLint configuration for test naming

### Expected Deliverables

1. `frontend/e2e/TESTING_STANDARDS.md` - Comprehensive naming documentation
2. `frontend/e2e/utils/selectors.ts` - Element ID constants
3. `frontend/e2e/utils/test-helpers.ts` - Test helper utilities
4. Updated existing test files with consistent naming
5. Updated `frontend/e2e/playwright.config.ts` if needed

---

## Implementation Plan

### Phase 1: Create Documentation and Constants

- [x] Create `frontend/e2e/utils/selectors.ts` with all element ID constants
- [x] Create `frontend/e2e/TESTING_STANDARDS.md` with comprehensive documentation
- [x] Create `frontend/e2e/utils/test-helpers.ts` with helper functions

### Phase 2: Update Existing Test Files

- [x] Update `frontend/e2e/auth.spec.ts` with consistent naming
- [x] Update `frontend/e2e/chatbot.spec.ts` with consistent naming
- [x] Update `frontend/e2e/smoke.spec.ts` with consistent naming
- [x] Update all widget test files with consistent naming
- [x] Update `frontend/e2e/mobile-responsiveness.spec.ts` with consistent naming
- [x] Update `frontend/e2e/chunk-inspector.spec.ts` with consistent naming

### Phase 3: Verification

- [x] Run linter to check for issues
- [x] Run existing tests to ensure nothing is broken
- [x] Verify all element selectors are working correctly
- [x] Document any edge cases or special considerations

---

## Technical Notes

### Selector Strategy

Use `data-testid` attributes for element identification. This provides:
- Stable selectors that don't break with UI changes
- Clear semantic meaning
- Easy debugging

Example:
```typescript
// Instead of:
await page.locator('.btn-primary').click();
await page.locator('button[type="submit"]').click();

// Use:
await page.locator('[data-testid="btn-login"]').click();
await page.locator('[data-testid="btn-submit"]').click();
```

### Test Organization

Follow the `test.describe()` pattern for grouping:

```typescript
test.describe('Login Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    // Test implementation
  });

  test('should show error with invalid email', async ({ page }) => {
    // Test implementation
  });
});
```

### Naming Best Practices

1. **Be Descriptive**: `btn-create-chatbot` not `btn-create`
2. **Use kebab-case**: All lowercase with hyphens
3. **Include Context**: `input-email` not `input`
4. **Prefix Actions**: `btn-` for actions, `link-` for navigation
5. **Use Semantic Names**: `btn-save` not `btn-s`

---

## Verification Steps

### 1. Documentation Review
- [x] Naming conventions document is complete
- [x] All element types are documented
- [x] Examples are clear and actionable

### 2. Code Review
- [x] All selectors use `data-testid` attributes
- [x] Selectors follow the naming convention
- [x] Test files are organized with `test.describe()`
- [x] Constants are properly exported and typed

### 3. Test Execution
- [x] All existing tests pass
- [x] New selectors work correctly
- [x] No breaking changes introduced

### 4. Consistency Check
- [x] All test files use consistent naming
- [x] Element IDs follow the convention table
- [x] No hardcoded selectors without `data-testid`

---

## Execution Notes for Developer Agent

When implementing this task:

1. **First, read the existing test files** to understand current patterns
2. **Create the selectors.ts file first** - this is the foundation
3. **Update test files incrementally** - one at a time
4. **Run tests after each update** to ensure nothing breaks
5. **Use the `test.only()` pattern** temporarily to debug individual tests

### Running Tests for This Task

To save time, run only the tests you're modifying:

```bash
# Run a specific test file
cd frontend && npx playwright test e2e/auth.spec.ts

# Run tests with a specific name
cd frontend && npx playwright test -g "login"

# Run in headed mode for debugging
cd frontend && npx playwright test auth.spec.ts --headed
```

### Expected Challenges

1. **Mixed naming conventions** in existing tests - fix consistently
2. **Missing `data-testid` attributes** - add to components if needed
3. **Dynamic elements** - create stable selectors for these

---

## Dependencies

- None (this is a foundational task)

---

## Related Tasks

- 02-auth-login.md - Uses these naming conventions
- 03-auth-register.md - Uses these naming conventions
- All subsequent test tasks depend on this foundation

---

*Task created from: docs/frontend/TEST_PATHS.md Section 1*
