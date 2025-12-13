# 2.2 Pro Plan Features Test Plan

## Overview
This test plan verifies all Pro plan features are available and properly configured.

---

## Pro Plan Specifications

| Feature | Limit |
|---------|-------|
| Allowed Models | gpt-4o-mini, gpt-4o |
| Monthly Tokens | 1,000,000 |
| RAG Top-K | 5 |
| RAG Context | 4,000 tokens |
| Files per Bot | 20 |
| File Size | 20MB |
| Total Storage | 500MB |
| OCR | Enabled |
| URLs per Bot | 10 |
| Dynamic Scraping | Enabled |
| Discovery Mode | Max 10 pages |
| Refresh | Manual & Auto |
| Hide Branding | Yes |
| Custom Branding | No |
| Secure Embed | Yes |

---

## Test Cases

### [x] 2.2.1 Model Selection - gpt-4o Available
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot with model `gpt-4o-mini` | 201 Created |
| 2 | Update to model `gpt-4o` | 200 OK |
| 3 | Update to model `claude-3-5-sonnet` | 403 Forbidden (Ultra only) |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
  - Create a chatbot.
- **Steps:**
  1. Update chatbot to `gpt-4o`. Expect `200 OK`.
  2. Update chatbot to `claude-3-5-sonnet`. Expect `403 Forbidden`.

---

### [x] 2.2.2 Monthly Token Limit - 1,000,000
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use tokens up to 999,999 | Chat succeeds |
| 2 | Exceed 1,000,000 total | 429 Too Many Requests |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
  - Use `db.IncrementAnalytics` to simulate usage of 999,999 tokens.
- **Steps:**
  1. Send a chat message (10 tokens). Expect `200 OK`.
  2. Use `db.IncrementAnalytics` to exceed 1M tokens.
  3. Send a chat message. Expect `429 Too Many Requests`.

---

### [x] 2.2.3 PDF Limits - 20 per Bot, 20MB each
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload 20 PDFs (< 20MB each) | All succeed |
| 2 | Upload 21st PDF | 403 Forbidden |
| 3 | Upload PDF > 20MB | 413 Payload Too Large |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
- **Steps:**
  1. Loop to upload 20 small PDF files. Expect `201 Created` for all.
  2. Upload 21st PDF. Expect `403 Forbidden`.
  3. Upload a dummy 21MB file (can be sparse file or mock size check). Expect `413 Payload Too Large`.

---

### [-] 2.2.4 OCR Enabled
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload image-based PDF | Source created |
| 2 | Check extracted text | OCR text extracted from images |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
  - Prepare an image-only PDF.
- **Steps:**
  1. Upload the PDF.
  2. Verify that the system attempts to perform OCR (check internal flag or result content if mock OCR is active).
  3. Verify extracted text is not empty.

---

### [x] 2.2.5 URL Limits - 10 per Bot
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 10 URL sources | All succeed |
| 2 | Add 11th URL | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
  - Create a chatbot.
- **Steps:**
  1. Loop to add 10 unique URLs. Expect `201 Created` for all.
  2. Add 11th URL. Expect `403 Forbidden`.

---

### [-] 2.2.6 Dynamic Scraping Enabled
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| [x] 1 | Add URL with JS-rendered content | Source created |
| [ ] 2 | Check scraped content | JS-rendered content included |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
- **Steps:**
  1. Add a URL source.
  2. Verify in the scraper mock (or internal logic check) that `is_dynamic` or `render_js` is passed as `true`.

---

### [x] 2.2.7 Discovery Mode - Max 10 Pages
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set discovery_mode = "auto" | 200 OK |
| 2 | Add URL with 15 sub-pages | Only 10 pages discovered |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
  - Create a chatbot with `discovery_mode="auto"`.
  - Mock a crawler that returns 15 links on the seed page.
- **Steps:**
  1. Add the seed URL.
  2. Wait for processing.
  3. Verify that exactly 11 sources exist (1 seed + 10 discovered), or 10 if seed counts towards limit (check logic). Assuming max 10 *discovered*.

---

### [x] 2.2.8 Refresh Available
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Manually refresh URL source | 200 OK |
| 2 | Set refresh_policy = "auto" | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
  - Create a chatbot and URL source.
- **Steps:**
  1. Send `POST /api/v1/sources/{id}/refresh`. Expect `200 OK`.
  2. Send `PUT /api/v1/chatbots/{id}` with `{"refresh_policy": "auto"}`. Expect `200 OK`.

---

### [x] 2.2.9 Hide Branding Allowed
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Update hide_branding = true | 200 OK |
| 2 | Widget does not show "Powered by Botla" | Branding hidden |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
  - Create a chatbot.
- **Steps:**
  1. Send `PUT` with `{"custom_branding": {"hide_branding": true}}`. Expect `200 OK`.
  2. Verify config is persisted.

---

### [x] 2.2.10 Custom Branding NOT Allowed
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to set custom_branding | 403 Forbidden (Ultra only) |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
- **Steps:**
  1. Send `PUT` with `{"custom_branding": {"logo_url": "..."}}`. Expect `403 Forbidden`.

---

### [x] 2.2.11 Secure Embed Enabled
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Update secure_embed_enabled = true | 200 OK |
| 2 | Set allowed_domains | 200 OK |
| 3 | Set embed_secret | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
- **Steps:**
  1. Send `PUT` with `{"secure_embed_enabled": true, "allowed_domains": ["example.com"]}`. Expect `200 OK`.

---

### [x] 2.2.12 Guardrails Full Access
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Customize threshold values | 200 OK |
| 2 | Set fallback_mode = "smart" | 200 OK |
| 3 | Set fallback_mode = "escalate" | 200 OK |
| 4 | Configure topic restrictions | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/plan_features_pro_test.go`
- **Setup:**
  - Create a user on the `pro` plan.
- **Steps:**
  1. Send `PUT` with `{"threshold_config": {"high": 0.8}, "fallback_mode": "smart", "topic_restrictions": {"allowed_topics": ["a"]}}`.
  2. Expect `200 OK`.

---

## How to Run Tests

```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/... -run "ProPlan"
```
