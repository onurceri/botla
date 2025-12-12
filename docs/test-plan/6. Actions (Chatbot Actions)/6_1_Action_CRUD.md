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

---

### 6.1.2 Create Zapier Action
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST with action_type = "zapier" | 201 Created |
| 2 | config includes webhook_url | Saved |

---

### 6.1.3 Create Builtin Action
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST with action_type = "builtin" | 201 Created |
| 2 | name identifies builtin function | Stored |

---

### 6.1.4 Action Parameters Schema
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create action with JSON schema parameters | 201 Created |
| 2 | Parameters stored as JSONB | Valid JSON |

---

### 6.1.5 List Actions for Chatbot
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}/actions` | 200 OK |
| 2 | Response is array | All chatbot actions |

---

### 6.1.6 Get Single Action
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/actions/{id}` | 200 OK |
| 2 | Full details returned | All fields |

---

### 6.1.7 Update Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | PUT `/api/v1/actions/{id}` | 200 OK |
| 2 | Fields updated | Changes persisted |
| 3 | updated_at updated | New timestamp |

---

### 6.1.8 Delete Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | DELETE `/api/v1/actions/{id}` | 200 OK |
| 2 | Action no longer exists | 404 on GET |

---

### 6.1.9 Enable/Disable Toggle
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set enabled = false | 200 OK |
| 2 | Disabled action not triggered | In chat |

---

### 6.1.10 Ownership Validation
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create action for own chatbot | Success |
| 2 | Create action for other's chatbot | 403 Forbidden |

---

## How to Run Tests

```bash
go test -v ./internal/integration/action_test.go
```
