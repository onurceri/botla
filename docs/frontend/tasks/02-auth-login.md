# Task: Implement Login Page Tests

> **Task ID**: 02-auth-login  
> **Source**: TEST_PATHS.md Section 2.1  
> **Priority**: Highest (Authentication)  
> **Estimated Effort**: 6-8 hours  
> **Prerequisite**: 01-test-naming-conventions.md (must be completed first)

---

## Detailed Prompt

Implement comprehensive E2E tests for the Login Page (`frontend/e2e/auth.spec.ts`). This task covers all login page functionality including form interactions, validation, hover states, keyboard navigation, and error handling.

### Context

The Login Page is the entry point for authenticated access to the Botla-Co dashboard. Testing this page thoroughly ensures:
- Users can successfully authenticate
- Invalid inputs are properly rejected
- Security vulnerabilities are prevented
- Accessibility requirements are met
- Edge cases are handled gracefully

### Reference Specifications

From `docs/frontend/TEST_PATHS.md` Section 2.1:

#### 2.1.1 Page Elements

| Element ID | Type | Description |
|------------|------|-------------|
| `input-email` | text | Email input field |
| `input-password` | password | Password input field |
| `btn-login` | submit | Login button |
| `link-forgot-password` | link | Forgot password link |
| `link-register` | link | Register new account link |
| `checkbox-remember` | checkbox | Remember me checkbox |
| `text-error` | text | Error message display |

#### 2.1.2 Login Flow Interactions

```
Login Flow
├── Load login page
│   ├── Hover: Email input label (shows tooltip)
│   ├── Hover: Password input label (shows tooltip)
│   ├── Click: Email input → focus state
│   ├── Click: Password input → focus state
│   ├── Tab: Navigate through fields
│   ├── Enter: Email input → focus password
│   ├── Enter: Password input → submit form
│   └── Type: Email field (validation on blur)
│
├── Submit with empty fields
│   ├── Click: btn-login
│   ├── Assert: `toast-error` - "Email is required"
│   └── Assert: `input-email` has error class
│
├── Submit with invalid email
│   ├── Type: "invalid-email"
│   ├── Blur: Email input
│   ├── Assert: `input-email` has error class
│   └── Assert: `text-error` - "Invalid email format"
│
├── Submit with valid credentials
│   ├── Type: Valid email
│   ├── Type: Valid password
│   ├── Click: btn-login
│   ├── Assert: Loading spinner visible
│   ├── Assert: btn-login disabled
│   ├── Wait: API response
│   └── Redirect: /dashboard
│
├── Remember me checkbox
│   ├── Check: checkbox-remember
│   ├── Login successfully
│   └── Assert: Refresh token stored in localStorage
│
└── Forgot password flow
    ├── Click: link-forgot-password
    ├── Assert: URL contains /forgot-password
    ├── Type: email
    ├── Click: btn-send-reset
    └── Assert: `toast-success` - "Reset link sent"
```

#### 2.1.3 Hover States

| Element | Expected Hover Behavior |
|---------|------------------------|
| `btn-login` | Darken background, scale 1.02 |
| `link-forgot-password` | Underline, color change |
| `link-register` | Underline, color change |
| Input labels | Slight color change |

#### 2.1.4 Keyboard Navigation

| Key | Action |
|-----|--------|
| `Tab` | Navigate forward through inputs |
| `Shift+Tab` | Navigate backward |
| `Enter` | Submit form (when focused on submit) |
| `Escape` | Close any open dropdowns/modals |

### Implementation Requirements

1. **Create/Update Login Test File** (`frontend/e2e/auth.spec.ts`)
   - Implement all test cases from the specification
   - Use consistent naming from task 01
   - Follow established test patterns

2. **Create Login Page Object** (`frontend/e2e/pages/login.page.ts`)
   - Encapsulate login page interactions
   - Provide reusable methods
   - Improve test maintainability

3. **Create Test Data Fixtures** (`frontend/e2e/fixtures/auth.fixture.ts`)
   - Valid user credentials
   - Invalid user credentials
   - Edge case data

