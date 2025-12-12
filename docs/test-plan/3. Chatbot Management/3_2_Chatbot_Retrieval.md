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

---

### 3.2.2 Filter by Organization
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set `X-Organization-ID` header | N/A |
| 2 | GET `/api/v1/chatbots` | Only org chatbots returned |

---

### 3.2.3 Filter by Workspace
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set `X-Workspace-ID` header | N/A |
| 2 | GET `/api/v1/chatbots` | Only workspace chatbots returned |

---

### 3.2.4 Get Single Chatbot
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}` | 200 OK |
| 2 | Response contains full details | All fields present |

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

---

### 3.2.6 Cannot Access Other User's Chatbot
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as User A | Token A |
| 2 | Create chatbot as User B | Chatbot B created |
| 3 | GET `/api/v1/chatbots/{chatbot_b_id}` with Token A | 403 Forbidden |

---

### 3.2.7 Invalid Chatbot ID Returns 404
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/invalid-uuid` | 404 Not Found |
| 2 | GET `/api/v1/chatbots/{non-existent-uuid}` | 404 Not Found |

---

### 3.2.8 Deleted Chatbots Not Returned
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create and delete chatbot | 200 OK |
| 2 | GET `/api/v1/chatbots` | Deleted bot not in list |
| 3 | GET `/api/v1/chatbots/{deleted_id}` | 404 Not Found |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "GetChatbot|ListChatbot"
```
