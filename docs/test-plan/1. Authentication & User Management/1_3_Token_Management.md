# 1.3 Token Management Test Plan

## Overview
This test plan covers JWT token validation, refresh token rotation, and logout functionality.

---

## Test Cases

### 1.3.1 Access Token Validation [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login to get access token | Token received [x] |
| 2 | GET `/api/v1/me` with `Authorization: Bearer <token>` | 200 OK with user info [x] |
| 3 | Verify response contains user data | user_id, email, plan details [x] |

**Existing Test:** `internal/integration/auth_edges_test.go` - `TestAuth_ValidAccessToken_Me200`

---

### 1.3.2 Expired Access Token [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use an expired access token | N/A [x] |
| 2 | GET `/api/v1/me` with expired token | 401 Unauthorized [x] |
| 3 | Response body | Token expired message [x] |

**Existing Test:** `internal/integration/auth_expired_access_test.go`

---

### 1.3.3 Invalid Access Token [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/me` with `Authorization: Bearer invalid-token` | 401 Unauthorized [x] |
| 2 | GET `/api/v1/me` with tampered JWT (modified payload) | 401 Unauthorized [x] |

**Existing Test:** `internal/integration/auth_edges_test.go`

---

### 1.3.4 Missing Authorization Header [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/me` without Authorization header | 401 Unauthorized [x] |

---

### 1.3.5 Malformed Bearer Token Format [x]
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/me` with `Authorization: token` (no Bearer) | 401 Unauthorized [x] |
| 2 | GET `/api/v1/me` with `Authorization: Bearer` (no token) | 401 Unauthorized [x] |
| 3 | GET `/api/v1/me` with `Authorization: Basic abc123` | 401 Unauthorized [x] |

**Existing Test:** `internal/integration/auth_bearer_format_test.go`

---

### 1.3.6 Refresh Token - Generate New Access Token [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login to get refresh token | Tokens received [x] |
| 2 | POST `/api/v1/auth/refresh` with refresh token | 200 OK [x] |
| 3 | Response contains new access token | Valid JWT [x] |
| 4 | Old access token still works (until expiry) | Depends on implementation [x] |

**Existing Test:** `internal/integration/auth_refresh_success_test.go` - `TestAuth_Refresh_GeneratesNewAccessToken`

---

### 1.3.7 Expired Refresh Token [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use an expired refresh token | N/A [x] |
| 2 | POST `/api/v1/auth/refresh` | 401 Unauthorized [x] |

**Existing Test:** `internal/integration/auth_refresh_expired_test.go`

---

### 1.3.8 Revoked Refresh Token [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login to get refresh token | Token received [x] |
| 2 | Logout (revokes refresh token) | 200 OK [x] |
| 3 | POST `/api/v1/auth/refresh` with revoked token | 401 Unauthorized [x] |

**Existing Test:** `internal/integration/auth_revoked_refresh_test.go`

---

### 1.3.9 Refresh Token Rotation [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login to get refresh token A | Token A received [x] |
| 2 | POST `/api/v1/auth/refresh` with token A | New tokens received [x] |
| 3 | Response includes new refresh token B | Token B different from A [x] |
| 4 | Old refresh token A is invalidated | Cannot use A again [x] |

**Existing Test:** `internal/integration/auth_refresh_rotation_test.go`

---

### 1.3.10 Logout Invalidates Refresh Token [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login to get tokens | Tokens received [x] |
| 2 | POST `/api/v1/auth/logout` with access token | 200 OK [x] |
| 3 | Try to use refresh token | 401 Unauthorized [x] |

---

### 1.3.11 JWT Claims Verification [x]
**Priority:** High  
**Type:** Unit Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Decode access token | Contains `user_id` claim [x] |
| 2 | Verify `exp` claim | Present and future timestamp [x] |
| 3 | Verify `iat` claim | Present and past timestamp [x] |

---

## How to Run Tests

### Run All Token Management Tests
```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/auth*.go
```

### Run Specific Token Tests
```bash
go test -v ./internal/integration/... -run TestExpiredAccess
go test -v ./internal/integration/... -run TestRefreshExpired
go test -v ./internal/integration/... -run TestRevokedRefresh
go test -v ./internal/integration/... -run TestRefreshRotation
go test -v ./internal/integration/... -run TestBearerFormat
```

### Run Middleware Unit Tests
```bash
go test -v ./pkg/middleware/auth_test.go
go test -v ./pkg/middleware/auth_format_test.go
```

---

## Coverage Notes
- All critical token scenarios have existing tests
- Consider adding test for JWT secret rotation scenario
- Rate limiting on refresh endpoint not yet tested
