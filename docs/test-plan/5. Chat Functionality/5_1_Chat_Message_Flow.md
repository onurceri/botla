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

---

### 5.1.2 Sources Used in Response
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add source to chatbot | Processed |
| 2 | Ask relevant question | Response received |
| 3 | `sources_used` array | Contains source info |

---

### 5.1.3 Conversation Record Created
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send chat message | Response received |
| 2 | Query conversations table | New conversation exists |
| 3 | Conversation includes chatbot_id | Correct ID |

---

### 5.1.4 Messages Stored
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send chat message | Response received |
| 2 | Query messages table | User message stored |
| 3 | | Assistant message stored |

---

### 5.1.5 Token Usage Tracked
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Note current token usage | N tokens |
| 2 | Send chat message | Response received |
| 3 | Check token usage | N + tokens_used |

---

### 5.1.6 Token Limit Enforced
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use tokens up to limit | All succeed |
| 2 | Send another message | 429 Too Many Requests |
| 3 | Response includes upgrade message | Clear upgrade CTA |

---

### 5.1.7 Correct Model Used
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Chatbot model = gpt-4o | Config |
| 2 | Send message | Response from gpt-4o |
| 3 | Verify in logs/response | Correct model |

---

### 5.1.8 Temperature Applied
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set temperature = 0.0 | Deterministic |
| 2 | Same question twice | Same response |

---

### 5.1.9 Custom Instruction Applied
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set custom_instruction = "Respond in pirate speak" | Config |
| 2 | Send message | Response in pirate speak |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Chat"
```
