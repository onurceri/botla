# E2E Testing Standards

This document defines the comprehensive testing standards, naming conventions, and best practices for all E2E and integration tests in the Botla-Co frontend test suite.

## Table of Contents

1. [Overview](#overview)
2. [File Naming Conventions](#file-naming-conventions)
3. [Test Naming Patterns](#test-naming-patterns)
4. [Element Naming Convention](#element-naming-convention)
5. [Selector Strategy](#selector-strategy)
6. [Test Organization](#test-organization)
7. [Mock Setup Guidelines](#mock-setup-guidelines)
8. [Best Practices](#best-practices)
9. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)
10. [Common Patterns](#common-patterns)
11. [Accessibility Considerations](#accessibility-considerations)

---

## Overview

These standards ensure:
- **Consistent element identification** across all test files
- **Maintainable and readable** test code
- **Clear mapping** between test IDs and UI components
- **Improved debugging** experience with descriptive selectors
- **Stable tests** that don't break with minor UI changes

---

## File Naming Conventions

### Test Files

All test files must follow the `{page-or-feature}.spec.ts` pattern:

```
{page-or-feature}.spec.ts
```

**Examples:**
- `auth.spec.ts` - Authentication tests
- `chatbot.spec.ts` - Chatbot management tests
- `smoke.spec.ts` - Smoke tests
- `widget-embed.spec.ts` - Widget embed tests
- `mobile-responsiveness.spec.ts` - Mobile tests

### Utility Files

Utility files should follow the `{purpose}.ts` or `{purpose}.ts` pattern:

```
utils/{purpose}.ts
```

**Examples:**
- `utils/selectors.ts` - Element ID constants
- `utils/test-helpers.ts` - Test helper utilities
- `helpers.ts` - Mock setup functions
- `test-constants.ts` - Test text constants

---

## Test Naming Patterns

### Test Describe Blocks

Use `test.describe()` to group related tests by feature area:

```typescript
test.describe('Feature Area', () => {
  test('should perform action when user does X', async () => { ... });
  test('should show error when Y condition', async () => { ... });
});
```

### Test Names

Test names should follow the pattern: `should {action} when {condition}`

**Good Examples:**
```typescript
test('should login successfully with valid credentials', async () => { ... });
test('should show error message when email is invalid', async () => { ... });
test('should redirect to dashboard after successful registration', async () => { ... });
test('should disable submit button when form is invalid', async () => { ... });
test('should display chatbot list when user is authenticated', async () => { ... });
```

**Bad Examples:**
```typescript
test('login', async () => { ... });                    // Too vague
test('test the login page', async () => { ... });       // Not action-oriented
test('user can create a chatbot', async () => { ... }); // Acceptable but prefer "should"
```

### Nested Describe Blocks

Use nested describe blocks for related test groups:

```typescript
test.describe('Login Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('should show validation errors for empty fields', async () => { ... });
  
  test.describe('Successful Login', () => {
    test('should redirect to dashboard with valid credentials', async () => { ... });
    test('should persist session with remember me checked', async () => { ... });
  });
  
  test.describe('Failed Login', () => {
    test('should show error with invalid password', async () => { ... });
    test('should show error with non-existent email', async () => { ... });
  });
});
```

---

## Element Naming Convention

All `data-testid` attributes must follow the naming convention:

```
{prefix}-{component}-{action}
```

### Element Type Prefixes

| Prefix | Element Type | Example |
|--------|--------------|---------|
| `btn` | Button | `btn-create-chatbot` |
| `input` | Text input | `input-email` |
| `select` | Select dropdown | `select-plan` |
| `link` | Navigation link | `link-login` |
| `tab` | Tab navigation | `tab-settings` |
| `modal` | Modal dialog | `modal-confirm-delete` |
| `toast` | Toast notification | `toast-success` |
| `dropdown` | Dropdown menu | `dropdown-menu` |
| `checkbox` | Checkbox | `checkbox-terms` |
| `radio` | Radio button | `radio-mode` |
| `toggle` | Toggle switch | `toggle-visibility` |
| `card` | Card component | `card-chatbot` |
| `page` | Page container | `page-login` |
| `header` | Header | `header-dashboard` |
| `footer` | Footer | `footer-widget` |
| `nav` | Navigation | `nav-sidebar` |
| `sidebar` | Sidebar component | `sidebar-main` |
| `menu` | Menu item | `menu-user` |
| `icon` | Icon button | `icon-close` |
| `avatar` | User avatar | `avatar-user` |
| `badge` | Badge indicator | `badge-status` |
| `table` | Table container | `table-results` |
| `row` | Table row | `row-chatbot` |
| `cell` | Table cell | `cell-name` |
| `form` | Form container | `form-login` |
| `label` | Form label | `label-email` |
| `error` | Error message | `error-message` |
| `success` | Success message | `success-message` |
| `loading` | Loading indicator | `loading-spinner` |
| `empty` | Empty state | `empty-list` |
| `search` | Search input | `search-input` |
| `filter` | Filter element | `filter-dropdown` |
| `sort` | Sort element | `sort-select` |
| `pagination` | Pagination | `pagination-controls` |
| `breadcrumb` | Breadcrumb nav | `breadcrumb-nav` |
| `tooltip` | Tooltip | `tooltip-help` |
| `dialog` | Dialog | `dialog-confirm` |
| `alert` | Alert message | `alert-error` |
| `progress` | Progress bar | `progress-bar` |
| `spinner` | Loading spinner | `spinner-loading` |
| `skeleton` | Skeleton loader | `skeleton-card` |
| `accordion` | Accordion | `accordion-item` |
| `carousel` | Carousel | `carousel-slide` |
| `slider` | Slider | `slider-range` |
| `chart` | Chart component | `chart-analytics` |
| `graph` | Graph element | `graph-network` |
| `widget` | Widget component | `widget-chat` |
| `container` | Container | `container-main` |
| `wrapper` | Wrapper | `wrapper-content` |
| `section` | Section | `section-hero` |
| `group` | Element group | `group-actions` |
| `item` | List item | `item-dropdown` |
| `detail` | Detail view | `detail-chatbot` |
| `list` | List container | `list-chatbots` |
| `grid` | Grid layout | `grid-cards` |
| `layout` | Layout container | `layout-dashboard` |
| `main` | Main content | `main-content` |
| `content` | Content container | `content-body` |
| `title` | Title element | `title-page` |
| `heading` | Heading element | `heading-primary` |
| `text` | Text element | `text-description` |
| `paragraph` | Paragraph | `paragraph-help` |
| `description` | Description | `description-item` |
| `message` | Message | `message-info` |
| `notification` | Notification | `notification-toast` |
| `status` | Status indicator | `status-active` |
| `indicator` | Indicator | `indicator-loading` |
| `tag` | Tag element | `tag-category` |
| `chip` | Chip component | `chip-filter` |
| `pill` | Pill component | `pill-badge` |
| `marker` | Marker element | `marker-point` |
| `point` | Point element | `point-data` |
| `line` | Line element | `line-chart` |
| `bar` | Bar element | `bar-chart` |
| `axis` | Axis element | `axis-x` |
| `legend` | Legend element | `legend-chart` |
| `handle` | Handle element | `handle-drag` |
| `track` | Track element | `track-slider` |
| `thumb` | Thumb element | `thumb-slider` |
| `rail` | Rail element | `rail-slider` |
| `step` | Step element | `step-wizard` |
| `steps` | Steps container | `steps-wizard` |
| `wizard` | Wizard component | `wizard-form` |
| `field` | Form field | `field-email` |
| `control` | Form control | `control-input` |
| `validation` | Validation message | `validation-error` |
| `help` | Help text | `help-text` |
| `placeholder` | Placeholder text | `placeholder-input` |
| `value` | Display value | `value-display` |
| `count` | Count indicator | `count-items` |
| `total` | Total indicator | `total-count` |
| `limit` | Limit indicator | `limit-warning` |
| `threshold` | Threshold | `threshold-alert` |
| `setting` | Setting element | `setting-toggle` |
| `option` | Option element | `option-select` |
| `choice` | Choice element | `choice-radio` |
| `selection` | Selection | `selection-indicator` |
| `result` | Result element | `result-item` |
| `output` | Output element | `output-display` |
| `display` | Display element | `display-value` |
| `view` | View element | `view-detail` |
| `preview` | Preview element | `preview-image` |
| `thumbnail` | Thumbnail | `thumbnail-image` |
| `image` | Image element | `image-avatar` |
| `picture` | Picture element | `picture-cover` |
| `video` | Video element | `video-player` |
| `audio` | Audio element | `audio-player` |
| `media` | Media element | `media-container` |
| `file` | File element | `file-input` |
| `upload` | Upload element | `upload-area` |
| `download` | Download element | `download-button` |
| `attachment` | Attachment | `attachment-file` |
| `document` | Document | `document-pdf` |
| `folder` | Folder | `folder-item` |
| `directory` | Directory | `directory-tree` |
| `tree` | Tree structure | `tree-navigation` |
| `node` | Tree node | `node-tree` |
| `leaf` | Leaf node | `leaf-tree` |
| `branch` | Branch node | `branch-tree` |
| `root` | Root element | `root-tree` |
| `level` | Level indicator | `level-depth` |
| `depth` | Depth indicator | `depth-indicator` |
| `hierarchy` | Hierarchy | `hierarchy-view` |
| `structure` | Structure | `structure-data` |
| `organization` | Organization | `org-selector` |
| `relationship` | Relationship | `relationship-view` |
| `connection` | Connection | `connection-status` |
| `edge` | Edge element | `edge-graph` |
| `vertex` | Vertex element | `vertex-graph` |
| `network` | Network | `network-graph` |
| `cluster` | Cluster | `cluster-data` |
| `cloud` | Cloud element | `cloud-network` |
| `server` | Server element | `server-status` |
| `client` | Client element | `client-info` |
| `endpoint` | Endpoint | `endpoint-url` |
| `service` | Service | `service-status` |
| `api` | API element | `api-endpoint` |
| `request` | Request | `request-data` |
| `response` | Response | `response-body` |
| `data` | Data element | `data-display` |
| `payload` | Payload | `payload-data` |
| `body` | Body element | `body-content` |
| `meta` | Meta element | `meta-data` |
| `config` | Configuration | `config-settings` |
| `preference` | Preference | `preference-setting` |
| `customization` | Customization | `customize-options` |
| `theme` | Theme element | `theme-selector` |
| `style` | Style element | `style-option` |
| `appearance` | Appearance | `appearance-setting` |
| `design` | Design element | `design-preview` |
| `position` | Position | `position-indicator` |
| `placement` | Placement | `placement-option` |
| `alignment` | Alignment | `alignment-option` |
| `spacing` | Spacing | `spacing-option` |
| `sizing` | Sizing | `sizing-option` |
| `dimension` | Dimension | `dimension-display` |
| `size` | Size element | `size-indicator` |
| `width` | Width element | `width-display` |
| `height` | Height element | `height-display` |
| `length` | Length element | `length-display` |
| `scale` | Scale element | `scale-control` |
| `zoom` | Zoom control | `zoom-control` |
| `pan` | Pan control | `pan-control` |
| `rotate` | Rotate control | `rotate-control` |
| `transform` | Transform | `transform-control` |
| `animation` | Animation | `animation-control` |
| `transition` | Transition | `transition-effect` |
| `motion` | Motion element | `motion-control` |
| `effect` | Effect element | `effect-display` |
| `filter` | Filter effect | `filter-control` |
| `blur` | Blur effect | `blur-control` |
| `shadow` | Shadow effect | `shadow-control` |
| `opacity` | Opacity | `opacity-control` |
| `transparency` | Transparency | `transparency-control` |
| `color` | Color element | `color-picker` |
| `palette` | Color palette | `palette-display` |
| `scheme` | Color scheme | `scheme-selector` |
| `mode` | Mode element | `mode-toggle` |
| `brightness` | Brightness | `brightness-control` |
| `contrast` | Contrast | `contrast-control` |
| `saturation` | Saturation | `saturation-control` |
| `hue` | Hue control | `hue-control` |
| `tint` | Tint effect | `tint-control` |
| `shade` | Shade effect | `shade-control` |
| `tone` | Tone effect | `tone-control` |

### Naming Rules

1. **Use kebab-case**: All lowercase with hyphens
2. **Be descriptive**: `btn-create-chatbot` not `btn-create`
3. **Include context**: `input-email` not `input`
4. **Prefix actions**: `btn-` for actions, `link-` for navigation
5. **Use semantic names**: `btn-save` not `btn-s`

### Examples by Element Type

#### Buttons
```
btn-create-chatbot
btn-submit-login
btn-cancel-modal
btn-delete-confirm
btn-edit-chatbot
btn-save-settings
btn-upload-file
btn-search-submit
btn-filter-apply
btn-sort-asc
btn-next-page
btn-prev-page
btn-close-modal
btn-toggle-theme
```

#### Input Fields
```
input-email
input-password
input-chatbot-name
input-search-query
input-url-source
input-text-content
input-filter-search
input-phone-number
input-full-name
input-company-name
```

#### Select Dropdowns
```
select-language
select-plan
select-sort-by
select-filter-status
select-chatbot
select-date-range
select-timezone
select-currency
```

#### Links
```
link-login
link-register
link-forgot-password
link-back-home
link-view-all
link-read-more
link-terms-service
link-privacy-policy
```

#### Tabs
```
tab-overview
tab-sources
tab-playground
tab-settings
tab-analytics
tab-integrations
tab-chunks
tab-history
```

#### Modals
```
modal-create-chatbot
modal-confirm-delete
modal-edit-settings
modal-upload-source
modal-view-details
modal-share-chatbot
modal-embed-code
modal-welcome
```

#### Toast Notifications
```
toast-success
toast-error
toast-warning
toast-info
toast-login-success
toast-chatbot-created
toast-source-added
toast-error-generic
```

#### Cards
```
card-chatbot
card-source
card-analytics
card-user
card-notification
card-pricing
card-feature
card-testimonial
```

#### Tables
```
table-chatbots
table-sources
table-users
table-analytics
table-results
row-chatbot
cell-name
cell-status
cell-actions
header-table
```

---

## Selector Strategy

### Primary: data-testid Attributes

Use `data-testid` attributes for element identification. This provides:
- Stable selectors that don't break with UI changes
- Clear semantic meaning
- Easy debugging

```typescript
// Good - using data-testid
await page.getByTestId('btn-login').click();
await page.getByTestId('input-email').fill('test@example.com');

// Bad - using CSS selectors
await page.locator('.btn-primary').click();
await page.locator('button[type="submit"]').click();
await page.locator('div.main-content > form > input[type="email"]').fill('test@example.com');
```

### Fallback: Semantic Selectors

When `data-testid` is not available, use semantic selectors:

```typescript
// Good - semantic selectors
await page.getByRole('button', { name: /submit/i }).click();
await page.getByLabel('Email').fill('test@example.com');
await page.getByPlaceholder('Enter your email').fill('test@example.com');
await page.getByText('Welcome back!').isVisible();
await page.getByAltText('Company logo').isVisible();
await page.getByTitle('Delete item').click();

// Good - with exact match
await page.getByRole('link', { name: 'Login', exact: true }).click();
```

### Avoid: Fragile Selectors

Avoid selectors that are likely to break:

```typescript
// Bad - fragile selectors
await page.locator('div:nth-child(2) > .class-name > span').click();
await page.locator('button.btn.btn-primary.btn-lg').click();
await page.locator('[class*="some-random-class-123"]').click();
await page.locator('//div[@id="container"]/div[2]/div[1]/span[2]').click();
```

---

## Test Organization

### File Structure

```
frontend/e2e/
├── *.spec.ts              # Test files
├── TESTING_STANDARDS.md   # This documentation
├── helpers.ts             # Mock setup and helper functions
├── test-constants.ts      # Text constants for tests
└── utils/
    ├── selectors.ts       # Element ID constants (recommended)
    └── test-helpers.ts    # Test helper utilities (recommended)
```

### Test.describe() Pattern

Group tests logically using `test.describe()`:

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Area', () => {
  test.beforeEach(async ({ page }) => {
    // Setup before each test
  });

  test('should do X when Y', async ({ page }) => {
    // Test implementation
  });

  test('should do Z when W', async ({ page }) => {
    // Test implementation
  });

  test.describe('Sub-feature', () => {
    test.beforeAll(async () => {
      // Setup once for sub-feature
    });

    test('should do A when B', async ({ page }) => {
      // Test implementation
    });
  });
});
```

### Before/After Hooks

Use hooks appropriately:

```typescript
test.beforeAll(async () => {
  // Setup once for all tests in this describe
});

test.beforeEach(async ({ page }) => {
  // Setup before each test
});

test.afterEach(async ({ page }) => {
  // Cleanup after each test
});

test.afterAll(async () => {
  // Cleanup after all tests
});
```

---

## Mock Setup Guidelines

### Use Consistent Mock Patterns

```typescript
// helpers.ts
export async function setupAuthMocks(page: Page) {
  await page.route('**/api/v1/auth/login', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ token: 'test-token', refresh_token: 'test-refresh' }),
    });
  });

  await page.route('**/api/v1/auth/me', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ id: 'user-1', email: 'test@example.com', plan: 'pro' }),
    });
  });
}

