# 2.1 Free Plan Limits Test Plan

## Overview
This test plan verifies all Free plan restrictions are properly enforced both on backend and frontend.

---

## Test Cases

10→### 2.1.1 Free Plan Default Assignment [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Register new user | 201 Created |
| 2 | GET `/api/v1/me` | User has plan.code = "free" |

---

21→### 2.1.2 Model Restriction - Only gpt-4o-mini [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot with model `gpt-4o-mini` | 201 Created |
| 2 | Update chatbot to model `gpt-4o` | 403 Forbidden |
| 3 | Update chatbot to model `claude-3-5-sonnet` | 403 Forbidden |

---

### 2.1.3 Monthly Token Limit - 100,000 [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use tokens up to 99,999 | Chat succeeds |
| 2 | Use chat that exceeds 100,000 total | 402 Payment Required |
| 3 | Response includes upgrade message | Contains plan upgrade info |

---

### 2.1.4 RAG Configuration Limits
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Verify plan config | top_k = 3, max_context_tokens = 2000 |
| 2 | Chat uses correct Top-K | Only 3 chunks retrieved |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot for this user.
  - Mock the Qdrant retrieval response to return more than 3 chunks (e.g., 5).
- **Steps:**
  1. Send a chat message to `/api/v1/chatbots/{id}/chat`.
  2. In the mock Qdrant handler, verify that the search query parameter `limit` (or `top_k`) passed from the backend is `3`.
  3. Verify the chat response generation uses only the top 3 chunks (check context passed to LLM mock).

---

### 2.1.5 PDF File Limits [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload 1 PDF (< 5MB) | 201 Created |
| 2 | Upload 2nd PDF to same chatbot | 403 Forbidden (max_files_per_bot = 1) |
| 3 | Upload PDF > 5MB | 413 Payload Too Large |

---

### 2.1.6 Storage Limit - 10MB Total [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload files totaling 10MB | Success |
| 2 | Upload additional file | 402 Payment Required (storage exceeded) |

---

### 2.1.7 OCR Disabled
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload image-based PDF | Source created |
| 2 | Check extracted text | No OCR text extracted (images skipped) |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot.
  - Prepare a sample PDF that contains only images (no selectable text).
- **Steps:**
  1. Upload the PDF via `POST /api/v1/chatbots/{id}/sources`.
  2. Wait for processing to complete.
  3. Query the source content (or check debug logs/internal state).
  4. Verify the `content` field is empty or contains a standard "no text found" message, confirming OCR was not triggered.

---

### 2.1.8 URL Limits [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 1 URL source | 201 Created |
| 2 | Add 2nd URL source | 403 Forbidden (max_urls_per_bot = 1) |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot.
- **Steps:**
  1. Add a URL source `http://example.com/1` -> Expect `201 Created`.
  2. Add a second URL source `http://example.com/2` -> Expect `403 Forbidden`.
  3. Verify the error message indicates the URL limit has been reached.

---

### 2.1.9 Dynamic Scraping Disabled
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL requiring JavaScript rendering | Source created |
| 2 | Check scraped content | Static HTML only (no JS-rendered content) |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Setup a mock HTML server that returns different content for static HTTP requests vs. JS-enabled browsers (or simply verify the scraping flag in the internal call).
- **Steps:**
  1. Add a URL source.
  2. During processing, intercept the scraping call or check the `is_dynamic` flag passed to the scraper service.
  3. Verify `is_dynamic` is `false`.

---

### 2.1.10 Discovery Mode Disabled [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to set discovery_mode = "auto" | 403 Forbidden |
| 2 | Verify max_pages_per_crawl = 0 | No sub-pages discovered |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot.
- **Steps:**
  1. Send `PUT /api/v1/chatbots/{id}` with `{"discovery_mode": "auto"}`.
  2. Expect `403 Forbidden`.
  3. Send `POST` to create a chatbot with `{"discovery_mode": "auto"}`.
  4. Expect `403 Forbidden` (or fallback to disabled with a warning, depending on implementation strictness).

---

### 2.1.11 Refresh Disabled [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to refresh URL source | 403 Forbidden |
| 2 | Try to set refresh_policy = "auto" | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot and add a URL source.
- **Steps:**
  1. Send `POST /api/v1/sources/{id}/refresh`.
  2. Expect `403 Forbidden`.
  3. Send `PUT /api/v1/chatbots/{id}` with `{"refresh_policy": "auto"}`.
  4. Expect `403 Forbidden`.

---

### 2.1.12 Branding Cannot Be Hidden [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to update hide_branding = true | 403 Forbidden |
| 2 | Widget displays "Powered by Botla" | Branding visible |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot.
- **Steps:**
  1. Send `PUT /api/v1/chatbots/{id}` with `{"custom_branding": {"hide_branding": true}}` (or flattened `hide_branding` field).
  2. Expect `403 Forbidden`.
  3. Retrieve chatbot config and verify `hide_branding` remains `false`.

---

### 2.1.13 Secure Embed Disabled [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to update secure_embed_enabled = true | 403 Forbidden |
| 2 | Try to set allowed_domains | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot.
- **Steps:**
  1. Send `PUT /api/v1/chatbots/{id}` with `{"secure_embed_enabled": true}`.
  2. Expect `403 Forbidden`.
  3. Send `PUT` with `{"allowed_domains": ["example.com"]}`.
  4. Expect `403 Forbidden`.

---

### 2.1.14 Ingestion Limits
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 50 sources in a month | All succeed |
| 2 | Add 51st source | 402 Payment Required (max_monthly_ingestions) |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Manually update `user_stats` or `analytics` table to set `monthly_ingestions = 50`.
- **Steps:**
  1. Attempt to add a new source.
  2. Expect `402 Payment Required` (or `403 Forbidden` with limit message).

---

### 2.1.15 Re-add Cooldown - 60 Minutes
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL source | 201 Created |
| 2 | Delete source | 200 OK |
| 3 | Re-add same URL immediately | 429 Too Many Requests (cooldown) |
| 4 | Wait 60 minutes, re-add | 201 Created |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot.
- **Steps:**
  1. Add a URL source `http://cool-down-test.com`.
  2. Delete the source.
  3. Immediately try to add `http://cool-down-test.com` again.
  4. Expect `429 Too Many Requests` (or `403 Forbidden` with cooldown message).
  5. (Optional) Manually update the `deleted_at` timestamp of the previous source to > 60 mins ago, then retry. Expect `201 Created`.

---

### 2.1.16 Guardrails Restricted [x]
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to customize thresholds | 403 Forbidden |
| 2 | Try to use fallback_mode = "smart" | 403 Forbidden |
| 3 | Try to use fallback_mode = "escalate" | 403 Forbidden |
| 4 | Try to manage topic restrictions | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_limits_free_test.go`
- **Setup:**
  - Create a user on the `free` plan.
  - Create a chatbot.
- **Steps:**
  1. Send `PUT` with `{"threshold_config": {"high": 0.9}}`. Expect `403 Forbidden`.
  2. Send `PUT` with `{"fallback_mode": "smart"}`. Expect `403 Forbidden`.
  3. Send `PUT` with `{"topic_restrictions": {"allowed_topics": ["tech"]}}`. Expect `403 Forbidden`.

---

## How to Run Tests

```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/... -run "FreePlan|PlanEnforcement"
```

---

## Coverage Notes
- Many enforcement tests exist in `internal/integration/` directory
- Consider grouped test suite for all Free plan limits
