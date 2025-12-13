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

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Create bot and HTTP GET action pointing to a mock server.
- **Steps:**
  1. Send chat message.
  2. Verify mock server received GET request.
  3. Verify chat response contains data from mock server.

---

### 6.2.2 HTTP POST Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create POST action | Created |
| 2 | Trigger via chat | POST with body sent |
| 3 | Response handled | Result to user |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Create bot and HTTP POST action.
- **Steps:**
  1. Send chat message.
  2. Verify mock server received POST request with expected JSON body.

---

### 6.2.3 HTTP Action with Auth
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | auth_type = "bearer" | Token included |
| 2 | auth_type = "api_key" | API key included |
| 3 | auth_type = "none" | No auth headers |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Create 3 actions with different auth types.
- **Steps:**
  1. Trigger Bearer action. Verify `Authorization: Bearer <token>` header.
  2. Trigger API Key action. Verify `X-API-Key: <key>` header.
  3. Trigger None action. Verify no auth headers.

---

### 6.2.4 HTTP Timeout Handling
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Slow endpoint | Action triggered |
| 2 | Timeout occurs | Error handled gracefully |
| 3 | Chat continues | Not broken |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Mock server delays response by 10s (timeout 5s).
- **Steps:**
  1. Trigger action.
  2. Verify chat response indicates failure/timeout but does not crash.

---

### 6.2.5 HTTP Error Handling
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Endpoint returns 500 | Error handled |
| 2 | Endpoint returns 404 | Error handled |
| 3 | Chat continues | Not broken |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Mock server returns 500.
- **Steps:**
  1. Trigger action.
  2. Verify chat response mentions the error.

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

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Mock Zapier webhook receiver.
- **Steps:**
  1. Trigger Zapier action.
  2. Verify webhook received POST with chat context.

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

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Bot with sources.
- **Steps:**
  1. Send chat "List sources".
  2. Verify response contains list of URLs/files.

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

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Bot with handoff enabled.
- **Steps:**
  1. Trigger handoff via low confidence or explicit request.
  2. Verify response asks for email.
  3. Reply with email.
  4. Verify `handoff_requests` table has new row.

---

## AI Model Selection Tests

### 6.2.9 AI Selects Appropriate Action
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Multiple actions available | Config |
| 2 | User query matches action | Correct action selected |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Bot with 2 distinct actions ("Weather", "Stock").
- **Steps:**
  1. Ask about Weather. Verify "Weather" action called.
  2. Ask about Stock. Verify "Stock" action called.

---

### 6.2.10 Parameter Extraction
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Action expects parameters | Schema defined |
| 2 | AI extracts from query | Parameters populated |
| 3 | Action receives correct values | Execution succeeds |

**Implementation Plan:**
- **Test File:** `internal/integration/action_full_test.go`
- **Setup:**
  - Action "Weather" with param `city`.
- **Steps:**
  1. Ask "Weather in London".
  2. Verify action called with `city="London"`.

---

## How to Run Tests

```bash
go test -v ./internal/integration/action_test.go -run "Execute|Tool"
```
