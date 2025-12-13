# 4.4 Source Refresh Test Plan

## Overview
This test plan covers manual and automatic source refresh functionality.

---

## Test Cases

### 4.4.1 Manual Refresh URL Source
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/sources/{id}/refresh` | 200 OK |
| 2 | Source re-fetched | Content updated |
| 3 | last_refreshed_at updated | New timestamp |

**Implementation Plan:**
- **Test File:** `internal/integration/source_refresh_test.go`
- **Setup:**
  - Create bot and URL source.
- **Steps:**
  1. Sleep 1s.
  2. Call POST refresh. Expect 200.
  3. Fetch source. Verify `last_refreshed_at` is recent.

---

### 4.4.2 Refresh Plan Enforcement
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free user refresh | 403 Forbidden |
| 2 | Pro user refresh | 200 OK |

**Implementation Plan:**
- **Test File:** `internal/integration/source_refresh_test.go`
- **Setup:**
  - Free User, Pro User. Both with URL sources.
- **Steps:**
  1. Free refresh -> 403.
  2. Pro refresh -> 200.

---

### 4.4.3 Only URL Sources Refreshable
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Refresh PDF source | 400 Bad Request |
| 2 | Refresh text source | 400 Bad Request |

**Implementation Plan:**
- **Test File:** `internal/integration/source_refresh_test.go`
- **Setup:**
  - Create bot with PDF source.
- **Steps:**
  1. Refresh PDF source. Expect 400.

---

### 4.4.4 Cannot Refresh Processing Source
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Source status = "processing" | N/A |
| 2 | Attempt refresh | 409 Conflict |

**Implementation Plan:**
- **Test File:** `internal/integration/source_refresh_test.go`
- **Setup:**
  - Create source, manually set status to "processing".
- **Steps:**
  1. Refresh. Expect 409.

---

### 4.4.5 Refresh Updates Embeddings
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | URL content changes | N/A |
| 2 | Refresh source | New content fetched |
| 3 | Embeddings updated | Qdrant updated |

**Implementation Plan:**
- **Test File:** `internal/integration/source_refresh_test.go`
- **Setup:**
  - Create source. Mock HTML content change.
- **Steps:**
  1. Refresh.
  2. Verify mock Qdrant received new vectors (or check `hash` changed in DB).

---

### 4.4.6 Refresh Does Not Increment Ingestion Count
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Note ingestion count | Count = N |
| 2 | Refresh existing source | Success |
| 3 | Ingestion count | Still = N |

**Implementation Plan:**
- **Test File:** `internal/integration/source_refresh_test.go`
- **Setup:**
  - User.
- **Steps:**
  1. Check ingestion count.
  2. Refresh source.
  3. Check ingestion count again. Verify unchanged.

---

## Auto-Refresh Test Cases

### 4.4.7 Auto-Refresh Scheduling
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set refresh_policy = "auto" | 200 OK |
| 2 | Set refresh_frequency = "daily" | 200 OK |
| 3 | next_refresh_at set | Tomorrow's date |

**Implementation Plan:**
- **Test File:** `internal/integration/source_refresh_test.go`
- **Setup:**
  - Pro Bot.
- **Steps:**
  1. Update `refresh_policy="auto"`, `refresh_frequency="daily"`.
  2. Fetch bot.
  3. Verify `next_refresh_at` is approx 24h from now.

---

### 4.4.8 Auto-Refresh Execution
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | next_refresh_at in past | Scheduler picks up |
| 2 | Source refreshed | Content updated |
| 3 | next_refresh_at updated | Next scheduled time |

**Implementation Plan:**
- **Test File:** `internal/integration/source_refresh_test.go`
- **Setup:**
  - Source with `next_refresh_at` in the past.
- **Steps:**
  1. Run the refresh worker (or wait for it if running in test env).
  2. Verify source status changes to `processing`/`completed`.
  3. Verify `next_refresh_at` is updated to future.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Refresh"
```
