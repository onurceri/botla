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

---

### 15.1.2 Localized Responses
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Turkish chatbot | Responds in Turkish |
| 2 | System prompt enforces | Language maintained |

---

### 15.1.3 Localized Error Messages
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Turkish chatbot error | Turkish error message |

---

### 15.1.4 Localized Fallback Messages
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set Turkish fallback | Custom Turkish message |
| 2 | Trigger fallback | Turkish message shown |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Turkish|Language|Localization"
```
