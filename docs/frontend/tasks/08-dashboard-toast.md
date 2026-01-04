# Task: Implement Toast Notification Tests

> **Task ID**: 08-dashboard-toast  
> **Source**: TEST_PATHS.md Section 3.3  
> **Priority**: High (Core Feature)  
> **Estimated Effort**: 6-8 hours  
> **Prerequisite**: None (standalone component)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Toast Notification System. This task covers success, error, and warning toasts with various interactions and animations.

### Context

The Toast Notification System provides feedback to users for various actions. Testing this system ensures:
- Users receive clear feedback for their actions
- Success, error, and warning messages are distinguished
- Toast auto-dismissal works correctly
- Manual dismissal works properly
- Multiple toasts are handled correctly

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 3.3:

#### Toast Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `toast-container` | container | Toast container |
| `toast-item` | toast | Individual toast |
| `btn-toast-close` | button | Close toast |
| `toast-progress` | progress | Auto-dismiss progress bar |

#### Toast Notification Flow

```
Toast Notification Flow
├── Success Toast
│   ├── Trigger: Successful operation
│   ├── Show: Green toast
│   ├── Icon: Checkmark
│   ├── Message: Operation completed
│   ├── Duration: 5 seconds
│   ├── Show: Progress bar (shrinking)
│   └── Auto-dismiss: After duration
│
├── Error Toast
│   ├── Trigger: Failed operation
│   ├── Show: Red toast
│   ├── Icon: X mark
│   ├── Message: Error description
│   ├── Duration: 8 seconds (longer)
│   └── Auto-dismiss: After duration
│
├── Warning Toast
│   ├── Trigger: Warning condition
│   ├── Show: Yellow toast
│   ├── Icon: Warning triangle
│   └── Message: Warning text
│
├── Dismiss toast manually
│   ├── Hover: Toast
│   ├── Click: btn-toast-close
│   └── Assert: Toast removed from DOM
│
├── Multiple toasts
│   ├── Stack: Vertical (newest on top)
│   ├── Max: 5 visible
│   ├── Older: Dismissed when max exceeded
│   └── Animation: Slide in/out
│
└── Toast interaction
    ├── Click: Toast body (if link)
    │   └── Navigate: Related page
    └── Hover: Pause auto-dismiss timer
```

### Implementation Requirements

