# Task 013: Frontend E2E Tests with Playwright

**Priority:** 🟡 Medium (Quality)  
**Phase:** 8 - Test Coverage  
**Estimated Time:** 4-5 hours  
**Dependencies:** None  

---

## Problem Statement

Frontend lacks end-to-end tests for critical user flows:
- User registration and login
- Chatbot creation and configuration
- Source upload and management
- Playground testing

---

## Objective

Create Playwright E2E tests for critical user journeys.

---

## Implementation

### Step 1: Setup Playwright

```bash
cd frontend
npm install -D @playwright/test
npx playwright install
```

### Step 2: Configure Playwright

**File:** `frontend/playwright.config.ts` (NEW)

```typescript
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  retries: 2,
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  webServer: {
    command: 'npm run dev',
    port: 5173,
    reuseExistingServer: !process.env.CI,
  },
});
```

### Step 3: Create E2E Tests

**File:** `frontend/e2e/auth.spec.ts` (NEW)

```typescript
import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('user can register', async ({ page }) => {
    await page.goto('/register');
    
    await page.fill('[name="fullName"]', 'Test User');
    await page.fill('[name="email"]', `test-${Date.now()}@example.com`);
    await page.fill('[name="password"]', 'SecurePass123!');
    
    await page.click('button[type="submit"]');
    
    // Should redirect to dashboard or onboarding
    await expect(page).toHaveURL(/\/(dashboard|onboarding)/);
  });

  test('user can login', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('[name="email"]', 'existing@example.com');
    await page.fill('[name="password"]', 'password');
    
    await page.click('button[type="submit"]');
    
    await expect(page).toHaveURL(/\/dashboard/);
  });
});
```

**File:** `frontend/e2e/chatbot.spec.ts` (NEW)

```typescript
import { test, expect } from '@playwright/test';
import { login } from './helpers';

test.describe('Chatbot Management', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('can create chatbot', async ({ page }) => {
    await page.goto('/dashboard');
    
    await page.click('text=Yeni Bot');
    
    await page.fill('[name="name"]', 'Test Bot');
    await page.click('button[type="submit"]');
    
    await expect(page.locator('text=Test Bot')).toBeVisible();
  });

  test('can add URL source', async ({ page }) => {
    await page.goto('/dashboard/chatbots/test-bot-id');
    
    await page.click('text=Kaynak Ekle');
    await page.fill('[name="url"]', 'https://example.com');
    await page.click('text=Ekle');
    
    await expect(page.locator('text=example.com')).toBeVisible();
  });
});
```

**File:** `frontend/e2e/helpers.ts` (NEW)

```typescript
import { Page } from '@playwright/test';

export async function login(page: Page) {
  await page.goto('/login');
  await page.fill('[name="email"]', 'test@example.com');
  await page.fill('[name="password"]', 'password');
  await page.click('button[type="submit"]');
  await page.waitForURL(/\/dashboard/);
}
```

---

## Acceptance Criteria

- [ ] Playwright configured
- [ ] Auth flow tests pass
- [ ] Chatbot creation test passes
- [ ] Source upload test passes
- [ ] Tests run in CI

---

## Files Changed

| File | Action |
|------|--------|
| `frontend/playwright.config.ts` | CREATE |
| `frontend/e2e/auth.spec.ts` | CREATE |
| `frontend/e2e/chatbot.spec.ts` | CREATE |
| `frontend/e2e/helpers.ts` | CREATE |
| `frontend/package.json` | MODIFY (add scripts) |
