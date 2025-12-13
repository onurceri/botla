# 18.1 Performance Test Plan

## Overview
This test plan covers load testing and performance benchmarks.

---

## Benchmarks

| Metric | Target |
|--------|--------|
| Concurrent users | 100+ |
| Requests/minute | 1000+ |
| DB query time | < 100ms |
| API response time | < 500ms |
| Chat response time | < 5s |
| Vector search time | < 200ms |

---

## Test Cases

### 18.1.1 Load Test - API
**Priority:** High  
**Type:** Performance Test

```bash
# Using hey or k6
hey -n 1000 -c 100 http://localhost:8080/api/v1/health
```

**Implementation Plan:**
- **Test Script:** `scripts/load_test_api.js` (k6)
- **Steps:**
  1. `k6 run scripts/load_test_api.js`
  2. Verify `http_req_duration` p95 < 500ms.
  3. Verify `http_req_failed` rate is 0%.

---

### 18.1.2 Database Query Performance
**Priority:** High  
**Type:** Performance Test

| Query | Target |
|-------|--------|
| List chatbots | < 50ms |
| Get chatbot by ID | < 10ms |
| Search sources | < 100ms |

**Implementation Plan:**
- **Test File:** `internal/db/performance_test.go` (Go Benchmarks)
- **Steps:**
  1. `go test -bench=. ./internal/db/...`
  2. Review `BenchmarkListChatbots`, `BenchmarkGetChatbot` output.
  3. Ensure `ns/op` converts to < target ms.

---

### 18.1.3 Chat Response Time
**Priority:** High  
**Type:** Performance Test

| Scenario | Target |
|----------|--------|
| Simple query | < 2s |
| RAG query | < 5s |

**Implementation Plan:**
- **Test Script:** `scripts/load_test_chat.js` (k6)
- **Steps:**
  1. Mock LLM/Qdrant to have fixed latency (e.g. 1s).
  2. Run k6 script simulating chat users.
  3. Verify system overhead is minimal (< 500ms over mock latency).

---

### 18.1.4 Pagination
**Priority:** Medium  
**Type:** Performance Test

| Test | Expected |
|------|----------|
| 1000+ messages | Paginated correctly |
| 100+ chatbots | Fast list response |

**Implementation Plan:**
- **Test File:** `internal/integration/db_pagination_test.go`
- **Setup:**
  - Insert 1000 dummy records.
- **Steps:**
  1. Measure time for `GET /resource?limit=10&offset=900`.
  2. Verify time is comparable to `offset=0`.

---

## How to Run Tests

```bash
# Install k6 or hey for load testing
k6 run load-test.js
```
