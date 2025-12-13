# 12.2 Frontend Dashboard Test Plan

## Overview
This test plan covers the dashboard page UI testing.

---

## Test Cases

### 12.2.1 Dashboard Renders
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to /dashboard | Page loads |
| 2 | Chatbots list visible | Shows user's bots |

**Implementation Plan:**
- **Test File:** `frontend/e2e/dashboard.spec.ts`
- **Setup:**
  - Login as user with bots.
- **Steps:**
  1. `await page.goto('/dashboard');`
  2. `await expect(page.getByTestId('chatbot-list')).toBeVisible();`
  3. `await expect(page.getByText('My Bot')).toBeVisible();`

---

### 12.2.2 Usage Statistics
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Dashboard loads | Stats displayed |
| 2 | Token usage shown | Correct values |

**Implementation Plan:**
- **Test File:** `frontend/e2e/dashboard.spec.ts`
- **Steps:**
  1. Locate stats card.
  2. `await expect(page.getByText('Tokens Used')).toBeVisible();`
  3. Verify format matches expected regex (e.g. `\d+ / \d+`).

---

### 12.2.3 Create Chatbot Button
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click "Create Chatbot" | Modal/page opens |
| 2 | Fill form, submit | Chatbot created |

**Implementation Plan:**
- **Test File:** `frontend/e2e/dashboard.spec.ts`
- **Steps:**
  1. `await page.getByRole('button', { name: 'Create Chatbot' }).click();`
  2. `await page.getByLabel('Name').fill('New E2E Bot');`
  3. `await page.getByRole('button', { name: 'Create' }).click();`
  4. `await expect(page.getByText('New E2E Bot')).toBeVisible();`

---

### 12.2.4 Empty State
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | New user with no bots | Empty state shown |
| 2 | CTA to create first bot | Visible |

**Implementation Plan:**
- **Test File:** `frontend/e2e/dashboard.spec.ts`
- **Setup:**
  - Login as new user.
- **Steps:**
  1. `await expect(page.getByText('No chatbots found')).toBeVisible();`
  2. `await expect(page.getByRole('button', { name: 'Create Chatbot' })).toBeVisible();`

---

### 12.2.5 Loading State
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Page loading | Skeleton/spinner |
| 2 | Data loads | Content appears |

**Implementation Plan:**
- **Test File:** `frontend/e2e/dashboard.spec.ts`
- **Steps:**
  1. `await route.continue({ delay: 1000 })` on API call.
  2. Reload page.
  3. `await expect(page.getByTestId('loading-spinner')).toBeVisible();`
  4. `await expect(page.getByTestId('chatbot-list')).toBeVisible();`

---

## How to Run Tests

```bash
cd frontend
npm run test:e2e -- --grep "Dashboard"
```
