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

---

### 12.1.2 Login Form Validation
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit empty form | Validation errors |
| 2 | Submit invalid email | Error message |

---

### 12.1.3 Successful Login
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Enter valid credentials | Submit |
| 2 | Redirect to dashboard | Logged in state |

---

### 12.1.4 Registration Page
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to /register | Page loads |
| 2 | Fill form, submit | Account created |
| 3 | Redirect to dashboard | Logged in |

---

### 12.1.5 Logout
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click logout button | Token cleared |
| 2 | Redirect to login | Logged out state |

---

## How to Run Tests

```bash
cd frontend
npm run test:e2e -- --grep "Auth"
```
