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

---

### 12.2.2 Usage Statistics
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Dashboard loads | Stats displayed |
| 2 | Token usage shown | Correct values |

---

### 12.2.3 Create Chatbot Button
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click "Create Chatbot" | Modal/page opens |
| 2 | Fill form, submit | Chatbot created |

---

### 12.2.4 Empty State
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | New user with no bots | Empty state shown |
| 2 | CTA to create first bot | Visible |

---

### 12.2.5 Loading State
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Page loading | Skeleton/spinner |
| 2 | Data loads | Content appears |

---

## How to Run Tests

```bash
cd frontend
npm run test:e2e -- --grep "Dashboard"
```
