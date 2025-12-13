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

**Implementation Plan:**
- **Test File:** `internal/integration/source_url_discovery_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. POST `{"source_type": "url", "source_url": "http://example.com"}`.
  2. Verify 201.
  3. Verify JSON body has `status: "pending"`.

---

### 4.1.2 Invalid URL Format
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add "not-a-url" | 400 Bad Request |
| 2 | Add empty string | 400 Bad Request |
| 3 | Add URL without protocol | 400 Bad Request |

**Implementation Plan:**
- **Test File:** `internal/integration/source_url_discovery_test.go`
- **Steps:**
  1. POST "not-a-url". Expect 400.
  2. POST "". Expect 400.
  3. POST "example.com" (no scheme). Expect 400.

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

**Implementation Plan:**
- **Test File:** `internal/integration/source_url_discovery_test.go`
- **Setup:**
  - Free User, Pro User.
- **Steps:**
  1. Free: Add 1 URL -> 201. Add 2nd -> 403.
  2. Pro: Add 10 URLs -> 201. Add 11th -> 403.

---

### 4.1.4 Ingestion Limit Enforcement
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 50 sources this month | All succeed |
| 2 | Add 51st source | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/source_url_discovery_test.go`
- **Setup:**
  - User with manually incremented usage.
- **Steps:**
  1. Attempt to add source. Expect 403/402.

---

### 4.1.5 Duplicate URL Prevention
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL A | 201 Created |
| 2 | Add URL A again | 409 Conflict (duplicate) |

**Implementation Plan:**
- **Test File:** `internal/integration/source_url_discovery_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. Add `http://example.com`. Expect 201.
  2. Add `http://example.com`. Expect 409.

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

**Implementation Plan:**
- **Test File:** `internal/integration/source_url_discovery_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. Add `http://cd.com`.
  2. Delete it.
  3. Add `http://cd.com` again. Expect 429/403.

---

### 4.1.7 Ownership Validation
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add source to other user's bot | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/source_url_discovery_test.go`
- **Setup:**
  - User A, Bot B (owned by User B).
- **Steps:**
  1. Login as A.
  2. POST source to Bot B. Expect 403.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "AddSource|URLSource"
```