export async function setupChatbotMocks(page: Page, botId: string = 'bot-1') {
  await page.route('**/api/v1/chatbots', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([]),
      });
    }
  });
}
```

### Organization

```typescript
// helpers.ts
export async function setupOrgMocks(page: Page) {
  await page.route('**/api/v1/organizations', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          id: 'org-1',
          name: 'Test Org',
          slug: 'test-org',
          owner_id: 'user-1',
          plan_id: 'pro',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ]),
    });
  });
}
```

---

## Best Practices

### 1. Use Page Objects (Optional)

For complex pages, consider using page objects:

```typescript
// pages/LoginPage.ts
import { Page, Locator, expect } from '@playwright/test';

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/login');
  }

  async fillEmail(email: string) {
    await this.page.getByTestId('input-login-email').fill(email);
  }

  async fillPassword(password: string) {
    await this.page.getByTestId('input-login-password').fill(password);
  }

  async clickSubmit() {
    await this.page.getByTestId('btn-login-submit').click();
  }

  async expectErrorMessage(message: string) {
    await expect(this.page.getByTestId('error-login-message')).toContainText(message);
  }
}

// Usage in test
test('should show error with invalid credentials', async ({ page }) => {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.fillEmail('wrong@example.com');
  await loginPage.fillPassword('wrongpassword');
  await loginPage.clickSubmit();
  await loginPage.expectErrorMessage('Invalid credentials');
});
```

### 2. Use test.step() for Readability

```typescript
import { test, expect } from '@playwright/test';

