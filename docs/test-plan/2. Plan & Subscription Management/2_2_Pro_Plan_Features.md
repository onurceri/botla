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

### 2.2.1 Model Selection - gpt-4o Available
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot with model `gpt-4o-mini` | 201 Created |
| 2 | Update to model `gpt-4o` | 200 OK |
| 3 | Update to model `claude-3-5-sonnet` | 403 Forbidden (Ultra only) |

---

### 2.2.2 Monthly Token Limit - 1,000,000
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use tokens up to 999,999 | Chat succeeds |
| 2 | Exceed 1,000,000 total | 429 Too Many Requests |

---

### 2.2.3 PDF Limits - 20 per Bot, 20MB each
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload 20 PDFs (< 20MB each) | All succeed |
| 2 | Upload 21st PDF | 403 Forbidden |
| 3 | Upload PDF > 20MB | 413 Payload Too Large |

---

### 2.2.4 OCR Enabled
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload image-based PDF | Source created |
| 2 | Check extracted text | OCR text extracted from images |

---

### 2.2.5 URL Limits - 10 per Bot
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 10 URL sources | All succeed |
| 2 | Add 11th URL | 403 Forbidden |

---

### 2.2.6 Dynamic Scraping Enabled
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL with JS-rendered content | Source created |
| 2 | Check scraped content | JS-rendered content included |

---

### 2.2.7 Discovery Mode - Max 10 Pages
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set discovery_mode = "auto" | 200 OK |
| 2 | Add URL with 15 sub-pages | Only 10 pages discovered |

---

### 2.2.8 Refresh Available
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Manually refresh URL source | 200 OK |
| 2 | Set refresh_policy = "auto" | 200 OK |

---

### 2.2.9 Hide Branding Allowed
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Update hide_branding = true | 200 OK |
| 2 | Widget does not show "Powered by Botla" | Branding hidden |

---

### 2.2.10 Custom Branding NOT Allowed
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Try to set custom_branding | 403 Forbidden (Ultra only) |

---

### 2.2.11 Secure Embed Enabled
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Update secure_embed_enabled = true | 200 OK |
| 2 | Set allowed_domains | 200 OK |
| 3 | Set embed_secret | 200 OK |

---

### 2.2.12 Guardrails Full Access
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Customize threshold values | 200 OK |
| 2 | Set fallback_mode = "smart" | 200 OK |
| 3 | Set fallback_mode = "escalate" | 200 OK |
| 4 | Configure topic restrictions | 200 OK |

---

## How to Run Tests

```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/... -run "ProPlan"
```
