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

---

### 20.1.2 Data Export
**Priority:** Medium  
**Type:** Compliance Test (PLANNED)

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Request data export | JSON/CSV generated |
| 2 | Verify completeness | All user data included |

---

### 20.1.3 Security Headers
**Priority:** High  
**Type:** Security Audit

| Header | Expected |
|--------|----------|
| X-Content-Type-Options | nosniff |
| X-Frame-Options | DENY |
| Content-Security-Policy | Configured |

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

---

## How to Run Tests

```bash
# Security audit
go list -m -json all | nancy sleuth
cd frontend && npm audit
```
