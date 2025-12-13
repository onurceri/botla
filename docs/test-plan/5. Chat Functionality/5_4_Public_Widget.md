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

**Implementation Plan:**
- **Test File:** `internal/integration/chat_widget_public_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. `GET /api/v1/widget/{id}/config`.
  2. Verify 200 OK without Authorization header.
  3. Verify JSON contains `theme`, `welcome_message`.

---

### 5.4.2 Widget Displays Welcome Message
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Open widget | Welcome message shown |
| 2 | Message matches config | Correct text |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - Bot with `welcome_message="Hello World"`.
- **Steps:**
  1. Load demo page with bot ID.
  2. Open chat.
  3. Verify text "Hello World" is visible.

---

### 5.4.3 Widget Displays Suggested Questions
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Open widget | Suggestions displayed |
| 2 | Click suggestion | Question sent |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - Bot with suggested questions.
- **Steps:**
  1. Open chat.
  2. Verify suggestion buttons exist.
  3. Click one. Verify message sent.

---

### 5.4.4 Widget Theme Applied
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Chatbot theme_color = #FF0000 | Widget uses red |
| 2 | All colors applied | Header, buttons, etc. |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - Bot with `theme_color="#ff0000"`.
- **Steps:**
  1. Open chat.
  2. Verify header background color is `rgb(255, 0, 0)`.

---

### 5.4.5 Widget Position
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | position = "bottom-right" | Widget in bottom-right |
| 2 | position = "bottom-left" | Widget in bottom-left |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - Bot with `position="bottom-left"`.
- **Steps:**
  1. Load page.
  2. Verify launcher button CSS `left` property is set (and `right` is auto or unset).

---

### 5.4.6 Branding Display
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free plan | "Powered by Botla" visible |
| 2 | Pro + hide_branding | Branding hidden |
| 3 | Ultra + custom_branding | Custom branding shown |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - 3 Bots: Free, Pro (hidden), Ultra (custom).
- **Steps:**
  1. Load Free bot. Verify "Powered by Botla".
  2. Load Pro bot. Verify "Powered by Botla" is absent.
  3. Load Ultra bot. Verify custom text/logo.

---

### 5.4.7 Auto-Open Functionality
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Widget with auto-open=1 | Opens automatically |
| 2 | Widget with auto-open=0 | Remains closed |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - Bot with `auto_open=true`.
- **Steps:**
  1. Load page.
  2. Wait 1-2s.
  3. Verify chat window is visible without clicking.

---

### 5.4.8 Sources Displayed in Widget
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Ask question | Response received |
| 2 | Sources used shown | Source links/names |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - Bot with sources.
- **Steps:**
  1. Ask question.
  2. Wait for response.
  3. Verify "Sources" section appears with links.

---

## Secure Embed Test Cases

### 5.4.9 Domain Validation
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Load widget from allowed domain | Widget loads |
| 2 | Load from non-allowed domain | 403 error |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_widget_public_test.go`
- **Setup:**
  - Bot with `secure_embed_enabled=true`, `allowed_domains=["allowed.com"]`.
- **Steps:**
  1. GET config with `Referer: https://allowed.com`. Expect 200.
  2. GET config with `Referer: https://evil.com`. Expect 403.

---

### 5.4.10 Embed Secret Validation
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Request with valid secret | Token returned |
| 2 | Request with invalid secret | 401 Unauthorized |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_widget_public_test.go`
- **Setup:**
  - Bot with `secure_embed_enabled=true`, `embed_secret="SECRET"`.
- **Steps:**
  1. POST `/api/v1/widget/token` with `secret="SECRET"`. Expect 200 + Token.
  2. POST with `secret="WRONG"`. Expect 401.

---

## How to Run Tests

```bash
# E2E tests with Playwright
cd frontend
npm run test:e2e -- --grep "Widget"
```
