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

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a user.
- **Steps:**
  1. Send `POST /api/v1/chatbots` with `{"name": "Basic Bot"}`.
  2. Verify response status `201`.
  3. Verify `id` is present in response.
  4. Query DB for this `id` to confirm persistence.

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

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a user.
- **Steps:**
  1. Send `POST /api/v1/chatbots` with minimal payload `{"name": "Defaults Bot"}`.
  2. Inspect response fields.
  3. Verify `temperature` is `0.7` (or system default).
  4. Verify `position` is `bottom-right`.
  5. Verify `refresh_policy` is `manual` or `null`.

---

### 3.1.3 Model Validation Against Plan
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free user creates with `gpt-4o-mini` | 201 Created |
| 2 | Free user creates with `gpt-4o` | 403 or coerced to gpt-4o-mini |
| 3 | Pro user creates with `gpt-4o` | 201 Created |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a `free` user and a `pro` user.
- **Steps:**
  1. Free user: create with `gpt-4o-mini`. Expect `201`.
  2. Free user: create with `gpt-4o`. Expect `403` or coercion (verify `model` in response is `gpt-4o-mini`).
  3. Pro user: create with `gpt-4o`. Expect `201` and `model` is `gpt-4o`.

---

### 3.1.4 Organization Context Assignment
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot without org context | Assigned to user's personal context |
| 2 | Create with `organization_id` header | Assigned to organization |
| 3 | Create with `workspace_id` | Assigned to workspace |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a user who owns an Organization and a Workspace.
- **Steps:**
  1. Create bot without headers. Verify `organization_id` is null (or personal org ID).
  2. Create bot with `X-Organization-ID`. Verify `organization_id` matches.
  3. Create bot with `X-Workspace-ID`. Verify `workspace_id` matches.

---

### 3.1.5 Max Chatbots Limit
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbots up to limit | All succeed |
| 2 | Create one more | 403 Forbidden (limit exceeded) |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a user with a plan limit of 1 chatbot.
- **Steps:**
  1. Create chatbot 1. Expect `201`.
  2. Create chatbot 2. Expect `403`.

---

### 3.1.6 Timestamps Set Correctly
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot | 201 Created |
| 2 | Verify created_at | Current timestamp |
| 3 | Verify updated_at | Same as created_at |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a user.
- **Steps:**
  1. Create chatbot.
  2. Verify `created_at` is close to `time.Now()`.
  3. Verify `updated_at` equals `created_at`.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "CreateChatbot"
```
