# 12.1 Frontend Authentication Pages Test Plan

## Overview
This test plan covers frontend authentication UI testing.

---

## Test Cases

### 12.1.1 Login Page Renders
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to /login | Page loads |
| 2 | Form elements visible | Email, password, submit |

**Implementation Plan:**
- **Test File:** `frontend/e2e/auth.spec.ts`
- **Steps:**
  1. `await page.goto('/login');`
  2. `await expect(page.getByLabel('Email')).toBeVisible();`
  3. `await expect(page.getByLabel('Password')).toBeVisible();`
  4. `await expect(page.getByRole('button', { name: 'Login' })).toBeVisible();`

---

### 12.1.2 Login Form Validation
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit empty form | Validation errors |
| 2 | Submit invalid email | Error message |

**Implementation Plan:**
- **Test File:** `frontend/e2e/auth.spec.ts`
- **Steps:**
  1. `await page.getByRole('button', { name: 'Login' }).click();`
  2. `await expect(page.getByText('Email is required')).toBeVisible();`
  3. `await page.getByLabel('Email').fill('invalid');`
  4. `await expect(page.getByText('Invalid email')).toBeVisible();`

---

### 12.1.3 Successful Login
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Enter valid credentials | Submit |
| 2 | Redirect to dashboard | Logged in state |

**Implementation Plan:**
- **Test File:** `frontend/e2e/auth.spec.ts`
- **Steps:**
  1. `await page.getByLabel('Email').fill('test@example.com');`
  2. `await page.getByLabel('Password').fill('password');`
  3. `await page.getByRole('button', { name: 'Login' }).click();`
  4. `await expect(page).toHaveURL(/\/dashboard/);`

---

### 12.1.4 Registration Page
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to /register | Page loads |
| 2 | Fill form, submit | Account created |
| 3 | Redirect to dashboard | Logged in |

**Implementation Plan:**
- **Test File:** `frontend/e2e/auth.spec.ts`
- **Steps:**
  1. `await page.goto('/register');`
  2. Fill Name, Email, Password.
  3. Click Register.
  4. `await expect(page).toHaveURL(/\/dashboard/);`

---

### 12.1.5 Logout
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click logout button | Token cleared |
| 2 | Redirect to login | Logged out state |

**Implementation Plan:**
- **Test File:** `frontend/e2e/auth.spec.ts`
- **Steps:**
  1. Login first.
  2. `await page.getByRole('button', { name: 'Logout' }).click();`
  3. `await expect(page).toHaveURL(/\/login/);`
  4. Check local storage token cleared (optional).

---

## How to Run Tests

```bash
cd frontend
npm run test:e2e -- --grep "Auth"
```
