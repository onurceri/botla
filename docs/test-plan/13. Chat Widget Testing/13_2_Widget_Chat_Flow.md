# 13.2 Widget Chat Flow Test Plan

## Overview
This test plan covers the widget chat user experience.

---

## Test Cases

### 13.2.1 Open/Close Widget
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click chat bubble | Window opens |
| 2 | Click X button | Window closes |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Steps:**
  1. `await page.getByTestId('launcher').click();`
  2. `await expect(page.getByTestId('chat-window')).toBeVisible();`
  3. `await page.getByTestId('close-button').click();`
  4. `await expect(page.getByTestId('chat-window')).toBeHidden();`

---

### 13.2.2 Send Message
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Type message | Text entered |
| 2 | Click send | Message sent |
| 3 | Response appears | Bot reply shown |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Steps:**
  1. Open chat.
  2. `await page.getByPlaceholder('Type a message...').fill('Hello');`
  3. `await page.getByTestId('send-button').click();`
  4. `await expect(page.getByText('Hello', { exact: true })).toBeVisible();`
  5. `await expect(page.getByTestId('bot-message')).toBeVisible();`

---

### 13.2.3 Enter Key Sends
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Type message | Text entered |
| 2 | Press Enter | Message sent |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Steps:**
  1. `await page.getByPlaceholder('Type...').fill('Enter test');`
  2. `await page.keyboard.press('Enter');`
  3. `await expect(page.getByText('Enter test')).toBeVisible();`

---

### 13.2.4 Loading Indicator
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send message | Loading shown |
| 2 | Response received | Loading hidden |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Steps:**
  1. Intercept chat API with delay.
  2. Send message.
  3. `await expect(page.getByTestId('typing-indicator')).toBeVisible();`
  4. Wait for response.
  5. `await expect(page.getByTestId('typing-indicator')).toBeHidden();`

---

### 13.2.5 Suggested Questions
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Widget opens | Suggestions shown |
| 2 | Click suggestion | Message sent |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - Mock config with suggestions.
- **Steps:**
  1. Open widget.
  2. `await page.getByText('What is this?').click();`
  3. Verify 'What is this?' appears as user message.

---

### 13.2.6 Sources Displayed
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Ask question | Response with sources |
| 2 | Sources visible | Clickable links |

**Implementation Plan:**
- **Test File:** `widget/e2e/chat.spec.ts`
- **Setup:**
  - Mock chat response with sources.
- **Steps:**
  1. Send message.
  2. Verify response contains "Sources" section.
  3. Verify link to source is correct.

---

## How to Run Tests

```bash
cd widget
npm run test -- --grep "Chat"
```