4. **Create Mock API Handlers** (`frontend/e2e/mocks/auth.mocks.ts`)
   - Mock successful login responses
   - Mock error responses
   - Handle authentication endpoints

### Expected Deliverables

1. `frontend/e2e/auth.spec.ts` - Comprehensive login tests
2. `frontend/e2e/pages/login.page.ts` - Page object model
3. `frontend/e2e/fixtures/auth.fixture.ts` - Test data fixtures
4. `frontend/e2e/mocks/auth.mocks.ts` - API mock handlers
5. Updated `playwright.config.ts` if additional configuration needed

---

## Implementation Plan

### Phase 1: Setup and Page Object

- [x] Create `frontend/e2e/pages/login.page.ts` with:
  - [x] Locators for all page elements
  - [x] Methods for all interactions
  - [x] Helper methods for assertions
- [x] Create `frontend/e2e/fixtures/auth.fixture.ts` with test data
- [x] Create `frontend/e2e/mocks/auth.mocks.ts` with API handlers

### Phase 2: Basic Functionality Tests

- [x] Test: Page loads successfully
- [x] Test: Email input focus state
- [x] Test: Password input focus state
- [x] Test: Tab navigation through fields
- [x] Test: Enter key in email moves to password
- [x] Test: Enter key in password submits form

### Phase 3: Validation Tests

- [x] Test: Empty fields validation
- [x] Test: Invalid email format validation
- [x] Test: Invalid password validation
- [x] Test: Error messages display correctly
- [x] Test: Error states on input fields

### Phase 4: Authentication Tests

- [x] Test: Successful login with valid credentials
- [x] Test: Loading state during login
- [x] Test: Button disabled during submission
- [x] Test: Redirect to dashboard after login
- [x] Test: Remember me functionality
- [x] Test: Token storage verification

### Phase 5: Forgot Password Flow

- [x] Test: Navigate to forgot password page
- [x] Test: Submit email for password reset
- [x] Test: Success message after submission
- [x] Test: Navigation to register page

### Phase 6: Hover and Visual Tests

- [x] Test: Login button hover state
- [x] Test: Forgot password link hover state
- [x] Test: Register link hover state
- [x] Test: Input label hover state

### Phase 7: Keyboard Navigation Tests

- [x] Test: Forward tab navigation
- [x] Test: Backward shift+tab navigation
- [x] Test: Enter submits form
- [x] Test: Escape closes modals/dropdowns

### Phase 8: Error Handling Tests

- [x] Test: Network error handling
- [x] Test: API error response handling
- [x] Test: Session expired handling
- [x] Test: Multiple failed login attempts

---

## Technical Notes

### Page Object Pattern

Create a reusable page object for login interactions:

```typescript
// frontend/e2e/pages/login.page.ts
import { Locator, Page, expect } from '@playwright/test';

export class LoginPage {
  readonly page: Page;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly loginButton: Locator;
  readonly forgotPasswordLink: Locator;
  readonly registerLink: Locator;
  readonly rememberMeCheckbox: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.emailInput = page.locator('[data-testid="input-email"]');
    this.passwordInput = page.locator('[data-testid="input-password"]');
    this.loginButton = page.locator('[data-testid="btn-login"]');
    this.forgotPasswordLink = page.locator('[data-testid="link-forgot-password"]');
    this.registerLink = page.locator('[data-testid="link-register"]');
    this.rememberMeCheckbox = page.locator('[data-testid="checkbox-remember"]');
    this.errorMessage = page.locator('[data-testid="text-error"]');
  }

  async goto() {
    await this.page.goto('/login');
  }

  async fillEmail(email: string) {
    await this.emailInput.fill(email);
  }

  async fillPassword(password: string) {
    await this.passwordInput.fill(password);
  }

  async clickLogin() {
    await this.loginButton.click();
  }

  async checkRememberMe() {
    await this.rememberMeCheckbox.check();
  }

  async login(email: string, password: string, rememberMe = false) {
    await this.fillEmail(email);
    await this.fillPassword(password);
    if (rememberMe) {
      await this.checkRememberMe();
    }
    await this.clickLogin();
  }

  async expectErrorMessage(message: string) {
    await expect(this.errorMessage).toHaveText(message);
  }

  async expectUrlContains(path: string) {
    await expect(this.page).toHaveURL(new RegExp(path));
  }
}
```

