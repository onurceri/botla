# 10.1 Migration Integrity Test Plan

## Overview
This test plan covers database migration validity and rollback capabilities.

---

## Test Cases

### 10.1.1 All Migrations Run Successfully
**Priority:** Critical  
**Type:** Integration Test

```bash
make migrate-up
# Expected: All migrations applied without error
```

---

### 10.1.2 Rollback Works
**Priority:** High  
**Type:** Integration Test

```bash
make migrate-down
# Expected: Migrations rolled back without error
```

---

### 10.1.3 Migration Idempotency
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Run migrations | Success |
| 2 | Run migrations again | No changes, no error |

---

### 10.1.4 Foreign Key Constraints
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Delete user | Cascade to chatbots |
| 2 | Delete chatbot | Cascade to sources |

---

### 10.1.5 Index Creation
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Check indexes exist | All expected indexes present |
| 2 | Query performance | Fast on indexed columns |

---

## How to Run Tests

```bash
cd /Users/onur/Documents/workspace/botla-co
make migrate-up
make migrate-down
```
