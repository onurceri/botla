# 5.1 Chat Message Flow Test Plan

## Overview
This test plan covers the core chat functionality including message handling and token tracking.

---

## Test Cases

### 5.1.1 Send Message Successfully
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/chat` with message | 200 OK |
| 2 | Response contains `response` | AI-generated answer |
| 3 | Response contains `tokens_used` | Token count > 0 |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. POST `/api/v1/chat` with `{"message": "Hello", "chatbot_id": "..."}`.
  2. Verify 200 OK.
  3. Verify `response` is non-empty.
  4. Verify `tokens_used` > 0.

---

### 5.1.2 Sources Used in Response
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add source to chatbot | Processed |
| 2 | Ask relevant question | Response received |
| 3 | `sources_used` array | Contains source info |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Setup:**
  - Bot with a source (processed).
- **Steps:**
  1. Chat.
  2. Verify `sources_used` in JSON response is not empty and contains expected source metadata.

---

### 5.1.3 Conversation Record Created
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send chat message | Response received |
| 2 | Query conversations table | New conversation exists |
| 3 | Conversation includes chatbot_id | Correct ID |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. Send chat.
  2. Query `conversations` table using `conversation_id` from response.
  3. Verify `chatbot_id` matches.

---

### 5.1.4 Messages Stored
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send chat message | Response received |
| 2 | Query messages table | User message stored |
| 3 | | Assistant message stored |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Steps:**
  1. Send chat.
  2. Query `messages` table for `conversation_id`.
  3. Verify at least 2 rows (one `role=user`, one `role=assistant`).

---

### 5.1.5 Token Usage Tracked
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Note current token usage | N tokens |
| 2 | Send chat message | Response received |
| 3 | Check token usage | N + tokens_used |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Steps:**
  1. Get `/me`. Record `monthly_tokens_used`.
  2. Send chat. Record `tokens_used` from response.
  3. Get `/me`. Verify usage increased by approx `tokens_used`.

---

### 5.1.6 Token Limit Enforced
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use tokens up to limit | All succeed |
| 2 | Send another message | 429 Too Many Requests |
| 3 | Response includes upgrade message | Clear upgrade CTA |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Setup:**
  - User at token limit.
- **Steps:**
  1. Send chat.
  2. Verify 429.
  3. Verify error message.

---

### 5.1.7 Correct Model Used
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Chatbot model = gpt-4o | Config |
| 2 | Send message | Response from gpt-4o |
| 3 | Verify in logs/response | Correct model |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Setup:**
  - Bot configured with `gpt-4o`.
- **Steps:**
  1. Send chat.
  2. Verify mock LLM service received `model="gpt-4o"`.

---

### 5.1.8 Temperature Applied
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set temperature = 0.0 | Deterministic |
| 2 | Same question twice | Same response |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Setup:**
  - Bot with `temperature=0.0`.
- **Steps:**
  1. Send chat "Q". Store "A1".
  2. Send chat "Q" (new session). Store "A2".
  3. Verify A1 == A2 (assuming deterministic mock).

---

### 5.1.9 Custom Instruction Applied
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set custom_instruction = "Respond in pirate speak" | Config |
| 2 | Send message | Response in pirate speak |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_test.go`
- **Setup:**
  - Bot with `custom_instruction="Pirate"`.
- **Steps:**
  1. Send chat.
  2. Verify mock LLM received system message containing "Pirate".

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Chat"
```