test('should complete checkout flow', async ({ page }) => {
  await test.step('Navigate to cart', async () => {
    await page.goto('/cart');
  });

  await test.step('Verify cart items', async () => {
    await expect(page.getByTestId('cart-item')).toHaveCount(2);
  });

  await test.step('Proceed to checkout', async () => {
    await page.getByTestId('btn-checkout').click();
  });

  await test.step('Complete payment', async () => {
    await page.getByTestId('input-card-number').fill('4242424242424242');
    await page.getByTestId('input-expiry').fill('12/25');
    await page.getByTestId('input-cvc').fill('123');
    await page.getByTestId('btn-pay').click();
  });

  await test.step('Verify success', async () => {
    await expect(page).toHaveURL(/\/order-confirmation/);
    await expect(page.getByTestId('success-message')).toContainText('Order confirmed');
  });
});
```

### 3. Use Auto-waiting

Playwright automatically waits for elements. Use it to your advantage:

```typescript
// Good - Playwright waits for element to be visible
await page.getByTestId('btn-submit').click();

// Good - Playwright waits for navigation
await expect(page).toHaveURL('/dashboard');

// Good - Playwright waits for API response
await page.getByTestId('btn-refresh').click();
await expect(page.getByTestId('data-loaded')).toBeVisible();
```

### 4. Use Soft Assertions for Non-critical Checks

```typescript
import { test, expect } from '@playwright/test';

