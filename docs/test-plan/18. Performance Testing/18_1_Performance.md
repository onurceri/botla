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

---

### 18.1.2 Database Query Performance
**Priority:** High  
**Type:** Performance Test

| Query | Target |
|-------|--------|
| List chatbots | < 50ms |
| Get chatbot by ID | < 10ms |
| Search sources | < 100ms |

---

### 18.1.3 Chat Response Time
**Priority:** High  
**Type:** Performance Test

| Scenario | Target |
|----------|--------|
| Simple query | < 2s |
| RAG query | < 5s |

---

### 18.1.4 Pagination
**Priority:** Medium  
**Type:** Performance Test

| Test | Expected |
|------|----------|
| 1000+ messages | Paginated correctly |
| 100+ chatbots | Fast list response |

---

## How to Run Tests

```bash
# Install k6 or hey for load testing
k6 run load-test.js
```
