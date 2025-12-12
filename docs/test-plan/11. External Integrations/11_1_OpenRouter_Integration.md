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

---

### 11.1.2 Chat Completion
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send chat message | OpenRouter API called |
| 2 | Response received | Valid AI response |

---

### 11.1.3 Model Selection
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use gpt-4o-mini | Correct model used |
| 2 | Use gpt-4o | Correct model used |
| 3 | Use claude-3-5-sonnet | Correct model used |

---

### 11.1.4 Token Tracking
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send message | Response received |
| 2 | Check tokens_used | Matches API response |

---

### 11.1.5 Tool Calls (Function Calling)
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Actions configured | Tools sent to API |
| 2 | AI selects tool | Tool call returned |
| 3 | Tool executed | Result sent back |

---

### 11.1.6 Error Handling
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | API returns error | Error handled gracefully |
| 2 | User sees message | Friendly error |

---

### 11.1.7 Rate Limit Handling
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | OpenRouter rate limits | 429 handled |
| 2 | Retry or error shown | Graceful handling |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "OpenRouter|AI"
```
