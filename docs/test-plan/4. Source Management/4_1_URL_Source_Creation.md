# 4.1 URL Source Creation Test Plan

## Overview
This test plan covers adding URL sources to chatbots including validation and limits.

---

## Test Cases

### 4.1.1 Add Valid URL Source
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/chatbots/{id}/sources` with valid URL | 201 Created |
| 2 | Source status | "pending" |
| 3 | Source queued for processing | Background job created |

---

### 4.1.2 Invalid URL Format
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add "not-a-url" | 400 Bad Request |
| 2 | Add empty string | 400 Bad Request |
| 3 | Add URL without protocol | 400 Bad Request |

---

### 4.1.3 URL Limit Enforcement
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: Add 1 URL | 201 Created |
| 2 | Free: Add 2nd URL | 403 Forbidden |
| 3 | Pro: Add 10 URLs | All succeed |
| 4 | Pro: Add 11th URL | 403 Forbidden |

---

### 4.1.4 Ingestion Limit Enforcement
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 50 sources this month | All succeed |
| 2 | Add 51st source | 403 Forbidden |

---

### 4.1.5 Duplicate URL Prevention
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL A | 201 Created |
| 2 | Add URL A again | 409 Conflict (duplicate) |

---

### 4.1.6 Re-add Cooldown
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL A | 201 Created |
| 2 | Delete URL A | 200 OK |
| 3 | Re-add URL A immediately | 403 Forbidden (cooldown) |
| 4 | Wait 60 minutes, re-add | 201 Created |

---

### 4.1.7 Ownership Validation
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add source to other user's bot | 403 Forbidden |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "AddSource|URLSource"
```
