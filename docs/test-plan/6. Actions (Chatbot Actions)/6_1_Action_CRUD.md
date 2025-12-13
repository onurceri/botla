# 6.1 Action CRUD Test Plan

## Overview
This test plan covers creating, reading, updating, and deleting chatbot actions.

---

## Test Cases

### 6.1.1 Create HTTP Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/chatbots/{id}/actions` | 201 Created |
| 2 | action_type = "http" | Stored correctly |
| 3 | config includes URL, method, headers | All fields saved |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. POST with `{"action_type": "http", "name": "GetWeather", "config": {"url": "...", "method": "GET"}}`.
  2. Verify 201 Created.
  3. Verify `action_type` and `config` in response match.

---

### 6.1.2 Create Zapier Action
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST with action_type = "zapier" | 201 Created |
| 2 | config includes webhook_url | Saved |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Steps:**
  1. POST with `{"action_type": "zapier", "config": {"webhook_url": "..."}}`.
  2. Verify 201.

---

### 6.1.3 Create Builtin Action
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST with action_type = "builtin" | 201 Created |
| 2 | name identifies builtin function | Stored |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Steps:**
  1. POST with `{"action_type": "builtin", "name": "list_sources"}`.
  2. Verify 201.

---

### 6.1.4 Action Parameters Schema
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create action with JSON schema parameters | 201 Created |
| 2 | Parameters stored as JSONB | Valid JSON |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Steps:**
  1. POST action with `parameters: {"type": "object", "properties": {"city": {"type": "string"}}}`.
  2. Verify 201.
  3. Verify `parameters` field in response matches input.

---

### 6.1.5 List Actions for Chatbot
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}/actions` | 200 OK |
| 2 | Response is array | All chatbot actions |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Create 2 actions for bot.
- **Steps:**
  1. GET `/actions`.
  2. Verify array length is 2.

---

### 6.1.6 Get Single Action
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/actions/{id}` | 200 OK |
| 2 | Full details returned | All fields |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Steps:**
  1. GET `/actions/{id}`.
  2. Verify 200 OK and fields match.

---

### 6.1.7 Update Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | PUT `/api/v1/actions/{id}` | 200 OK |
| 2 | Fields updated | Changes persisted |
| 3 | updated_at updated | New timestamp |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Steps:**
  1. PUT with new `name` or `config`.
  2. Verify 200 OK.
  3. Fetch action and verify updates.

---

### 6.1.8 Delete Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | DELETE `/api/v1/actions/{id}` | 200 OK |
| 2 | Action no longer exists | 404 on GET |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Steps:**
  1. DELETE action. Verify 200/204.
  2. GET action. Verify 404.

---

### 6.1.9 Enable/Disable Toggle
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set enabled = false | 200 OK |
| 2 | Disabled action not triggered | In chat |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Steps:**
  1. PUT `{"enabled": false}`.
  2. Verify `enabled` is false.
  3. (Optional) Trigger chat query that matches action. Verify action is NOT executed in mock.

---

### 6.1.10 Ownership Validation
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create action for own chatbot | Success |
| 2 | Create action for other's chatbot | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - User A, Bot B (owned by B).
- **Steps:**
  1. Login as A.
  2. POST action to Bot B. Expect 403.

---

## How to Run Tests

```bash
go test -v ./internal/integration/action_test.go
```
