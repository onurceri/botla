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

**Implementation Plan:**
- **Test Script:** `scripts/test_migrations.sh`
- **Steps:**
  1. Drop test database.
  2. Create test database.
  3. Run `make migrate-up`.
  4. Verify exit code is 0.

---

### 10.1.2 Rollback Works
**Priority:** High  
**Type:** Integration Test

```bash
make migrate-down
# Expected: Migrations rolled back without error
```

**Implementation Plan:**
- **Test Script:** `scripts/test_migrations.sh`
- **Steps:**
  1. After migrate-up, run `make migrate-down`.
  2. Verify exit code is 0.
  3. Verify database is empty (no tables).

---

### 10.1.3 Migration Idempotency
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Run migrations | Success |
| 2 | Run migrations again | No changes, no error |

**Implementation Plan:**
- **Test Script:** `scripts/test_migrations.sh`
- **Steps:**
  1. Run `make migrate-up`.
  2. Run `make migrate-up` again.
  3. Verify exit code 0 and output indicates "no change".

---

### 10.1.4 Foreign Key Constraints
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Delete user | Cascade to chatbots |
| 2 | Delete chatbot | Cascade to sources |

**Implementation Plan:**
- **Test File:** `internal/integration/db_constraints_test.go`
- **Setup:**
  - Create User -> Chatbot -> Source.
- **Steps:**
  1. Delete User.
  2. Verify Chatbot is gone.
  3. Verify Source is gone.

---

### 10.1.5 Index Creation
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Check indexes exist | All expected indexes present |
| 2 | Query performance | Fast on indexed columns |

**Implementation Plan:**
- **Test File:** `internal/integration/db_schema_test.go`
- **Steps:**
  1. Query `pg_indexes` to verify presence of critical indexes (e.g., `idx_chatbots_user_id`).
  2. (Optional) Run `EXPLAIN ANALYZE` on a sample query to verify index usage.

---

## How to Run Tests

```bash
cd /Users/onur/Documents/workspace/botla-co
make migrate-up
make migrate-down
```