1. **Create Toast Test File** (`frontend/e2e/toast.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Toast Page Object** (`frontend/e2e/pages/toast.page.ts`)
   - Encapsulate toast interactions
   - Toast assertion methods
   - Container management

3. **Create Toast Trigger Utilities** (`frontend/e2e/utils/toast-trigger.ts`)
   - Helper functions to trigger different toast types
   - Toast data fixtures

### Expected Deliverables

1. `frontend/e2e/toast.spec.ts` - Comprehensive toast tests
2. `frontend/e2e/pages/toast.page.ts` - Toast page object
3. `frontend/e2e/utils/toast-trigger.ts` - Toast trigger utilities
4. Updated existing tests to use toast assertions

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [ ] Create `frontend/e2e/pages/toast.page.ts`:
  - Toast container locator
  - Individual toast locators
  - Toast close button locators
  - Progress bar locators
- [ ] Create `frontend/e2e/utils/toast-trigger.ts`:
  - Helper functions to trigger toasts
  - Toast data fixtures
  - Custom toast events

### Phase 2: Success Toast Tests

- [ ] Test: Success toast appears on success
- [ ] Test: Success toast has green color
- [ ] Test: Success toast has checkmark icon
- [ ] Test: Success toast shows message
- [ ] Test: Success toast duration (5 seconds)
- [ ] Test: Progress bar appears and shrinks
- [ ] Test: Success toast auto-dismisses
- [ ] Test: Success toast not in DOM after dismiss

### Phase 3: Error Toast Tests

- [ ] Test: Error toast appears on error
- [ ] Test: Error toast has red color
- [ ] Test: Error toast has X icon
- [ ] Test: Error toast shows error message
- [ ] Test: Error toast duration (8 seconds)
- [ ] Test: Error toast progress bar
- [ ] Test: Error toast auto-dismisses
- [ ] Test: Error toast longer duration than success

### Phase 4: Warning Toast Tests

- [ ] Test: Warning toast appears on warning
- [ ] Test: Warning toast has yellow color
- [ ] Test: Warning toast has warning icon
- [ ] Test: Warning toast shows warning message
- [ ] Test: Warning toast default duration
- [ ] Test: Warning toast auto-dismisses

### Phase 5: Manual Dismiss Tests

- [ ] Test: Close button visible on toast
- [ ] Test: Hover shows close button
- [ ] Test: Click close button dismisses toast
- [ ] Test: Toast removed from DOM after close
- [ ] Test: Close button click stops auto-dismiss
- [ ] Test: Click outside toast doesn't dismiss
- [ ] Test: Escape key dismisses toast (if implemented)

### Phase 6: Multiple Toast Tests

- [ ] Test: Multiple toasts stack vertically
- [ ] Test: Newest toast on top
- [ ] Test: Maximum 5 toasts visible
- [ ] Test: Oldest dismissed when max exceeded
- [ ] Test: Toast slide-in animation
- [ ] Test: Toast slide-out animation
- [ ] Test: Toast order maintained

### Phase 7: Interaction Tests

- [ ] Test: Hover pauses auto-dismiss timer
- [ ] Test: Click on toast navigates (if link)
- [ ] Test: Toast hover state
- [ ] Test: Toast accessible by keyboard

### Phase 8: Animation and Timing Tests

- [ ] Test: Toast appears with animation
- [ ] Test: Toast dismisses with animation
- [ ] Test: Progress bar animation
- [ ] Test: Multiple toast staggered appearance

---

## Technical Notes

### Toast Page Object

```typescript
// frontend/e2e/pages/toast.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class ToastPage {
  readonly page: Page;
  readonly container: Locator;
  readonly toasts: Locator;
  readonly toastItem: Locator;
  readonly closeButton: Locator;
  readonly progressBar: Locator;
  readonly successToast: Locator;
  readonly errorToast: Locator;
  readonly warningToast: Locator;

  constructor(page: Page) {
    this.page = page;
    this.container = page.locator('[data-testid="toast-container"]');
    this.toasts = page.locator('[data-testid="toast-item"]');
    this.toastItem = page.locator('[data-testid="toast-item"]');
    this.closeButton = page.locator('[data-testid="btn-toast-close"]');
    this.progressBar = page.locator('[data-testid="toast-progress"]');
    this.successToast = page.locator('[data-testid="toast-success"]');
    this.errorToast = page.locator('[data-testid="toast-error"]');
    this.warningToast = page.locator('[data-testid="toast-warning"]');
  }

  async expectVisible() {
    await expect(this.container).toBeVisible();
  }

  async expectHidden() {
    await expect(this.container).toBeHidden();
  }

  async expectToastCount(count: number) {
    await expect(this.toasts).toHaveCount(count);
  }

  async expectLatestToast(type: 'success' | 'error' | 'warning') {
    const toast = this.toasts.first();
    await expect(toast).toHaveClass(new RegExp(type));
  }

  async expectSuccessToast(message: string) {
    await expect(this.successToast.locator('[data-testid="toast-message"]')).toHaveText(message);
  }

  async expectErrorToast(message: string) {
    await expect(this.errorToast.locator('[data-testid="toast-message"]')).toHaveText(message);
  }

  async expectWarningToast(message: string) {
    await expect(this.warningToast.locator('[data-testid="toast-message"]')).toHaveText(message);
  }

  async clickCloseOnToast(index: number = 0) {
    await this.toasts.nth(index).locator('[data-testid="btn-toast-close"]').click();
  }

  async hoverToast(index: number = 0) {
    await this.toasts.nth(index).hover();
  }

  async expectProgressBarVisible(index: number = 0) {
    await expect(this.toasts.nth(index).locator('[data-testid="toast-progress"]')).toBeVisible();
  }

  async expectProgressBarHidden(index: number = 0) {
    await expect(this.toasts.nth(index).locator('[data-testid="toast-progress"]')).toBeHidden();
  }

  async getToastMessage(index: number = 0): Promise<string> {
    return this.toasts.nth(index).locator('[data-testid="toast-message"]').textContent();
  }

  async dismissAllToasts() {
    const count = await this.toasts.count();
    for (let i = 0; i < count; i++) {
      await this.clickCloseOnToast(0);
      await this.page.waitForTimeout(100);
    }
  }
}
```

### Toast Trigger Utilities

```typescript
// frontend/e2e/utils/toast-trigger.ts
import { Page } from '@playwright/test';

export interface ToastConfig {
  type: 'success' | 'error' | 'warning';
  message: string;
  duration?: number;
  link?: string;
}

// Helper function to trigger toast via custom event
export async function triggerToast(page: Page, config: ToastConfig) {
  await page.evaluate(({ type, message, duration, link }) => {
    const event = new CustomEvent('show-toast', {
      detail: { type, message, duration, link },
    });
    window.dispatchEvent(event);
  }, config);
}

// Helper to trigger success toast
export async function triggerSuccessToast(page: Page, message: string, duration: number = 5000) {
  await triggerToast(page, { type: 'success', message, duration });
}

// Helper to trigger error toast
export async function triggerErrorToast(page: Page, message: string, duration: number = 8000) {
  await triggerToast(page, { type: 'error', message, duration });
}

// Helper to trigger warning toast
export async function triggerWarningToast(page: Page, message: string, duration: number = 5000) {
  await triggerToast(page, { type: 'warning', message, duration });
}

// Trigger multiple toasts
export async function triggerMultipleToasts(page: Page, count: number) {
  for (let i = 0; i < count; i++) {
    await triggerSuccessToast(page, `Toast ${i + 1}`);
    await page.waitForTimeout(100);
  }
}

// Clear all toasts
export async function clearAllToasts(page: Page) {
  await page.evaluate(() => {
    const event = new CustomEvent('clear-all-toasts');
    window.dispatchEvent(event);
  });
}

