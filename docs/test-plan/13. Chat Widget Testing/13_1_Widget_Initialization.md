# 13.1 Widget Initialization Test Plan

## Overview
This test plan covers widget loading and initialization.

---

## Test Cases

### 13.1.1 Widget Script Loads
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Include widget script | Script loads |
| 2 | No console errors | Clean load |

---

### 13.1.2 Widget Injects DOM
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Script executes | Widget element injected |
| 2 | Shadow DOM used | Styles isolated |

---

### 13.1.3 Config Fetched
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Widget initialized | API call made |
| 2 | Config applied | Theme, messages set |

---

### 13.1.4 Invalid Chatbot ID
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Invalid chatbot_id | Error displayed |
| 2 | Widget gracefully fails | Not broken |

---

## How to Run Tests

```bash
cd widget
npm run test
```
