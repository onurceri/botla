# 3.4 Chatbot Deletion Test Plan

## Overview
This test plan covers chatbot deletion including cascade effects and authorization.

---

## Test Cases

### 3.4.1 Delete Chatbot Successfully
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | DELETE `/api/v1/chatbots/{id}` | 200 OK |
| 2 | GET `/api/v1/chatbots/{id}` | 404 Not Found |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create a bot.
- **Steps:**
  1. Send `DELETE /api/v1/chatbots/{id}`. Verify 200 OK.
  2. Send `GET /api/v1/chatbots/{id}`. Verify 404 Not Found.

---

### 3.4.2 Soft Delete Sets Timestamp
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Delete chatbot | 200 OK |
| 2 | Query database directly | deleted_at is set |
| 3 | Record still exists | But excluded from queries |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create a bot.
- **Steps:**
  1. Send `DELETE /api/v1/chatbots/{id}`.
  2. Query DB: `SELECT deleted_at FROM chatbots WHERE id = $1`.
  3. Verify `deleted_at` is not null.

---

### 3.4.3 Cascade Delete Sources
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add sources to chatbot | Sources exist |
| 2 | Delete chatbot | 200 OK |
| 3 | Query sources | Sources deleted/inaccessible |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create bot with a source.
- **Steps:**
  1. Delete bot.
  2. Query DB for source: `SELECT deleted_at FROM data_sources`. Verify it's set.
  3. Or try `GET /api/v1/sources/{source_id}` -> Expect 404.

---

### 3.4.4 Cascade Delete Conversations
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Chat with chatbot | Conversations exist |
| 2 | Delete chatbot | 200 OK |
| 3 | Query conversations | Conversations deleted |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create bot, create conversation.
- **Steps:**
  1. Delete bot.
  2. Query DB for conversation. Verify it's soft-deleted (or hard deleted based on schema).

---

### 3.4.5 Cascade Delete Actions
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add actions to chatbot | Actions exist |
| 2 | Delete chatbot | 200 OK |
| 3 | Query actions | Actions deleted |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create bot, add action.
- **Steps:**
  1. Delete bot.
  2. Query DB for action. Verify deleted.

---

### 3.4.6 Analytics Preserved After Delete
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Chat with chatbot (generates analytics) | Analytics recorded |
| 2 | Delete chatbot | 200 OK |
| 3 | Query analytics | Analytics preserved for reporting |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create bot, generate analytics data (insert to DB).
- **Steps:**
  1. Delete bot.
  2. Query analytics table for `chatbot_id`.
  3. Verify records still exist.

---

### 3.4.7 Cannot Delete Other User's Chatbot
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as User A | Token A |
| 2 | DELETE `/api/v1/chatbots/{user_b_bot}` | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - User A, User B with Bot B.
- **Steps:**
  1. Login as A.
  2. Delete Bot B. Expect 403.

---

### 3.4.8 Invalid ID Returns 404
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | DELETE `/api/v1/chatbots/non-existent-uuid` | 404 Not Found |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Login.
- **Steps:**
  1. Delete `random-uuid`. Expect 404.

---

### 3.4.9 Cannot Chat with Deleted Chatbot
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Delete chatbot | 200 OK |
| 2 | POST `/api/v1/chat` with deleted bot ID | 404 Not Found |

**Implementation Plan:**
- **Test File:** `internal/integration/chatbot_update_delete_test.go`
- **Setup:**
  - Create and delete bot.
- **Steps:**
  1. POST `/api/v1/chatbots/{id}/chat`.
  2. Verify 404 Not Found.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "DeleteChatbot"
```
