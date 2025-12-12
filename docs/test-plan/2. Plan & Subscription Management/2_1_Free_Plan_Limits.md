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

### 2.1.3 Monthly Token Limit - 100,000
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use tokens up to 99,999 | Chat succeeds |
| 2 | Use chat that exceeds 100,000 total | 429 Too Many Requests |
| 3 | Response includes upgrade message | Contains plan upgrade info |

---

### 2.1.4 RAG Configuration Limits
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Verify plan config | top_k = 3, max_context_tokens = 2000 |
| 2 | Chat uses correct Top-K | Only 3 chunks retrieved |

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
| 2 | Upload additional file | 403 Forbidden (storage exceeded) |

---

### 2.1.7 OCR Disabled
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload image-based PDF | Source created |
| 2 | Check extracted text | No OCR text extracted (images skipped) |

---

### 2.1.8 URL Limits
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 1 URL source | 201 Created |
| 2 | Add 2nd URL source | 403 Forbidden (max_urls_per_bot = 1) |

---

### 2.1.9 Dynamic Scraping Disabled
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL requiring JavaScript rendering | Source created |
| 2 | Check scraped content | Static HTML only (no JS-rendered content) |

---

### 2.1.10 Discovery Mode Disabled
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to set discovery_mode = "auto" | 403 Forbidden |
| 2 | Verify max_pages_per_crawl = 0 | No sub-pages discovered |

---

### 2.1.11 Refresh Disabled
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to refresh URL source | 403 Forbidden |
| 2 | Try to set refresh_policy = "auto" | 403 Forbidden |

---

### 2.1.12 Branding Cannot Be Hidden
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to update hide_branding = true | 403 Forbidden |
| 2 | Widget displays "Powered by Botla" | Branding visible |

---

### 2.1.13 Secure Embed Disabled
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to update secure_embed_enabled = true | 403 Forbidden |
| 2 | Try to set allowed_domains | 403 Forbidden |

---

### 2.1.14 Ingestion Limits
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 50 sources in a month | All succeed |
| 2 | Add 51st source | 403 Forbidden (max_monthly_ingestions) |

---

### 2.1.15 Re-add Cooldown - 60 Minutes
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL source | 201 Created |
| 2 | Delete source | 200 OK |
| 3 | Re-add same URL immediately | 403 Forbidden (cooldown) |
| 4 | Wait 60 minutes, re-add | 201 Created |

---

### 2.1.16 Guardrails Restricted
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to customize thresholds | 403 Forbidden |
| 2 | Try to use fallback_mode = "smart" | 403 Forbidden |
| 3 | Try to use fallback_mode = "escalate" | 403 Forbidden |
| 4 | Try to manage topic restrictions | 403 Forbidden |

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
