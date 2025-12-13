# 3.2 Chatbot Retrieval Test Plan

## Overview
This test plan covers chatbot listing and single chatbot retrieval with authorization checks.

---

## Test Cases

### 3.2.1 List All User Chatbots
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots` | 200 OK |
| 2 | Response is array | Contains user's chatbots |
| 3 | Does not include other users' bots | Only owned bots |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create User A and User B.
  - Create 2 bots for A, 1 bot for B.
- **Steps:**
  1. Login as A. Call `GET /api/v1/chatbots`.
  2. Verify response is a JSON array of length 2.
  3. Verify IDs match A's bots.

---

### 3.2.2 Filter by Organization
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set `X-Organization-ID` header | N/A |
| 2 | GET `/api/v1/chatbots` | Only org chatbots returned |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create User with Org O1.
  - Create Bot 1 in personal context.
  - Create Bot 2 in Org O1.
- **Steps:**
  1. Call `GET /api/v1/chatbots` with `X-Organization-ID: {O1_ID}`.
  2. Verify response contains only Bot 2.

---

### 3.2.3 Filter by Workspace
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set `X-Workspace-ID` header | N/A |
| 2 | GET `/api/v1/chatbots` | Only workspace chatbots returned |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create User with Workspace W1 in Org O1.
  - Create Bot 1 in O1 (no workspace).
  - Create Bot 2 in W1.
- **Steps:**
  1. Call `GET /api/v1/chatbots` with `X-Organization-ID: {O1_ID}` and `X-Workspace-ID: {W1_ID}`.
  2. Verify response contains only Bot 2.

---

### 3.2.4 Get Single Chatbot
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}` | 200 OK |
| 2 | Response contains full details | All fields present |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a bot.
- **Steps:**
  1. Call `GET /api/v1/chatbots/{id}`.
  2. Verify 200 OK and JSON body matches created bot.

---

### 3.2.5 Response Includes Configuration Fields
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}` | 200 OK |
| 2 | Response includes `custom_branding` | If set |
| 3 | Response includes `threshold_config` | If set |
| 4 | Response includes `fallback_messages` | If set |
| 5 | Response includes `topic_restrictions` | If set |
| 6 | Response includes `handoff_config` | If enabled |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a bot with `handoff_enabled=true`.
- **Steps:**
  1. Call `GET /api/v1/chatbots/{id}`.
  2. Verify `handoff_config` field is present in JSON.

---

### 3.2.6 Cannot Access Other User's Chatbot
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as User A | Token A |
| 2 | Create chatbot as User B | Chatbot B created |
| 3 | GET `/api/v1/chatbots/{chatbot_b_id}` with Token A | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create User A and User B.
  - Create Bot B (owned by B).
- **Steps:**
  1. Login as A.
  2. Call `GET /api/v1/chatbots/{bot_b_id}`.
  3. Verify `403 Forbidden`.

---

### 3.2.7 Invalid Chatbot ID Returns 404
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/invalid-uuid` | 404 Not Found |
| 2 | GET `/api/v1/chatbots/{non-existent-uuid}` | 404 Not Found |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a user.
- **Steps:**
  1. Call `GET /api/v1/chatbots/not-a-uuid`. Expect `400` or `404`.
  2. Call `GET /api/v1/chatbots/{random_uuid}`. Expect `404`.

---

### 3.2.8 Deleted Chatbots Not Returned
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create and delete chatbot | 200 OK |
| 2 | GET `/api/v1/chatbots` | Deleted bot not in list |
| 3 | GET `/api/v1/chatbots/{deleted_id}` | 404 Not Found |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_lifecycle_test.go`
- **Setup:**
  - Create a bot.
  - Delete it.
- **Steps:**
  1. Call `GET /api/v1/chatbots`. Verify list is empty.
  2. Call `GET /api/v1/chatbots/{id}`. Verify `404 Not Found`.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "GetChatbot|ListChatbot"
```
