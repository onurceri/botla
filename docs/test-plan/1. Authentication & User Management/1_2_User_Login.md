# 1.2 User Login Test Plan

## Overview
This test plan covers all user login functionality including authentication, token generation, and error handling.

---

## Test Cases

### 1.2.1 Valid Login Flow
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/login` with valid email and password | 200 OK |
| 2 | Response contains `token` and `refresh_token` | Both tokens present and valid JWT format |
| 3 | Decode access token | Contains correct `user_id` claim |
| 4 | Verify token expiry | Access token expires in configured time (e.g., 15 min) |

**Existing Test:** `internal/integration/auth_test.go` - `TestLogin`

---

### 1.2.2 Invalid Email Login
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/login` with non-existent email | 401 Unauthorized |
| 2 | Response body | Generic "invalid credentials" message (no email enumeration) |

**Existing Test:** `internal/integration/auth_test.go`

---

### 1.2.3 Invalid Password Login
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/login` with valid email, wrong password | 401 Unauthorized |
| 2 | Response body | Generic "invalid credentials" message |

**Existing Test:** `internal/integration/auth_test.go`

---

### 1.2.4 Case-Insensitive Email Matching
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Register user with email `Test@Example.com` | 201 Created |
| 2 | Login with email `test@example.com` (lowercase) | 200 OK |
| 3 | Login with email `TEST@EXAMPLE.COM` (uppercase) | 200 OK |

---

### 1.2.5 Refresh Token Tracking
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login successfully | Tokens returned |
| 2 | Query `refresh_tokens` table | New record with token hash exists |
| 3 | Record contains user_id and expiry | Correct values |

---

### 1.2.6 Multiple Login Sessions
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login from "device A" | Get tokens A |
| 2 | Login from "device B" | Get tokens B (different) |
| 3 | Both refresh tokens valid | Both can be used independently |
| 4 | Logout from device A | Only token A invalidated |

---

### 1.2.7 Malformed JSON Request
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/login` with invalid JSON | 400 Bad Request |
| 2 | Response body | JSON parse error message |

**Existing Test:** `internal/api/handlers/auth_badjson_test.go`

---

### 1.2.8 Empty Credentials
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/auth/login` with empty email | 400 Bad Request |
| 2 | POST `/api/v1/auth/login` with empty password | 400 Bad Request |

---

## How to Run Tests

### Run All Login Tests
```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/auth_test.go -run "Login"
```

### Run Bad JSON Test
```bash
go test -v ./internal/api/handlers/auth_badjson_test.go
```

---

## Manual Testing

### Test Login via cURL
```bash
# Register a user first
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpassword123", "full_name": "Test User"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpassword123"}'
```

---

## Coverage Notes
- Rate limiting tests not yet implemented (marked in checklist)
- Consider adding tests for login attempt logging/auditing
