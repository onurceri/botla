# E2E Testing - Learnings & Session Notes

This document captures critical architectural insights and common pitfalls encountered during E2E test implementation.

---

## Core Architecture

### Authentication (HttpOnly Cookies)
- **Mechanism:** Backend sets `botla_token` (access) and `botla_refresh_token` (refresh) as HttpOnly cookies.
- **Frontend Identity:** `AuthContext` uses `/api/v1/me` to determine auth state. It does **not** read cookies directly.
- **Test Implication:** Tests must mock `/api/v1/me` (and legacy `/api/v1/auth/me`) to simulate a logged-in state.
- **Cookie Injection:** Use `context.addCookies([...])` is possible, but mocking the valid response for the *identity endpoint* is often more critical for the frontend app state.

### Dashboard Dependencies
- **Context Hell:** The Dashboard Layout requires multiple contexts to be active: `AuthContext`, `OrganizationContext`.
- **Mock Requirement:** Dashboard tests fail silently or get stuck in "loading" if `Organization` or `Analytics` endpoints are not mocked. Always use `setupOrgMocks(page)` and `setupAnalyticsMocks(page)`.

---

## Critical Patterns & Fixes

### 1. Robust Page Load Verification
**Problem:** Tests failing because the page wasn't fully ready.
**Fix:** Added `data-testid="page-dashboard"` to the root layout div.
**Pattern:**
```typescript
await page.goto('/dashboard');
await expect(page.getByTestId('page-dashboard')).toBeVisible();
```

### 2. URL Redirection
**Problem:** `expect(page.url()).toContain(...)` fails because checks happen before the router finishes the transition.
**Fix:** Use `await expect(page).toHaveURL(...)`.

### 3. Mobile Interactions
**Problem:** 'Element not visible' errors when trying to click hamburger menus.
**Fix:**
- Ensure correct Viewport size is set.
- Use explicit, specific selectors for mobile-only elements: `page.locator('header button.lg\\:hidden')`.
- Ensure elements are scrolled into view if necessary (Playwright usually handles this, but explicit logic in Page Objects helps).

### 4. CSRF & Security Testing
**Problem:** Generic mock routes (like `page.route('**/*')`) can inadvertently block specific security test scenarios.
**Fix:** Define specific security mocks (e.g., 403 response for missing token) *after* or carefully ordered relative to generic mocks.

### 5. Mock Setup Timing ⚠️ CRITICAL
**Problem:** Tests failing with "Cannot navigate to invalid URL" or unexpected redirects to `/login` before form submission.
**Root Cause:** AuthContext initializes immediately on page load. If mocks aren't set up beforehand, AuthContext sees 401 responses and triggers redirects.
**Evidence:** Call logs show pattern like `2 × /register` → `17 × /login` before any form interaction.
**Fix:** Always set up mocks BEFORE calling `page.goto()`:
```typescript
test.beforeEach(async ({ page }) => {
    await mockSuccessfulRegistration(page)  // Mocks FIRST
    await page.goto('/register')
    await clearAuthStorage(page)
})
```

### 6. URL Navigation with baseURL ⚠️ CRITICAL
**Problem:** Using relative URLs (`/register`) with `baseURL` from environment variable fails with "Cannot navigate to invalid URL".
**Root Cause:** Playwright's `baseURL` resolution can be inconsistent when using relative paths in certain configurations.
**Fix:** Use full URLs for reliable navigation:
```typescript
// Use full URLs
await page.goto('http://localhost:5173/register')

// Or set E2E_BASE_URL explicitly when running
E2E_BASE_URL=http://localhost:5173 npx playwright test e2e/register.spec.ts
```

### 7. HttpOnly Cookie Setting in Mocks
**Problem:** Refresh token cookie not being set when using `Set-Cookie` header in route fulfillment.
**Root Cause:** Multiple cookies in a single `Set-Cookie` header may not be parsed correctly, or the header approach doesn't work for all cookies.
**Fix:** Use `page.context().addCookies()` after route fulfillment:
```typescript
await page.route('**/api/v1/auth/login', async (route) => {
    if (route.request().method() === 'POST') {
        await route.fulfill({
            status: 200,
            body: JSON.stringify({ token: accessToken, refresh_token: refreshToken }),
        });
        // Set refresh token separately using context
        await page.context().addCookies([{
            name: 'botla_refresh_token',
            value: refreshToken,
            url: 'http://localhost:5173',
            httpOnly: true,
            sameSite: 'Lax',
        }]);
    }
});
```

---

```bash
# Run specific test file
npx playwright test e2e/dashboard.spec.ts

# debug mode
npx playwright test --debug

# Run usage
npx playwright test --ui
```

---

## Infrastructure Refactoring (Jan 2026)

### Changes Made
1. **Created `utils/index.ts`** - Single entry point for all utility imports
2. **Created `mocks/index.ts`** - Single entry point for all mock imports  
3. **Removed `sessions.spec.ts.bak`** - Cleaned up backup file
4. **Updated `TESTING_STANDARDS.md`** - Added project structure documentation

### Import Recommendations
```typescript
// ✅ Preferred: Import from index files
import { setAuthCookies, createSession, TEST_IDS } from './utils';
import { mockSuccessfulLogin, mockUserInfo } from './mocks';

// ⚠️ Legacy: Still works but prefer index imports
import { setupAuthMocks, setupOrgMocks } from './helpers';
```

### Files to Consider for Future Cleanup
- `utils/test-helpers.ts` - Contains unused helper classes (ButtonHelper, InputHelper, etc.)
- `utils/selectors.ts` - Has extensive documentation (200+ lines) but selectors are minimally used
- Both files are kept for potential future use but are not imported by any spec files

*Last updated: 2026-01-08*

