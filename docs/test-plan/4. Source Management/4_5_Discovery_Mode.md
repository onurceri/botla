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

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Create bot.
- **Steps:**
  1. Add source `http://example.com` (mock returns links).
  2. Wait for completion.
  3. Verify `SELECT count(*) FROM data_sources` is 1.
  4. Verify `SELECT count(*) FROM pending_urls` is 0.

---

### 4.5.2 Discovery Mode Auto
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pro user: discovery_mode = "auto" | 200 OK |
| 2 | Add URL with sub-links | Sub-pages auto-added |
| 3 | Sources created | Up to max_pages_per_crawl |

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Pro user. Bot with `discovery_mode="auto"`.
- **Steps:**
  1. Add source (mock returns 15 links).
  2. Wait for completion.
  3. Verify multiple sources created (count <= limit).

---

### 4.5.3 Discovery Mode Pending
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | discovery_mode = "pending" | 200 OK |
| 2 | Add URL with sub-links | Links added to pending_urls |
| 3 | Pending URLs in UI | Displayed for approval |

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Bot with `discovery_mode="pending"`.
- **Steps:**
  1. Add source.
  2. Wait for completion.
  3. Verify `pending_urls` table has entries.
  4. Verify `data_sources` table has only 1 entry.

---

### 4.5.4 Approve Pending URL
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pending URL exists | In pending_urls table |
| 2 | POST approve | 201 Created |
| 3 | New source created | Source added to chatbot |

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Create a pending URL entry manually.
- **Steps:**
  1. Send `POST /api/v1/chatbots/{id}/pending_urls/{id}/approve`.
  2. Expect `201 Created`.
  3. Verify `data_sources` count increased.
  4. Verify pending URL removed (or marked approved).

---

### 4.5.5 Reject Pending URL
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pending URL exists | In pending_urls table |
| 2 | POST reject | 200 OK |
| 3 | URL removed | No source created |

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Create a pending URL entry.
- **Steps:**
  1. Send `POST /api/v1/chatbots/{id}/pending_urls/{id}/reject` (or `DELETE`).
  2. Expect `200 OK` (or 204).
  3. Verify no new source.
  4. Verify pending URL removed.

---

### 4.5.6 Max Pages Per Crawl Limit
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Pro: max_pages_per_crawl = 10 | Config limit |
| 2 | Add URL with 15 sub-pages | Only 10 discovered |
| 3 | Ultra: 100 sub-pages | Only 100 discovered |

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Pro User (limit 10).
- **Steps:**
  1. Add source with 15 links.
  2. Wait for completion.
  3. Verify total sources <= 11 (1 + 10).

---

### 4.5.7 Same Domain Restriction
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add example.com URL | Discovery starts |
| 2 | Links to other-domain.com | Ignored |
| 3 | Only example.com links | Discovered |

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Bot with auto discovery.
- **Steps:**
  1. Add source `http://example.com` (mock returns links to `http://other.com`).
  2. Verify `http://other.com` is NOT added as a source.

---

### 4.5.8 Include/Exclude Paths
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set include_paths = ["/docs/*"] | Only docs crawled |
| 2 | Set exclude_paths = ["/admin/*"] | Admin pages skipped |

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Bot with `scraping_config` set.
- **Steps:**
  1. Add source `http://example.com`.
  2. Mock returns `http://example.com/docs/1` and `http://example.com/admin/1`.
  3. Verify only `/docs/1` is added.

---

### 4.5.9 Free Plan Discovery Blocked
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free: discovery_mode = "auto" | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/url_discovery_test.go`
- **Setup:**
  - Free User.
- **Steps:**
  1. Update `discovery_mode="auto"`.
  2. Verify 403.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Discovery"
```
