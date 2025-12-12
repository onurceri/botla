# 4.5 Discovery Mode Test Plan

## Overview
This test plan covers automatic sub-page discovery during URL crawling.

---

## Test Cases

### 4.5.1 Discovery Mode Disabled
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | discovery_mode = "disabled" | Default |
| 2 | Add URL with links | Only main page scraped |
| 3 | No pending_urls created | Table empty |

---

### 4.5.2 Discovery Mode Auto
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pro user: discovery_mode = "auto" | 200 OK |
| 2 | Add URL with sub-links | Sub-pages auto-added |
| 3 | Sources created | Up to max_pages_per_crawl |

---

### 4.5.3 Discovery Mode Pending
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | discovery_mode = "pending" | 200 OK |
| 2 | Add URL with sub-links | Links added to pending_urls |
| 3 | Pending URLs in UI | Displayed for approval |

---

### 4.5.4 Approve Pending URL
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pending URL exists | In pending_urls table |
| 2 | POST approve | 201 Created |
| 3 | New source created | Source added to chatbot |

---

### 4.5.5 Reject Pending URL
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pending URL exists | In pending_urls table |
| 2 | POST reject | 200 OK |
| 3 | URL removed | No source created |

---

### 4.5.6 Max Pages Per Crawl Limit
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pro: max_pages_per_crawl = 10 | Config limit |
| 2 | Add URL with 15 sub-pages | Only 10 discovered |
| 3 | Ultra: 100 sub-pages | Only 100 discovered |

---

### 4.5.7 Same Domain Restriction
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add example.com URL | Discovery starts |
| 2 | Links to other-domain.com | Ignored |
| 3 | Only example.com links | Discovered |

---

### 4.5.8 Include/Exclude Paths
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set include_paths = ["/docs/*"] | Only docs crawled |
| 2 | Set exclude_paths = ["/admin/*"] | Admin pages skipped |

---

### 4.5.9 Free Plan Discovery Blocked
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: discovery_mode = "auto" | 403 Forbidden |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Discovery"
```
