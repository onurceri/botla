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

---

### 4.4.2 Refresh Plan Enforcement
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free user refresh | 403 Forbidden |
| 2 | Pro user refresh | 200 OK |

---

### 4.4.3 Only URL Sources Refreshable
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Refresh PDF source | 400 Bad Request |
| 2 | Refresh text source | 400 Bad Request |

---

### 4.4.4 Cannot Refresh Processing Source
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Source status = "processing" | N/A |
| 2 | Attempt refresh | 409 Conflict |

---

### 4.4.5 Refresh Updates Embeddings
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | URL content changes | N/A |
| 2 | Refresh source | New content fetched |
| 3 | Embeddings updated | Qdrant updated |

---

### 4.4.6 Refresh Does Not Increment Ingestion Count
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Note ingestion count | Count = N |
| 2 | Refresh existing source | Success |
| 3 | Ingestion count | Still = N |

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

---

### 4.4.8 Auto-Refresh Execution
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | next_refresh_at in past | Scheduler picks up |
| 2 | Source refreshed | Content updated |
| 3 | next_refresh_at updated | Next scheduled time |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Refresh"
```
