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

**Implementation Plan:**
- **Test File:** `internal/integration/security_validation_test.go`
- **Setup:**
  - Create user.
- **Steps:**
  1. POST `/api/v1/chatbots` with `{"name": "<script>alert('xss')</script>"}`.
  2. Verify 201.
  3. GET chatbot. Verify `name` is sanitized (e.g., tags removed) or escaped (`&lt;script&gt;`).

---

### 9.1.2 SQL Injection Prevention
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit `'; DROP TABLE users;--` | Safe handling |
| 2 | Database integrity | Tables intact |

**Implementation Plan:**
- **Test File:** `internal/integration/security_validation_test.go`
- **Steps:**
  1. POST `/api/v1/auth/register` with `email="'; DROP TABLE users;--"`.
  2. Verify 400 Bad Request (invalid email) or 201 (safe insertion).
  3. Verify `users` table count has not decreased unexpectedly.

---

### 9.1.3 File MIME Type Validation
**Priority:** High  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload .exe renamed to .pdf | Rejected |
| 2 | MIME type checked | Not just extension |

**Implementation Plan:**
- **Test File:** `internal/integration/security_validation_test.go`
- **Setup:**
  - Create bot.
  - Create a dummy EXE file content but name it `fake.pdf`.
- **Steps:**
  1. Upload `fake.pdf` via multipart form.
  2. Expect 400 Bad Request.

---

### 9.1.4 URL Format Validation
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Valid URL | Accepted |
| 2 | Invalid format | 400 Bad Request |

**Implementation Plan:**
- **Test File:** `internal/integration/security_validation_test.go`
- **Steps:**
  1. POST source with `url="http://valid.com"`. Expect 201.
  2. POST source with `url="javascript:alert(1)"`. Expect 400.
  3. POST source with `url="ftp://server"`. Expect 400 (if only HTTP/S allowed).

---

### 9.1.5 JSON Payload Validation
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Invalid JSON | 400 Bad Request |
| 2 | Missing required fields | 400 with field error |

**Implementation Plan:**
- **Test File:** `internal/integration/security_validation_test.go`
- **Steps:**
  1. Send POST with body `{"name":`. Expect 400 (Syntax error).
  2. Send POST `/chatbots` with `{}`. Expect 400 (Missing name).

---

### 9.1.6 Oversized Request Rejection
**Priority:** Medium  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | 100MB request body | 413 Payload Too Large |

**Implementation Plan:**
- **Test File:** `internal/integration/security_validation_test.go`
- **Steps:**
  1. Send POST with a 100MB body string.
  2. Expect 413 Payload Too Large.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Validation|Security"
```