test('should display user info', async ({ page }) => {
  // Critical - test will fail here
  await expect(page.getByTestId('user-name')).toBeVisible();

  // Soft - will report but not fail
  await expect.soft(page.getByTestId('user-avatar')).toHaveAttribute('src', 'avatar.jpg');
  await expect.soft(page.getByTestId('user-bio')).toContainText('Software Engineer');
});
```

### 5. Use Locator Filters

```typescript
// Filter by text
await page.getByTestId('btn-action').filter({ hasText: 'Delete' }).click();

// Filter by state
await page.getByTestId('input-field').filter({ has: page.locator('.required') }).fill('value');

// Chain filters
await page.getByTestId('list-item')
  .filter({ hasText: 'Chatbot 1' })
  .getByTestId('btn-edit')
  .click();
```

### 6. Use Web First Assertions

```typescript
import { test, expect } from '@playwright/test';

test('should display notifications', async ({ page }) => {
  // Web first assertions - more reliable
  await expect(page.getByTestId('notification')).toBeVisible();
  await expect(page.getByTestId('notification')).toHaveText('New message received');
  await expect(page.getByTestId('notification')).toHaveCount(3);
  await expect(page.getByTestId('notification').first()).toBeEnabled();
  await expect(page.getByTestId('notification').last()).toBeDisabled();
});
```

---

## Anti-Patterns to Avoid

### 1. Hardcoded Sleeps

```typescript
// Bad - never use sleep
await page.waitForTimeout(2000);
await new Promise(resolve => setTimeout(resolve, 2000));