// Toast data fixtures
export const toastMessages = {
  success: [
    'Chatbot created successfully',
    'Changes saved',
    'Source uploaded successfully',
    'Settings updated',
    'User invited successfully',
  ],
  error: [
    'Failed to create chatbot',
    'An error occurred',
    'Network error',
    'Authentication failed',
    'Upload failed',
  ],
  warning: [
    'Session expiring soon',
    'Storage space low',
    'Rate limit approaching',
    'Password expiring',
  ],
};
```

### Toast Test Fixtures

```typescript
// frontend/e2e/fixtures/toast.fixture.ts
import { test as base } from '@playwright/test';

export const test = base.extend({
  toastPage: async ({ page }, use) => {
    const toastPage = new ToastPage(page);
    await use(toastPage);
  },

  // Helper to trigger and wait for toast
  withToast: async ({ page }, use) => {
    const triggerToast = async (config: { type: string; message: string }) => {
      await page.evaluate(({ type, message }) => {
        const event = new CustomEvent('show-toast', { detail: { type, message, duration: 5000 } });
        window.dispatchEvent(event);
      }, config);
    };
    
    await use(triggerToast);
  },
});
```

### Animation Testing

```typescript
// Test for toast animation
test('toast appears with animation', async ({ page }) => {
  // Trigger toast
  await page.evaluate(() => {
    const event = new CustomEvent('show-toast', { detail: { type: 'success', message: 'Test' } });
    window.dispatchEvent(event);
  });

  // Check for animation class
  const toast = page.locator('[data-testid="toast-item"]').first();
  await expect(toast).toHaveClass(/animate-slide-in/);
});

// Test progress bar animation
test('progress bar shrinks during toast duration', async ({ page }) => {
  await page.evaluate(() => {
    const event = new CustomEvent('show-toast', { detail: { type: 'success', message: 'Test', duration: 1000 } });
    window.dispatchEvent(event);
  });

  const progressBar = page.locator('[data-testid="toast-progress"]').first();
  
  // Initial width should be 100%
  await expect(progressBar).toHaveCSS('width', '100%');
  
  // Wait for progress
  await page.waitForTimeout(500);
  
  // Width should have decreased
  const width = await progressBar.evaluate(el => parseFloat(getComputedStyle(el).width));
  const containerWidth = await progressBar.evaluate(el => el.parentElement?.clientWidth || 100);
  expect(width).toBeLessThan(containerWidth);
});
```

### Running Specific Tests

```bash
# Run all toast tests
cd frontend && npx playwright test toast.spec.ts

# Run success toast tests
cd frontend && npx playwright test toast.spec.ts -g "success"

# Run error toast tests
cd frontend && npx playwright test toast.spec.ts -g "error"

# Run multiple toast tests
cd frontend && npx playwright test toast.spec.ts -g "multiple"

# Run in headed mode
cd frontend && npx playwright test toast.spec.ts --headed
```

---

## Verification Steps

### 1. Test Coverage Verification
- [ ] All toast types tested
- [ ] All interactions tested
- [ ] All animations tested
- [ ] Timing tests passed
- [ ] Multiple toasts tested

### 2. Test Execution Verification
- [ ] All tests pass locally
- [ ] No flaky tests
- [ ] Proper timing tolerance
- [ ] Clean test isolation

### 3. UX Verification
- [ ] Clear visual distinction between types
- [ ] Progress bar visible
- [ ] Close button accessible
- [ ] Animations smooth

### 4. Accessibility Verification
- [ ] Keyboard accessible
- [ ] ARIA live regions
- [ ] Screen reader announcements
- [ ] Focus management

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Timing** - Toast tests need careful timing handling
2. **Auto-dismiss** - Tests may need longer timeouts
3. **Custom Events** - Use custom events to trigger toasts in tests
4. **Animation** - Account for animation duration in tests

### Common Issues to Avoid

1. **Race conditions** - Wait for toast to appear before interacting
2. **Hardcoded messages** - Use fixtures for consistency
3. **Skipping animation tests** - Test animations when possible
4. **Not cleaning up** - Always dismiss toasts between tests

### Timing Considerations

```typescript
// Use longer timeouts for toast tests
test.setTimeout(15000);

// Wait for auto-dismiss
test('toast auto-dismisses after duration', async ({ page }) => {
  await triggerSuccessToast(page, 'Test message', 1000); // 1 second
  
  // Toast should be visible
  await expect(page.locator('[data-testid="toast-item"]')).toBeVisible();
  
  // Wait for auto-dismiss
  await page.waitForTimeout(1500);
  
  // Toast should be gone
  await expect(page.locator('[data-testid="toast-item"]')).toBeHidden();
});
```

---

## Dependencies

- **Prerequisites**: None (can be done independently)
- **Environment**: Frontend with toast component
- **Test Data**: Various toast messages

---

## Related Tasks

- 02-auth-login.md - Toast validation errors
- 03-auth-register.md - Toast registration feedback
- 04-auth-logout.md - Toast logout confirmation
- All tasks that use toast notifications

---

*Task created from: docs/frontend/TEST_PATHS.md Section 3.3*