### Test Data Management

Use Playwright's test.extend for custom fixtures:

```typescript
// frontend/e2e/fixtures/auth.fixture.ts
import { test as base } from '@playwright/test';

export const test = base.extend({
  // Valid test user
  validUser: async () => ({
    email: 'test@example.com',
    password: 'SecurePass123!',
  }),

  // Invalid test data
  invalidEmail: async () => 'not-an-email',
  invalidPassword: async () => 'wrongpassword',

  // Login page instance
  loginPage: async ({ page }, use) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await use(loginPage);
  },
});
```

### API Mocking

Intercept and mock authentication API calls:

```typescript
// frontend/e2e/mocks/auth.mocks.ts
import { APIRequestContext } from '@playwright/test';

export async function mockSuccessfulLogin(request: APIRequestContext) {
  await request.post('/api/v1/auth/login', {
    status: 200,
    body: {
      access_token: 'mock-access-token',
      refresh_token: 'mock-refresh-token',
      user: {
        id: 'user-123',
        email: 'test@example.com',
        name: 'Test User',
      },
    },
  });
}

export async function mockFailedLogin(request: APIRequestContext, reason: string) {
  await request.post('/api/v1/auth/login', {
    status: 401,
    body: {
      error: reason,
      message: 'Invalid email or password',
    },
  });
}
```

### State Management

Handle authentication state in tests:

```typescript
// Use storage state for authenticated tests
test.use({
  storageState: 'e2e/.auth/user.json',
});
```

### Running Specific Tests

```bash
# Run all login tests
cd frontend && npx playwright test auth.spec.ts

# Run a specific test
cd frontend && npx playwright test auth.spec.ts -g "should login successfully"

# Run in debug mode
cd frontend && npx playwright test auth.spec.ts --debug

# Run with trace
cd frontend && npx playwright test auth.spec.ts --trace on
```

---

## Verification Steps

### 1. Test Coverage Verification
- [x] All element interactions are tested
- [x] All validation scenarios covered
- [x] All error states tested
- [x] All keyboard navigation tested
- [x] All hover states tested

### 2. Test Execution Verification
- [x] All tests pass locally
- [x] Tests work in CI environment
- [x] No flaky tests
- [x] Proper timeout handling

### 3. Code Quality Verification
- [x] Page object pattern used correctly
- [x] Selectors follow naming conventions
- [x] Tests are maintainable and readable
- [x] No hardcoded values without explanation

### 4. Accessibility Verification
- [x] Keyboard navigation works
- [x] Focus states visible
- [x] ARIA labels present

---

## Execution Notes for Developer Agent

### Key Considerations

1. **Wait for API responses** - Use `await page.waitForResponse()` to wait for authentication
2. **Check loading states** - Verify spinner appears and button is disabled
3. **Verify redirects** - Use `await expect(page).toHaveURL()` for navigation
4. **Handle toasts** - Wait for toast notifications using test IDs
5. **Use test.beforeEach()** - Reset page state before each test

### Common Issues to Avoid

1. **Race conditions** - Wait for elements to be visible before interacting
2. **Hardcoded waits** - Use explicit waits instead of `sleep()`
3. **Magic values** - Use constants for expected messages
4. **Skipping edge cases** - Test boundary conditions

### Test Order

Run tests in this order for efficiency:
1. Basic functionality first (page load, focus states)
2. Validation tests (error cases)
3. Authentication tests (happy path)
4. Visual/interaction tests (hover, keyboard)
5. Error handling tests (last, as they may pollute state)

---

## Dependencies

- **Prerequisite**: 01-test-naming-conventions.md (must be completed)
- **Environment**: Backend API must be running for integration tests
- **Test Data**: Valid test user credentials required

---

## Related Tasks

- 01-test-naming-conventions.md - Foundation for naming conventions
- 03-auth-register.md - Registration page tests
- 04-auth-logout.md - Logout flow tests
- 05-auth-sessions.md - Session management tests

---

*Task created from: docs/frontend/TEST_PATHS.md Section 2.1*