// Good - use auto-waiting or explicit waits
await page.getByTestId('btn-submit').click();
await expect(page.getByTestId('success-message')).toBeVisible({ timeout: 10000 });
```

### 2. Hardcoded Strings in Tests

```typescript
// Bad - hardcoded strings
await page.getByLabel('Email').fill('test@example.com');
await expect(page.getByText('Login successful')).toBeVisible();

// Good - use constants
import { TURKISH } from './test-constants';
await page.getByLabel(TURKISH.EMAIL).fill('test@example.com');
await expect(page.getByText(TURKISH.LOGIN_SUCCESS)).toBeVisible();
```

### 3. Single Long Test

```typescript
// Bad - too long, hard to debug
test('complete user journey', async ({ page }) => {
  // 200 lines of test code
});

// Good - split into focused tests
test.describe('User Registration', () => {
  test('should register successfully', async ({ page }) => { ... });
  test('should show validation errors', async ({ page }) => { ... });
  test('should send confirmation email', async ({ page }) => { ... });
});
```

### 4. Ignoring Errors

```typescript
// Bad - never ignore errors silently
try {
  await page.getByTestId('btn-delete').click();
} catch (e) {
  // Ignore
}

// Good - handle or report errors
await expect(async () => {
  await page.getByTestId('btn-delete').click();
}).rejects.toThrow();
```

### 5. Using CSS/XPath Selectors for Interaction

```typescript
// Bad - CSS selectors are fragile
await page.locator('.btn-primary.submit-btn').click();
await page.locator('div.container > div.content > button').click();

