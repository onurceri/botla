# 11.1 OpenRouter Integration Test Plan

## Overview
This test plan covers the OpenRouter AI provider integration.

---

## Test Cases

### 11.1.1 API Key Configuration
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | OPENROUTER_API_KEY set | App starts |
| 2 | Invalid API key | Error on first request |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_openrouter_test.go`
- **Steps:**
  1. Initialize service with valid key -> Success.
  2. Initialize service with empty key -> Expect Error/Log.

---

### 11.1.2 Chat Completion
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send chat message | OpenRouter API called |
| 2 | Response received | Valid AI response |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_openrouter_test.go`
- **Setup:**
  - Mock OpenRouter API.
- **Steps:**
  1. Call `llm.Chat(...)`.
  2. Verify mock received POST.
  3. Verify response content.

---

### 11.1.3 Model Selection
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use gpt-4o-mini | Correct model used |
| 2 | Use gpt-4o | Correct model used |
| 3 | Use claude-3-5-sonnet | Correct model used |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_openrouter_test.go`
- **Steps:**
  1. Call `llm.Chat` with `model="gpt-4o"`. Verify mock received `model="openai/gpt-4o"` (or mapped name).
  2. Repeat for others.

---

### 11.1.4 Token Tracking
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send message | Response received |
| 2 | Check tokens_used | Matches API response |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_openrouter_test.go`
- **Setup:**
  - Mock response includes `usage: {total_tokens: 42}`.
- **Steps:**
  1. Chat.
  2. Verify result struct `TokensUsed` == 42.

---

### 11.1.5 Tool Calls (Function Calling)
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Actions configured | Tools sent to API |
| 2 | AI selects tool | Tool call returned |
| 3 | Tool executed | Result sent back |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_openrouter_test.go`
- **Setup:**
  - Config with tool definitions.
  - Mock response "tool_calls".
- **Steps:**
  1. Chat.
  2. Verify service executes tool logic (if integrated) or returns tool call data.

---

### 11.1.6 Error Handling
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | API returns error | Error handled gracefully |
| 2 | User sees message | Friendly error |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_openrouter_test.go`
- **Setup:**
  - Mock returns 500.
- **Steps:**
  1. Chat.
  2. Verify error returned is wrapped/typed correctly.

---

### 11.1.7 Rate Limit Handling
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | OpenRouter rate limits | 429 handled |
| 2 | Retry or error shown | Graceful handling |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_openrouter_test.go`
- **Setup:**
  - Mock returns 429.
- **Steps:**
  1. Chat.
  2. Verify specific RateLimit error type is returned.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "OpenRouter|AI"
```
