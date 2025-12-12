# 9.1 Input Validation Test Plan

## Overview
This test plan covers input validation and sanitization across all endpoints.

---

## Test Cases

### 9.1.1 XSS Prevention
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit `<script>alert('xss')</script>` in name field | Sanitized or escaped |
| 2 | Retrieved data | No raw script tags |

---

### 9.1.2 SQL Injection Prevention
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit `'; DROP TABLE users;--` | Safe handling |
| 2 | Database integrity | Tables intact |

---

### 9.1.3 File MIME Type Validation
**Priority:** High  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload .exe renamed to .pdf | Rejected |
| 2 | MIME type checked | Not just extension |

---

### 9.1.4 URL Format Validation
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Valid URL | Accepted |
| 2 | Invalid format | 400 Bad Request |

---

### 9.1.5 JSON Payload Validation
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Invalid JSON | 400 Bad Request |
| 2 | Missing required fields | 400 with field error |

---

### 9.1.6 Oversized Request Rejection
**Priority:** Medium  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | 100MB request body | 413 Payload Too Large |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Validation|Security"
```
