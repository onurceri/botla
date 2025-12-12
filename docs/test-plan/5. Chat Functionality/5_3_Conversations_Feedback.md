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

---

### 5.3.2 Messages Linked to Conversation
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send 3 messages | All stored |
| 2 | Query by conversation_id | All 3 messages returned |

---

### 5.3.3 Conversation Includes Chatbot ID
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create conversation | Record created |
| 2 | Verify chatbot_id | Correctly set |

---

### 5.3.4 List Conversations for Chatbot
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}/conversations` | 200 OK |
| 2 | Response is array | Contains conversations |
| 3 | Paginated | Supports limit/offset |

---

### 5.3.5 Get Conversation Messages
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/conversations/{id}` | 200 OK |
| 2 | Response includes messages | All messages |
| 3 | Messages in order | Chronological |

---

## Feedback Test Cases

### 5.3.6 Submit Positive Feedback
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST feedback with type = "thumbs_up" | 200 OK |
| 2 | Feedback stored | In message or feedback table |

---

### 5.3.7 Submit Negative Feedback
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST feedback with type = "thumbs_down" | 200 OK |
| 2 | Feedback stored | Record created |

---

### 5.3.8 Feedback with Comment
**Priority:** Low  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST feedback with comment | 200 OK |
| 2 | Comment stored | Text preserved |

---

### 5.3.9 Update Feedback
**Priority:** Low  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit thumbs_up | Stored |
| 2 | Update to thumbs_down | Updated |

---

### 5.3.10 Feedback in Analytics
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit feedback | Stored |
| 2 | GET analytics | Feedback counts included |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Conversation|Feedback"
```