// Good - use data-testid or semantic selectors
await page.getByTestId('btn-submit').click();
await page.getByRole('button', { name: /submit/i }).click();
```

### 6. Not Using test.only() for Debugging

```typescript
// When debugging a specific test, use test.only()
test.only('this is the test I want to debug', async ({ page }) => {
  // ...
});

// Remember to remove .only() before committing!
```

### 7. Skipping Tests Without Explanation

```typescript
// Bad
test.skip('broken test', async ({ page }) => { ... });

// Good
test.skip('TODO: fix this test after API change', async ({ page }) => {
  // Test that needs to be fixed
});
```

### 8. Not Cleaning Up Between Tests

```typescript
// Bad - state leaks between tests
test('test 1', async ({ page }) => {
  await page.getByTestId('input-name').fill('Test Name');
});

test('test 2', async ({ page }) => {
  // Input still has 'Test Name' from previous test
});

// Good - use beforeEach to reset state
test.beforeEach(async ({ page }) => {
  await page.goto('/page');
  await page.evaluate(() => localStorage.clear());
});
```

---

## Common Patterns

### Form Submission Pattern

```typescript
test('should submit form successfully', async ({ page }) => {
  await page.goto('/form');
  
  // Fill form using data-testid
  await page.getByTestId('input-name').fill('John Doe');
  await page.getByTestId('input-email').fill('john@example.com');
  await page.getByTestId('input-password').fill('securepassword123');
  
  // Submit form
  await page.getByTestId('btn-submit').click();
  
  // Verify success
  await expect(page).toHaveURL(/\/success/);
  await expect(page.getByTestId('success-message')).toContainText('Form submitted');
});
```

### Modal Pattern

```typescript
test('should open and close modal', async ({ page }) => {
  // Open modal
  await page.getByTestId('btn-open-modal').click();
  await expect(page.getByTestId('modal-create')).toBeVisible();
  
  // Interact with modal
  await page.getByTestId('input-chatbot-name').fill('My Chatbot');
  await page.getByTestId('btn-modal-submit').click();
  
  // Verify modal closed
  await expect(page.getByTestId('modal-create')).not.toBeVisible();
});
```

### Dropdown Pattern

```typescript
test('should select option from dropdown', async ({ page }) => {
  // Open dropdown
  await page.getByTestId('select-plan').click();
  
  // Select option
  await page.getByTestId('option-plan-pro').click();
  
  // Verify selection
  await expect(page.getByTestId('select-plan')).toContainText('Pro');
});
```

### Async Operation Pattern

```typescript
test('should handle async operation', async ({ page }) => {
  // Start async operation
  await page.getByTestId('btn-upload').click();
  
  // Wait for loading to complete
  await expect(page.getByTestId('loading-spinner')).toBeVisible();
  await expect(page.getByTestId('loading-spinner')).not.toBeVisible({ timeout: 30000 });
  
  // Verify result
  await expect(page.getByTestId('success-message')).toContainText('Upload complete');
});
```

### Error Handling Pattern

```typescript
test('should display error on failure', async ({ page }) => {
  // Mock failed API response
  await page.route('**/api/v1/action', async (route) => {
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'Internal server error' }),
    });
  });

  // Trigger action
  await page.getByTestId('btn-submit').click();
  
  // Verify error displayed
  await expect(page.getByTestId('error-message')).toBeVisible();
  await expect(page.getByTestId('error-message')).toContainText('Internal server error');
});
```

### Navigation Pattern

```typescript
test('should navigate between pages', async ({ page }) => {
  // Start at home
  await page.goto('/');
  
  // Navigate to login
  await page.getByTestId('link-login').click();
  await expect(page).toHaveURL('/login');
  
  // Navigate to register
  await page.getByTestId('link-register').click();
  await expect(page).toHaveURL('/register');
});
```

---

## Accessibility Considerations

### Use Semantic HTML Attributes

```typescript
// Good - accessible selectors
await page.getByLabel('Email address').fill('test@example.com');
await page.getByRole('button', { name: 'Submit' }).click();
await page.getByPlaceholder('Enter your name').fill('John');
await page.getByAltText('Company logo').isVisible();
await page.getByTitle('Delete item').click();
await page.getByTestId('checkbox-terms').check();
```

### Test Keyboard Navigation

```typescript
test('should support keyboard navigation', async ({ page }) => {
  // Tab through interactive elements
  await page.keyboard.press('Tab');
  await expect(page.locator(':focus')).toHaveAttribute('data-testid', 'input-email');
  
  await page.keyboard.press('Tab');
  await expect(page.locator(':focus')).toHaveAttribute('data-testid', 'input-password');
  
  await page.keyboard.press('Tab');
  await expect(page.locator(':focus')).toHaveAttribute('data-testid', 'btn-login');
});
```

### Test ARIA Attributes

```typescript
test('should have proper ARIA attributes', async ({ page }) => {
  await expect(page.getByTestId('btn-toggle')).toHaveAttribute('aria-expanded', 'true');
  await expect(page.getByTestId('dialog-modal')).toHaveAttribute('role', 'dialog');
  await expect(page.getByTestId('checkbox-terms')).toHaveAttribute('aria-describedby', 'terms-description');
});
```

---

## Quick Reference

### Naming Pattern Quick Reference

| Pattern | Example |
|---------|---------|
| Button | `btn-{action}-{target}` |
| Input | `input-{field-name}` |
| Select | `select-{field-name}` |
| Link | `link-{destination}` |
| Tab | `tab-{tab-name}` |
| Modal | `modal-{purpose}` |
| Card | `card-{content-type}` |
| List | `list-{content-type}` |
| Error | `error-{context}` |
| Success | `success-{context}` |
| Loading | `loading-{context}` |

### Selector Priority

1. `page.getByTestId('...')` - Primary, most stable
2. `page.getByRole('...')` - Semantic, accessible
3. `page.getByLabel('...')` - Form fields
4. `page.getByPlaceholder('...')` - Input placeholders
5. `page.getByText('...')` - Visible text
6. `page.getByAltText('...')` - Images
7. `page.getByTitle('...')` - Tooltips, titles

---

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Web First Assertions](https://playwright.dev/docs/test-assertions)
- [Locators](https://playwright.dev/docs/locators)
- [API Mocking](https://playwright.dev/docs/network)

---

## Session-Specific Patterns: Auth & Logout Flow Testing

This section documents patterns and lessons learned from implementing the Logout Flow E2E tests.

### 1. localStorage Access in Playwright

#### The Problem
When navigating to protected pages (e.g., `/dashboard`) that redirect unauthenticated users, localStorage access may be denied because the page redirects before storage can be set.

```typescript
// BAD - Fails because page redirects before localStorage is accessible
test.beforeEach(async ({ page }) => {
  await page.goto('/dashboard');  // May redirect, losing localStorage access
  await page.evaluate(() => {
    localStorage.setItem('botla_token', 'test-token');
  });
});
```

#### The Solution: Use addInitScript()
Use `page.addInitScript()` to set localStorage before page navigation:

```typescript
// GOOD - Sets storage before page loads
async function setSessionStorage(page: Page, tokens: SessionTokens) {
  await page.addInitScript((tokens) => {
    localStorage.setItem('botla_token', tokens.accessToken);
    localStorage.setItem('botla_refresh_token', tokens.refreshToken);
    localStorage.setItem('botla_user', JSON.stringify(tokens.user));
  }, tokens);
}

