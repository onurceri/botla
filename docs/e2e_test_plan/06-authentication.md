# 06. Authentication & Session Tests

> **Priority**: Critical  
> **Test Count**: 18  
> **Source Files**: `internal/auth/`, `internal/api/handlers/auth.go`, `pkg/middleware/auth.go`

---

## 6.1 Registration & Login

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| AUTH-001 | Register with valid email/password | 201, tokens returned | ✅ |
| AUTH-002 | Register with existing email | 409 Conflict | ✅ |
| AUTH-003 | Register with weak password | 400 Bad Request | ✅ |
| AUTH-004 | Login with valid credentials | 200, tokens returned | ✅ |
| AUTH-005 | Login with invalid password | 401 Unauthorized | ✅ |
| AUTH-006 | Login with non-existent email | 401 Unauthorized | ✅ |
| AUTH-007 | Auto-create default org/workspace | Org and WS created on register | ✅ |

### Technical Notes

```go
// Registration creates:
// 1. User record
// 2. Default organization: "<user_name> Organizasyonu" or "Kişisel Organizasyon"
// 3. Default workspace: "Varsayılan"
// 4. Membership with role "owner"
```

---

## 6.2 Token Management

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| TKN-001 | Access token format (JWT) | Valid JWT structure | ✅ |
| TKN-002 | Access token expiration (15 min) | Expires correctly | ✅ |
| TKN-003 | Refresh token format | Valid JWT structure | ✅ |
| TKN-004 | Refresh token expiration (7 days) | Expires correctly | ✅ |
| TKN-005 | Refresh with valid token | New access token returned | ✅ |
| TKN-006 | Refresh with expired token | 401, redirect to login | ✅ |
| TKN-007 | Refresh with revoked token | 401, redirect to login | ✅ |
| TKN-008 | Refresh token rotation | Old token invalidated | ✅ |
| TKN-009 | Bearer format validation | "Bearer " prefix required | ✅ |

---

## 6.3 Authorization Header

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| HDR-001 | Missing Authorization header | 401 Unauthorized | ✅ |
| HDR-002 | Invalid Authorization format | 401 Unauthorized | ✅ |
| HDR-003 | Expired access token | 401 Unauthorized | ✅ |
| HDR-004 | Tampered token (invalid signature) | 401 Unauthorized | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/auth/jwt_test.go` | JWT generation/validation |
| `internal/auth/jwt_negative_test.go` | Invalid tokens |
| `internal/auth/password_test.go` | Password hashing |
| `internal/integration/auth_test.go` | Login/register flow |
| `internal/integration/auth_refresh_rotation_test.go` | Token rotation |
| `internal/integration/auth_revoked_refresh_test.go` | Revoked tokens |
| `internal/integration/auth_bearer_format_test.go` | Bearer format |
| `internal/integration/auth_expired_access_test.go` | Expired access token |
| `internal/integration/auth_refresh_expired_test.go` | Expired refresh token |
| `internal/integration/auth_registration_side_effects_test.go` | Default org/workspace creation |
| `internal/integration/auth_weak_password_test.go` | Weak password registration |
