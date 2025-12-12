# 5.4 Public Chat Widget Test Plan

## Overview
This test plan covers the public-facing chat widget and embed functionality.

---

## Test Cases

### 5.4.1 Widget Loads Config
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Load widget with chatbot_id | Config fetched |
| 2 | No authentication required | Public endpoint |
| 3 | Config includes theme, welcome message | All fields |

---

### 5.4.2 Widget Displays Welcome Message
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Open widget | Welcome message shown |
| 2 | Message matches config | Correct text |

---

### 5.4.3 Widget Displays Suggested Questions
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Open widget | Suggestions displayed |
| 2 | Click suggestion | Question sent |

---

### 5.4.4 Widget Theme Applied
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Chatbot theme_color = #FF0000 | Widget uses red |
| 2 | All colors applied | Header, buttons, etc. |

---

### 5.4.5 Widget Position
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | position = "bottom-right" | Widget in bottom-right |
| 2 | position = "bottom-left" | Widget in bottom-left |

---

### 5.4.6 Branding Display
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free plan | "Powered by Botla" visible |
| 2 | Pro + hide_branding | Branding hidden |
| 3 | Ultra + custom_branding | Custom branding shown |

---

### 5.4.7 Auto-Open Functionality
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Widget with auto-open=1 | Opens automatically |
| 2 | Widget with auto-open=0 | Remains closed |

---

### 5.4.8 Sources Displayed in Widget
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Ask question | Response received |
| 2 | Sources used shown | Source links/names |

---

## Secure Embed Test Cases

### 5.4.9 Domain Validation
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Load widget from allowed domain | Widget loads |
| 2 | Load from non-allowed domain | 403 error |

---

### 5.4.10 Embed Secret Validation
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Request with valid secret | Token returned |
| 2 | Request with invalid secret | 401 Unauthorized |

---

## How to Run Tests

```bash
# E2E tests with Playwright
cd frontend
npm run test:e2e -- --grep "Widget"
```
