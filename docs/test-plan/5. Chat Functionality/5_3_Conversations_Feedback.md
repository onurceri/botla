# 5.3 Conversations & Feedback Test Plan

## Overview
This test plan covers conversation management and user feedback functionality.

---

## Test Cases

### 5.3.1 Unique Conversation ID
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Start chat session | conversation_id returned |
| 2 | Send multiple messages | Same conversation_id |
| 3 | New session | Different conversation_id |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. POST chat (no session_id). Store `conversation_id` as C1.
  2. POST chat (session_id=C1). Verify response `conversation_id` == C1.
  3. POST chat (no session_id). Store `conversation_id` as C2.
  4. Verify C1 != C2.

---

### 5.3.2 Messages Linked to Conversation
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send 3 messages | All stored |
| 2 | Query by conversation_id | All 3 messages returned |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Setup:**
  - Create conversation.
- **Steps:**
  1. Send 3 chat messages in loop.
  2. Query `messages` table where `conversation_id` = C1.
  3. Verify count >= 6 (3 user + 3 bot).

---

### 5.3.3 Conversation Includes Chatbot ID
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create conversation | Record created |
| 2 | Verify chatbot_id | Correctly set |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Steps:**
  1. Create conversation.
  2. Fetch conversation via API or DB.
  3. Verify `chatbot_id` matches bot ID.

---

### 5.3.4 List Conversations for Chatbot
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}/conversations` | 200 OK |
| 2 | Response is array | Contains conversations |
| 3 | Paginated | Supports limit/offset |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Setup:**
  - Create 15 conversations.
- **Steps:**
  1. GET `/conversations?limit=10`. Verify 10 items.
  2. GET `/conversations?limit=10&offset=10`. Verify 5 items.

---

### 5.3.5 Get Conversation Messages
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/conversations/{id}` | 200 OK |
| 2 | Response includes messages | All messages |
| 3 | Messages in order | Chronological |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Setup:**
  - Create conversation with messages.
- **Steps:**
  1. GET `/conversations/{id}`.
  2. Verify `messages` array exists and is ordered by `created_at`.

---

## Feedback Test Cases

### 5.3.6 Submit Positive Feedback
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST feedback with type = "thumbs_up" | 200 OK |
| 2 | Feedback stored | In message or feedback table |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Setup:**
  - Create chat message. Get `message_id`.
- **Steps:**
  1. POST `/api/v1/messages/{id}/feedback` with `{"type": "thumbs_up"}`.
  2. Verify 200 OK.
  3. Verify DB record.

---

### 5.3.7 Submit Negative Feedback
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST feedback with type = "thumbs_down" | 200 OK |
| 2 | Feedback stored | Record created |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Steps:**
  1. POST `/api/v1/messages/{id}/feedback` with `{"type": "thumbs_down"}`.
  2. Verify 200 OK.

---

### 5.3.8 Feedback with Comment
**Priority:** Low  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST feedback with comment | 200 OK |
| 2 | Comment stored | Text preserved |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Steps:**
  1. POST with `{"type": "thumbs_down", "comment": "Not helpful"}`.
  2. Verify comment is stored in DB.

---

### 5.3.9 Update Feedback
**Priority:** Low  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit thumbs_up | Stored |
| 2 | Update to thumbs_down | Updated |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Steps:**
  1. POST `thumbs_up`.
  2. POST `thumbs_down`.
  3. Verify DB shows `thumbs_down`.

---

### 5.3.10 Feedback in Analytics
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit feedback | Stored |
| 2 | GET analytics | Feedback counts included |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_feedback_test.go`
- **Steps:**
  1. Submit 1 up, 1 down.
  2. GET `/api/v1/chatbots/{id}/analytics`.
  3. Verify `thumbs_up_count=1`, `thumbs_down_count=1`.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Conversation|Feedback"
```