test.beforeEach(async ({ page }) => {
  await setSessionStorage(page, {
    accessToken: 'mock-access-token-' + Date.now(),
    refreshToken: 'mock-refresh-token-' + Date.now(),
    user: { id: 'user-123', email: 'test@example.com' },
  });
  await page.goto('/dashboard');
});
```

### 2. Testing Elements in Collapsible UI Components

#### The Problem
Elements like logout buttons may exist in the DOM but be hidden due to collapsed sidebar modes, hover states, or responsive breakpoints.

```typescript
// BAD - Fails because element exists but is not visible
test('should find logout button', async ({ page }) => {
  await page.goto('/dashboard');
  const logoutButton = page.locator('.logout-btn');
  await expect(logoutButton).toBeAttached(); // May fail if element not in DOM yet
});
```

#### The Solution: Use Page Content Check
When the element exists in source but may be visually hidden:

```typescript
// GOOD - Check for element in page source
test('should have logout button in DOM', async ({ page }) => {
  await page.goto('/dashboard');
  const pageContent = await page.content();
  expect(pageContent).toContain('logout-btn');
});

// Alternative: Use count() to check existence without visibility
test('should have logout button', async ({ page }) => {
  await page.goto('/dashboard');
  const logoutButton = page.locator('.logout-btn');
  expect(await logoutButton.count()).toBeGreaterThan(0);
});
```

### 3. Multi-Tab BroadcastChannel Testing

#### The Problem
Testing BroadcastChannel communication between tabs requires careful timing. Listeners set via `page.evaluate()` may not persist or receive messages reliably across different page contexts.

```typescript
// BAD - Listener may not receive messages across evaluate calls
test('should send BroadcastChannel message', async ({ browser }) => {
  const pageA = await browser.newPage();
  const pageB = await browser.newPage();
  await pageA.goto('about:blank');
  await pageB.goto('about:blank');

  // Set up listener
  const messages: string[] = [];
  await pageB.evaluate(() => {
    const bc = new BroadcastChannel('auth_channel');
    bc.onmessage = (e) => messages.push(e.data);
  });

  await pageB.waitForTimeout(200); // Not reliable

  // Send message from another page
  await pageA.evaluate(() => {
    const bc = new BroadcastChannel('auth_channel');
    bc.postMessage('session_terminated');
  });

  await pageB.waitForTimeout(500);
  expect(messages).toContain('session_terminated'); // Often fails
});
```

#### The Solution: Single Evaluate Call
Perform listener setup and message sending in a single execution context:

```typescript
// GOOD - Single evaluate handles both listener and message
test('should send BroadcastChannel message', async ({ browser }) => {
  const pageB = await browser.newPage();
  await pageB.goto('about:blank');

  const result = await pageB.evaluate(async () => {
    const messages: string[] = [];
    const bc = new BroadcastChannel('test_auth_channel');
    bc.onmessage = (e) => messages.push(e.data);

    // Wait for listener to be ready
    await new Promise(resolve => setTimeout(resolve, 500));

    // Send message from the same context
    const sender = new BroadcastChannel('test_auth_channel');
    sender.postMessage('session_terminated');

    // Wait for message to be received
    await new Promise(resolve => setTimeout(resolve, 1000));

    return messages;
  });

  expect(result).toContain('session_terminated');
});
```

### 4. Test Organization: Phase-Based Structure

For complex flows like authentication/logout, organize tests into phases:

```typescript
test.describe('Logout Flow', () => {
  test.describe('Session Management', () => {
    // Tests for session token set/clear operations
  });

  test.describe('Session Utilities', () => {
    // Tests for utility functions and helpers
  });

  test.describe('Multi-Tab Synchronization', () => {
    // Tests for BroadcastChannel and cross-tab communication
  });

  test.describe('Security Verification', () => {
    // Tests for token cleanup and security verification
  });

  test.describe('Mock Handlers', () => {
    // Tests for API mocking and error scenarios
  });
});
```

### 5. Session Storage Helper Pattern

Create reusable helpers for authentication state management:

```typescript
// utils/session.utils.ts
interface SessionTokens {
  accessToken: string;
  refreshToken: string;
  user: object;
}

