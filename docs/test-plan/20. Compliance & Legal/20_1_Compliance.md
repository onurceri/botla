# 20.1 Compliance Test Plan

## Overview
This test plan covers compliance and legal requirements.

---

## Test Cases

### 20.1.1 User Data Deletion
**Priority:** High  
**Type:** Compliance Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Request data deletion | All user data removed |
| 2 | Verify database | No user records |
| 3 | Verify Qdrant | Embeddings removed |

**Implementation Plan:**
- **Test File:** `internal/integration/compliance_test.go`
- **Setup:**
  - Create User, Bot, Source, Chat Logs.
- **Steps:**
  1. Call `DELETE /api/v1/user` (or equivalent admin endpoint).
  2. Verify 200/204.
  3. Query `users`, `chatbots`, `data_sources` by user ID. Expect 0 rows (or soft deleted).
  4. Verify mock Qdrant received delete collection call.

---

### 20.1.2 Data Export
**Priority:** Medium  
**Type:** Compliance Test (PLANNED)

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Request data export | JSON/CSV generated |
| 2 | Verify completeness | All user data included |

**Implementation Plan:**
- **Test File:** `internal/integration/compliance_test.go`
- **Steps:**
  1. Call `GET /api/v1/user/export`.
  2. If implemented: Verify JSON contains bots, sources, chats.
  3. If not implemented: Verify 501 Not Implemented (placeholder test).

---

### 20.1.3 Security Headers
**Priority:** High  
**Type:** Security Audit

| Header | Expected |
|--------|----------|
| X-Content-Type-Options | nosniff |
| X-Frame-Options | DENY |
| Content-Security-Policy | Configured |

**Implementation Plan:**
- **Test File:** `internal/integration/compliance_test.go`
- **Steps:**
  1. Call `GET /`.
  2. Check `X-Content-Type-Options` == `nosniff`.
  3. Check `X-Frame-Options` == `DENY` or `SAMEORIGIN`.
  4. Check `Content-Security-Policy` exists.

---

### 20.1.4 Dependency Vulnerabilities
**Priority:** High  
**Type:** Security Audit

```bash
# Check Go dependencies
go list -m -json all | nancy sleuth

# Check npm dependencies
cd frontend && npm audit
```

**Implementation Plan:**
- **Test Script:** `scripts/audit.sh`
- **Steps:**
  1. Run `govulncheck ./...` (modern replacement for nancy).
  2. Run `npm audit` in frontend.
  3. Fail build if critical vulnerabilities found.

---

## How to Run Tests

```bash
# Security audit
go list -m -json all | nancy sleuth
cd frontend && npm audit
```
