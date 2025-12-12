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

---

### 9.2.2 Invalid JWT Handling
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Malformed JWT | 401 Unauthorized |
| 2 | Tampered JWT | 401 Unauthorized |
| 3 | Expired JWT | 401 Unauthorized |

---

### 9.2.3 Cross-User Access Prevention
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | User A accesses User B's chatbot | 403 Forbidden |
| 2 | User A accesses User B's source | 403 Forbidden |
| 3 | User A accesses User B's conversation | 403 Forbidden |

---

### 9.2.4 Organization Role Enforcement
**Priority:** High  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Owner has full access | All operations |
| 2 | Admin limited access | Per RBAC rules |
| 3 | Member limited access | Read + create only |

---

### 9.2.5 Non-Member Org Access
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Non-member accesses org | 403 Forbidden |
| 2 | Non-member accesses workspace | 403 Forbidden |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Auth|RBAC|Access"
```
