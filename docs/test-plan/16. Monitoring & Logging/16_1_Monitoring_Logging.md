# 16.1 Monitoring & Logging Test Plan

## Overview
This test plan covers logging and health check functionality.

---

## Test Cases

### 16.1.1 Health Check Endpoint
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET /health | 200 OK |
| 2 | Response | Health status |

**Implementation Plan:**
- **Test File:** `internal/integration/health_test.go`
- **Steps:**
  1. GET `/health`.
  2. Verify 200 OK.
  3. Verify JSON body `status: "ok"`.

---

### 16.1.2 Error Logging
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Trigger error | Error logged |
| 2 | Stack trace included | For debugging |

**Implementation Plan:**
- **Test File:** `internal/integration/monitoring_test.go`
- **Setup:**
  - Configure logger to write to buffer.
  - Mock endpoint that panics or returns error.
- **Steps:**
  1. Call endpoint.
  2. Verify buffer contains error message.

---

### 16.1.3 Request Logging
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Make API request | Request logged |
| 2 | Includes request ID | Traceable |

**Implementation Plan:**
- **Test File:** `internal/integration/monitoring_test.go`
- **Setup:**
  - Capture logs.
- **Steps:**
  1. Call `GET /health`.
  2. Verify logs contain "GET /health".
  3. Verify logs contain `request_id`.

---

### 16.1.4 Sensitive Data Not Logged
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Review logs | No passwords |
| 2 | | No JWT secrets |
| 3 | | No API keys |

**Implementation Plan:**
- **Test File:** `internal/integration/monitoring_test.go`
- **Setup:**
  - Capture logs.
- **Steps:**
  1. Call login with `password="SECRET123"`.
  2. Check logs. Verify "SECRET123" is NOT present.

---

## How to Run Tests

```bash
curl http://localhost:8080/health
```
