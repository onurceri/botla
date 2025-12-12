# 3.1 Chatbot Creation Test Plan

## Overview
This test plan covers chatbot creation including validation, defaults, and context assignment.

---

## Test Cases

### 3.1.1 Create Chatbot with Required Fields
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/chatbots` with name | 201 Created |
| 2 | Response contains chatbot ID | UUID format |
| 3 | Chatbot exists in database | Record found |

---

### 3.1.2 Chatbot Defaults Set Correctly
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot with minimal fields | 201 Created |
| 2 | Verify defaults | custom_instruction = "" |
| 3 | | welcome_message = default |
| 4 | | temperature = 0.7 |
| 5 | | position = "bottom-right" |
| 6 | | refresh_policy = "manual" |
| 7 | | discovery_mode = "disabled" |

---

### 3.1.3 Model Validation Against Plan
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free user creates with `gpt-4o-mini` | 201 Created |
| 2 | Free user creates with `gpt-4o` | 403 or coerced to gpt-4o-mini |
| 3 | Pro user creates with `gpt-4o` | 201 Created |

---

### 3.1.4 Organization Context Assignment
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot without org context | Assigned to user's personal context |
| 2 | Create with `organization_id` header | Assigned to organization |
| 3 | Create with `workspace_id` | Assigned to workspace |

---

### 3.1.5 Max Chatbots Limit
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbots up to limit | All succeed |
| 2 | Create one more | 403 Forbidden (limit exceeded) |

---

### 3.1.6 Timestamps Set Correctly
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot | 201 Created |
| 2 | Verify created_at | Current timestamp |
| 3 | Verify updated_at | Same as created_at |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "CreateChatbot"
```
