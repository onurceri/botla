# 4.3 Source Processing Test Plan

## Overview
This test plan covers the background processing of sources including status transitions and error handling.

---

## Test Cases

### 4.3.1 Status Transitions
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add source | status = "pending" |
| 2 | Processing starts | status = "processing" |
| 3 | Processing completes | status = "completed" |

---

### 4.3.2 Failed Processing
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add inaccessible URL | Source created |
| 2 | Processing fails | status = "failed" |
| 3 | Error message stored | error_message field set |

---

### 4.3.3 URL Content Fetching
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add valid URL | Source created |
| 2 | Processing completes | Content extracted |
| 3 | chunk_count > 0 | Chunks created |

---

### 4.3.4 Handle 404 Errors
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL returning 404 | Source created |
| 2 | Processing | status = "failed" |
| 3 | Error message | "Page not found" or similar |

---

### 4.3.5 Handle Timeout
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add slow/timing out URL | Source created |
| 2 | Processing | Timeout handled |
| 3 | Status | "failed" with timeout error |

---

### 4.3.6 PDF Text Extraction
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload text-based PDF | Source created |
| 2 | Processing | Text extracted |
| 3 | Chunks created | chunk_count > 0 |

---

### 4.3.7 Embeddings Generated
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Process source | Completed |
| 2 | Query Qdrant | Embeddings exist for chatbot |

---

### 4.3.8 Embedding Token Tracking
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Process source | Completed |
| 2 | Check usage | Embedding tokens incremented |

---

### 4.3.9 Embedding Token Limit
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use 250,000 embedding tokens | All succeed |
| 2 | Add source requiring more | 403 Forbidden |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "SourceProcessing|Processing"
```
