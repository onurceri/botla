# 9.2 Authentication & Authorization Test Plan

## Overview
This test plan covers auth security across all protected endpoints.

---

## Test Cases

### 9.2.1 Protected Endpoints Require JWT
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET /api/v1/chatbots without token | 401 Unauthorized |
| 2 | GET /api/v1/me without token | 401 Unauthorized |

**Implementation Plan:**
- **Test File:** `internal/integration/auth_security_test.go`
- **Steps:**
  1. GET `/api/v1/me` with no Authorization header.
  2. Verify 401 Unauthorized.

---

### 9.2.2 Invalid JWT Handling
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Malformed JWT | 401 Unauthorized |
| 2 | Tampered JWT | 401 Unauthorized |
| 3 | Expired JWT | 401 Unauthorized |

**Implementation Plan:**
- **Test File:** `internal/integration/auth_security_test.go`
- **Steps:**
  1. GET `/me` with `Bearer invalid.token`. Expect 401.
  2. GET `/me` with a valid token signed by a different secret. Expect 401.
  3. GET `/me` with an expired token. Expect 401.

---

### 9.2.3 Cross-User Access Prevention
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | User A accesses User B's chatbot | 403 Forbidden |
| 2 | User A accesses User B's source | 403 Forbidden |
| 3 | User A accesses User B's conversation | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/auth_security_test.go`
- **Setup:**
  - User A, User B.
  - Bot B owned by B.
- **Steps:**
  1. Login as A.
  2. GET `/api/v1/chatbots/{bot_b_id}`.
  3. Verify 403 Forbidden.

---

### 9.2.4 Organization Role Enforcement
**Priority:** High  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Owner has full access | All operations |
| 2 | Admin limited access | Per RBAC rules |
| 3 | Member limited access | Read + create only |

**Implementation Plan:**
- **Test File:** `internal/integration/auth_security_test.go`
- **Setup:**
  - Org with Owner, Admin, Member.
- **Steps:**
  1. Member attempts to update Org settings. Expect 403.
  2. Admin attempts to delete Org. Expect 403.
  3. Owner attempts to delete Org. Expect 200.

---

### 9.2.5 Non-Member Org Access
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Non-member accesses org | 403 Forbidden |
| 2 | Non-member accesses workspace | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/auth_security_test.go`
- **Setup:**
  - User A (in Org A). User B (not in Org A).
- **Steps:**
  1. Login as B.
  2. GET `/api/v1/organizations/{org_a_id}`.
  3. Verify 403 Forbidden.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Auth|RBAC|Access"
```