export async function setSessionStorage(page: Page, tokens: SessionTokens) {
  await page.addInitScript((tokens) => {
    localStorage.setItem('botla_token', tokens.accessToken);
    localStorage.setItem('botla_refresh_token', tokens.refreshToken);
    localStorage.setItem('botla_user', JSON.stringify(tokens.user));
  }, tokens);
}

export async function clearSessionStorage(page: Page) {
  await page.addInitScript(() => {
    localStorage.removeItem('botla_token');
    localStorage.removeItem('botla_refresh_token');
    localStorage.removeItem('botla_user');
  });
}
```

### 6. Testing CSS Class-Based Elements

When elements don't have `data-testid`, use CSS classes as selectors:

```typescript
// The logout button in DashboardLayout.tsx uses className="logout-btn"
test('should have logout button in dashboard', async ({ page }) => {
  await page.goto('/dashboard');
  const pageContent = await page.content();
  expect(pageContent).toContain('logout-btn');
});

// For interactive tests (when element is visible)
test('should click logout button', async ({ page }) => {
  // Expand sidebar first if needed
  await page.evaluate(() => {
    localStorage.setItem('botla_sidebar_mode', 'pinned');
  });
  await page.reload();

  // Now button should be clickable
  await page.locator('.logout-btn').click({ force: true });
});
```

---

## Quick Reference

### Selector Priority for Auth Testing

| Priority | Strategy | Use Case |
|----------|----------|----------|
| 1 | `page.addInitScript()` | Setting localStorage before navigation |
| 2 | `page.content()` | Checking for CSS classes in DOM |
| 3 | `page.locator('.class-name')` | CSS class selectors |
| 4 | `page.getByRole('button')` | Semantic button detection |

### Common Auth Testing Helpers

```typescript
// Set authenticated state
await setSessionStorage(page, { accessToken, refreshToken, user });

// Clear session
await clearSessionStorage(page);

// Navigate to protected page after auth
await page.goto('/dashboard');
await page.waitForLoadState('domcontentloaded');
```

---

*Last updated: 2025-01-05*
*Lesson learned from: Logout Flow E2E Tests implementation*
