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

**Implementation Plan:**
- **Test File:** `widget/e2e/init.spec.ts`
- **Steps:**
  1. `await page.goto('/demo.html');`
  2. Verify network request for `widget.js` returns 200.
  3. Verify no console errors.

---

### 13.1.2 Widget Injects DOM
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Script executes | Widget element injected |
| 2 | Shadow DOM used | Styles isolated |

**Implementation Plan:**
- **Test File:** `widget/e2e/init.spec.ts`
- **Steps:**
  1. `await expect(page.locator('#botla-widget-container')).toBeAttached();`
  2. Verify shadow root exists.

---

### 13.1.3 Config Fetched
**Priority:** Critical  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Widget initialized | API call made |
| 2 | Config applied | Theme, messages set |

**Implementation Plan:**
- **Test File:** `widget/e2e/init.spec.ts`
- **Steps:**
  1. Intercept `GET */api/v1/widget/*/config`.
  2. `await page.goto('/demo.html');`
  3. Verify request was made.

---

### 13.1.4 Invalid Chatbot ID
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Invalid chatbot_id | Error displayed |
| 2 | Widget gracefully fails | Not broken |

**Implementation Plan:**
- **Test File:** `widget/e2e/init.spec.ts`
- **Setup:**
  - Inject bad ID in script tag.
- **Steps:**
  1. Load page.
  2. Verify API returns 404.
  3. Verify console error logged (graceful failure, no crash loop).

---

## How to Run Tests

```bash
cd widget
npm run test
```
