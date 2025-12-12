# 6.2 Action Execution Test Plan

## Overview
This test plan covers the execution of actions during chat conversations.

---

## HTTP Action Tests

### 6.2.1 HTTP GET Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create GET action | Created |
| 2 | Trigger via chat | HTTP GET sent |
| 3 | Response returned to user | Result included |

---

### 6.2.2 HTTP POST Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create POST action | Created |
| 2 | Trigger via chat | POST with body sent |
| 3 | Response handled | Result to user |

---

### 6.2.3 HTTP Action with Auth
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | auth_type = "bearer" | Token included |
| 2 | auth_type = "api_key" | API key included |
| 3 | auth_type = "none" | No auth headers |

---

### 6.2.4 HTTP Timeout Handling
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Slow endpoint | Action triggered |
| 2 | Timeout occurs | Error handled gracefully |
| 3 | Chat continues | Not broken |

---

### 6.2.5 HTTP Error Handling
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Endpoint returns 500 | Error handled |
| 2 | Endpoint returns 404 | Error handled |
| 3 | Chat continues | Not broken |

---

## Zapier Action Tests

### 6.2.6 Zapier Webhook Trigger
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create Zapier action | Created |
| 2 | Trigger via chat | Webhook called |
| 3 | Payload sent | Correct data |

---

## Built-in Tools Tests

### 6.2.7 list_sources Tool
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Chat asks "what sources do you have?" | Tool triggered |
| 2 | list_sources executed | Sources returned |
| 3 | Response to user | Source list displayed |

---

### 6.2.8 request_human_handoff Tool
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Low confidence + escalate mode | Handoff triggered |
| 2 | request_human_handoff tool called | Handoff flow starts |
| 3 | User prompted for email | Email collection |
| 4 | Handoff request created | Record in database |

---

## AI Model Selection Tests

### 6.2.9 AI Selects Appropriate Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Multiple actions available | Config |
| 2 | User query matches action | Correct action selected |

---

### 6.2.10 Parameter Extraction
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Action expects parameters | Schema defined |
| 2 | AI extracts from query | Parameters populated |
| 3 | Action receives correct values | Execution succeeds |

---

## How to Run Tests

```bash
go test -v ./internal/integration/action_test.go -run "Execute|Tool"
```
