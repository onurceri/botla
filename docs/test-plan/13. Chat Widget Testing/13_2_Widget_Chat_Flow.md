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

---

### 13.2.2 Send Message
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Type message | Text entered |
| 2 | Click send | Message sent |
| 3 | Response appears | Bot reply shown |

---

### 13.2.3 Enter Key Sends
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Type message | Text entered |
| 2 | Press Enter | Message sent |

---

### 13.2.4 Loading Indicator
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send message | Loading shown |
| 2 | Response received | Loading hidden |

---

### 13.2.5 Suggested Questions
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Widget opens | Suggestions shown |
| 2 | Click suggestion | Message sent |

---

### 13.2.6 Sources Displayed
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Ask question | Response with sources |
| 2 | Sources visible | Clickable links |

---

## How to Run Tests

```bash
cd widget
npm run test -- --grep "Chat"
```
