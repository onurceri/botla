# 1.1 User Registration Test Plan

## Overview
This test plan covers all user registration functionality including validation, security, and side effects.

---

## Test Cases

### 1.1.1 Valid Registration Flow [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/register` with valid email, password (8+ chars), and full_name | 201 Created |
| 2 | Response contains `token` and `refresh_token` | Both tokens present and valid JWT format |
| 3 | Query database for user | User exists with hashed password (bcrypt) |
| 4 | Verify user plan assignment | User has `free` plan assigned |

**Existing Test:** `internal/integration/auth_test.go` - `TestRegister`

---

### 1.1.2 Invalid Email Format [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/register` with email `invalid-email` | 400 Bad Request |
| 2 | Response body | Contains validation error message |

**Existing Test:** `internal/integration/auth_test.go` - covers validation

---

### 1.1.3 Duplicate Email Registration [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Register user with email `test@example.com` | 201 Created |
| 2 | Register another user with same email | 409 Conflict |

**Existing Test:** `internal/integration/auth_test.go`

---

### 1.1.4 Weak Password Validation [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/register` with password `short` (< 8 chars) | 400 Bad Request |
| 2 | Response body | Contains password requirements error |

**Existing Test:** `internal/integration/auth_weak_password_test.go`

---

### 1.1.5 Missing Required Fields [x]
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/register` with missing email | 400 Bad Request |
| 2 | POST `/api/v1/auth/register` with missing password | 400 Bad Request |

**Existing Test:** `internal/integration/auth_test.go`

---

### 1.1.6 SQL Injection Prevention [x]
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/register` with email `'; DROP TABLE users;--` | 400 Bad Request (validation) or safe handling |
| 2 | Verify database integrity | No tables dropped, user not created |

**Test Command:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "'"'"'; DROP TABLE users;--", "password": "testpassword123"}'
```

**Existing Test:** `internal/integration/auth_security_test.go` - `TestAuth_Register_SQLInjectionEmail`

---

### 1.1.7 XSS Prevention in Registration Fields [x]
**Priority:** High  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/register` with full_name `<script>alert('xss')</script>` | User created with sanitized/escaped name |
| 2 | Retrieve user profile | No raw script tags returned |
**Existing Test:** `internal/integration/auth_security_test.go` - `TestAuth_Register_XSSFullName`
---

### 1.1.8 Password Hashing Verification [x]
**Priority:** Critical  
**Type:** Unit/Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Register user with password `mypassword123` | 201 Created |
| 2 | Query database directly | `password_hash` is bcrypt hash, not plaintext |
| 3 | Verify hash format | Starts with `$2a$` or `$2b$` (bcrypt) |

**Existing Test:** `internal/integration/auth_password_hash_test.go` - `TestAuth_Register_PasswordHashing`

---

### 1.1.9 Registration Side Effects [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Register new user | User created |
| 2 | Check organization creation | Default personal organization created |
| 3 | Check workspace creation | Default workspace created in organization |
| 4 | Check membership | User is owner of organization |

**Existing Test:** `internal/integration/auth_registration_side_effects_test.go`

---

## How to Run Tests

### Run All Auth Tests
```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/auth*.go -run "Register"
```

### Run Specific Test
```bash
go test -v ./internal/integration/... -run TestRegister
go test -v ./internal/integration/... -run TestWeakPassword
go test -v ./internal/integration/... -run TestRegistrationSideEffects
```

---

## Coverage Notes
- Existing tests cover most happy path and validation scenarios
- Security tests (SQL injection, XSS) now have automated coverage plus optional manual verification
- Password hashing is implicitly tested by login tests
