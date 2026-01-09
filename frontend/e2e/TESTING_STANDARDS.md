# E2E Testing Standards

This document defines the comprehensive testing standards, naming conventions, and best practices for all E2E and integration tests in the Botla-Co frontend test suite.

## Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [File Naming Conventions](#file-naming-conventions)
4. [Test Naming Patterns](#test-naming-patterns)
5. [Selector Strategy](#selector-strategy)
6. [Mock Setup Guidelines](#mock-setup-guidelines)
7. [Best Practices](#best-practices)
8. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)

---

## Overview

These standards ensure:
- **Consistent element identification** across all test files
- **Maintainable and readable** test code
- **Stable tests** that don't break with minor UI changes

---

## Project Structure

```
e2e/
├── fixtures/           # Playwright custom fixtures
│   ├── auth.fixture.ts
│   └── register.fixture.ts
├── mocks/              # API mock utilities
│   ├── index.ts        # ⭐ Single entry point for all mocks
│   ├── auth.mocks.ts
│   ├── tokens.mocks.ts
│   ├── register.mocks.ts
│   └── session.mocks.ts
├── pages/              # Page Object Models
│   ├── login.page.ts
│   ├── sidebar.page.ts
│   └── ...
├── utils/              # Test utilities
│   ├── index.ts        # ⭐ Single entry point for utils
│   ├── cookie-auth.ts  # Auth/session helpers
│   └── session-manager.ts
├── *.spec.ts           # Test spec files
├── helpers.ts          # General mock setup helpers
├── test-constants.ts   # UI text constants (Turkish/English)
└── TESTING_STANDARDS.md
```

### Recommended Import Patterns

```typescript
// ✅ Preferred: Import from index files
import { setAuthCookies, mockUserInfo, TEST_IDS } from './utils';
import { mockSuccessfulLogin, setupSessionMocks } from './mocks';

// ⚠️ Acceptable: Import from specific files when needed
import { mockSuccessfulTokenRefresh } from './mocks/tokens.mocks';
import { setupAuthMocks, setupOrgMocks } from './helpers';
```

---

## File Naming Conventions

### Test Files
`{page-or-feature}.spec.ts` (e.g., `auth.spec.ts`, `dashboard.spec.ts`)

### Utility Files
`utils/{purpose}.ts` (e.g., `utils/cookie-auth.ts`)

---


## Test Naming Patterns

### Describe Blocks and Tests
Use `test.describe()` for feature areas and `should {action} when {condition}` for test names.

```typescript
test.describe('Feature Area', () => {
    test.beforeEach(async ({ page }) => { ... });
    test('should perform action when user does X', async () => { ... });
});
```

---

## Selector Strategy

### Primary: `data-testid`
Use `data-testid` attributes heavily. Add `data-testid` to the root element of critical pages (`page-dashboard`) to ensure robust page load verification.

```typescript
// Good
await expect(page.getByTestId('page-dashboard')).toBeVisible();
await page.getByTestId('btn-login').click();
```

### Secondary: Semantic Selectors
Use roles, labels, and text when `data-testid` is overkill or unavailable.

```typescript
await page.getByRole('button', { name: /save/i }).click();
await page.getByLabel('Email').fill('user@example.com');
```

### Mobile & Hidden Elements
For elements visible only on specific viewports (e.g., mobile hamburger menu), use specific CSS selectors if necessary, ensuring checking visibility or scrolling into view.

```typescript
// Mobile menu button
await page.locator('header button.lg\\:hidden').first().click();
```

---

## Mock Setup Guidelines

### Authentication & Essential Data
Properly mock the session **and** essential dependencies like Organization and Analytics for complex pages (Dashboard).

```typescript
import { setupAuthMocks, setupOrgMocks, setupAnalyticsMocks } from './helpers';
import { mockUserInfo } from './mocks/tokens.mocks';

test.beforeEach(async ({ page }) => {
    await setupAuthMocks(page);       // Login/Register routes
    await setupOrgMocks(page);        // Org context dependency
    await setupAnalyticsMocks(page);  // Dashboard charts dependency
    await mockUserInfo(page);         // /api/v1/me and /api/v1/auth/me
});
```

### Mock Setup Timing ⚠️ CRITICAL
**Always set up mocks BEFORE navigating to the page.** AuthContext initializes immediately on page load and will trigger redirects if auth endpoints aren't properly mocked.

```typescript
// Bad - Navigate before mock setup
test.beforeEach(async ({ page }) => {
    await page.goto('/register');
    await clearAuthStorage(page);
    await mockSuccessfulRegistration(page); // Too late! AuthContext already initialized
});

// Good - Mock setup before navigation
test.beforeEach(async ({ page }) => {
    await mockSuccessfulRegistration(page);  // Set up mocks FIRST
    await page.goto('/register');
    await clearAuthStorage(page);
});
```

### URL Navigation ⚠️ CRITICAL
Use **full URLs** instead of relative paths to avoid baseURL resolution issues in some Playwright configurations.

```typescript
// Use full URLs for reliable navigation
await page.goto('http://localhost:5173/register');
await page.goto('http://localhost:5173/dashboard');

// If using relative URLs, always set E2E_BASE_URL explicitly
E2E_BASE_URL=http://localhost:5173 npx playwright test e2e/register.spec.ts
```

### Cookie Setting in Mocks
For HttpOnly cookie authentication, set cookies using `context.addCookies()` after route fulfillment:

```typescript
await page.route('**/api/v1/auth/login', async (route) => {
    if (route.request().method() === 'POST') {
        const accessToken = 'mock-token';
        const refreshToken = 'mock-refresh';

        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ token: accessToken, refresh_token: refreshToken }),
        });

        // Set refresh token cookie using context (Set-Cookie header may not work for all cookies)
        await page.context().addCookies([
            {
                name: 'botla_refresh_token',
                value: refreshToken,
                url: 'http://localhost:5173',
                httpOnly: true,
                sameSite: 'Lax',
            },
        ]);
    }
});
```

### Mock Ordering
**Specific mocks must be defined AFTER generic mocks** if using `page.route` overrides, though usually Playwright respects the most recently defined route for a matching URL. However, grouping mocks logically (e.g., blocking CSRF checks) often requires specific handling.

---

## Best Practices

### 1. Waiting for Navigation
Do NOT assert `page.url()` immediately after an action. Use `expect(page).toHaveURL(...)` to retry until the condition is met.

```typescript
// Bad
await page.click('button');
expect(page.url()).toContain('/dashboard'); // Fails if redirect takes 50ms

// Good
await page.click('button');
await expect(page).toHaveURL(/\/dashboard/); // Retries automatically
```

### 2. Session Storage Helper
Use `setSessionStorage` helper instead of `page.evaluate` manually, to ensure types are respected.

### 3. Verify Page Load
Always verify the page has fully loaded before interacting with elements, preferably by asserting the visibility of a root `data-testid`.

```typescript
await page.goto('/dashboard');
await expect(page.getByTestId('page-dashboard')).toBeVisible();
```

---

## Anti-Patterns to Avoid

### Hardcoded Waits
Never use `page.waitForTimeout(1000)`. Use auto-retrying assertions or specific event waits.

### Fragile XPath/CSS
Avoid `div > div:nth-child(3) > span`. Use `data-testid` or stable text/role selectors.

### Ignoring 401 Redirects in Tests
If simulating an expired token, ensure your test expects a redirect to `/login`, rather than just failing to find a dashboard element.
