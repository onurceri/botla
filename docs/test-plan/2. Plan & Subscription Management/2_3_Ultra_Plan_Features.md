# 2.3 Ultra Plan Features Test Plan

## Overview
This test plan verifies all Ultra plan features including Claude model access and custom branding.

---

## Ultra Plan Specifications

| Feature | Limit |
|---------|-------|
| Allowed Models | gpt-4o-mini, gpt-4o, claude-3-5-sonnet |
| Monthly Tokens | 5,000,000 |
| RAG Top-K | 10 |
| RAG Context | 8,000 tokens |
| Files per Bot | 100 |
| File Size | 50MB |
| Total Storage | 2,000MB |
| OCR | Enabled |
| URLs per Bot | 50 |
| Dynamic Scraping | Enabled |
| Discovery Mode | Max 100 pages |
| Refresh | Manual & Auto |
| Hide Branding | Yes |
| Custom Branding | Yes |
| Secure Embed | Yes |

---

## Test Cases

### 2.3.1 Claude Model Available
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot with model `claude-3-5-sonnet` | 201 Created |
| 2 | Chat with Claude model | Response from Claude |

---

### 2.3.2 Monthly Token Limit - 5,000,000
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use tokens up to 4,999,999 | Chat succeeds |
| 2 | Exceed 5,000,000 total | 429 Too Many Requests |

---

### 2.3.3 PDF Limits - 100 per Bot, 50MB each
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload 100 PDFs (< 50MB each) | All succeed |
| 2 | Upload 101st PDF | 403 Forbidden |
| 3 | Upload PDF > 50MB | 413 Payload Too Large |

---

### 2.3.4 Storage Limit - 2GB Total
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload files totaling 2GB | Success |
| 2 | Upload additional file | 403 Forbidden |

---

### 2.3.5 URL Limits - 50 per Bot
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add 50 URL sources | All succeed |
| 2 | Add 51st URL | 403 Forbidden |

---

### 2.3.6 Discovery Mode - Max 100 Pages
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set discovery_mode = "auto" | 200 OK |
| 2 | Add URL with 150 sub-pages | Only 100 pages discovered |

---

### 2.3.7 Custom Branding Allowed
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set custom_branding.logo_url | 200 OK |
| 2 | Set custom_branding.text | 200 OK |
| 3 | Set custom_branding.link | 200 OK |
| 4 | Widget shows custom branding | Custom logo/text displayed |

---

### 2.3.8 RAG Enhanced Configuration
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Verify plan config | top_k = 10, max_context_tokens = 8000 |
| 2 | Chat retrieves up to 10 sources | Correct number of chunks |

---

## How to Run Tests

```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/... -run "UltraPlan"
```
