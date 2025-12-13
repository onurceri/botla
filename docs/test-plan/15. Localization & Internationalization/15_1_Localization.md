# 15.1 Localization Test Plan

## Overview
This test plan covers language support and localization.

---

## Test Cases

### 15.1.1 Chatbot Language Setting
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set language_code = "tr" | Turkish |
| 2 | Set language_code = "en" | English |

**Implementation Plan:**
- **Test File:** `internal/integration/localization_test.go`
- **Steps:**
  1. POST bot with `language_code="tr"`. Verify 201.
  2. POST bot with `language_code="en"`. Verify 201.

---

### 15.1.2 Localized Responses
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Turkish chatbot | Responds in Turkish |
| 2 | System prompt enforces | Language maintained |

**Implementation Plan:**
- **Test File:** `internal/integration/localization_test.go`
- **Setup:**
  - Create `tr` bot.
- **Steps:**
  1. Send chat.
  2. Verify mock LLM received system message containing "Turkish".

---

### 15.1.3 Localized Error Messages
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Turkish chatbot error | Turkish error message |

**Implementation Plan:**
- **Test File:** `internal/integration/localization_test.go`
- **Setup:**
  - Create `tr` bot.
- **Steps:**
  1. Trigger error (e.g. timeout).
  2. Verify user-facing message is in Turkish (if implemented).

---

### 15.1.4 Localized Fallback Messages
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set Turkish fallback | Custom Turkish message |
| 2 | Trigger fallback | Turkish message shown |

**Implementation Plan:**
- **Test File:** `internal/integration/localization_test.go`
- **Setup:**
  - Create `tr` bot.
- **Steps:**
  1. Trigger "no sources" condition.
  2. Verify response matches default Turkish fallback message.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Turkish|Language|Localization"
```
