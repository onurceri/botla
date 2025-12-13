# 19.1 Disaster Recovery Test Plan

## Overview
This test plan covers backup and recovery procedures.

---

## Test Cases

### 19.1.1 Database Backup
**Priority:** Critical  
**Type:** Operations Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create backup | Backup file created |
| 2 | Verify backup | Data intact |

**Implementation Plan:**
- **Test Script:** `scripts/test_backup.sh`
- **Steps:**
  1. Populate DB with test data.
  2. Run `pg_dump ... > test_backup.sql`.
  3. Verify file exists and size > 0.
  4. Grep file for inserted data.

---

### 19.1.2 Database Restore
**Priority:** Critical  
**Type:** Operations Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Restore from backup | Data restored |
| 2 | Verify data | All records present |

**Implementation Plan:**
- **Test Script:** `scripts/test_backup.sh`
- **Steps:**
  1. Drop current test DB.
  2. Create new DB.
  3. Run `psql ... < test_backup.sql`.
  4. Connect and verify row counts match original.

---

### 19.1.3 Service Failure Handling
**Priority:** High  
**Type:** Resilience Test

| Failure | Expected |
|---------|----------|
| Database down | Error shown, app doesn't crash |
| Redis down | Fallback to memory cache |
| OpenRouter down | Error message to user |
| Qdrant down | Graceful degradation |

**Implementation Plan:**
- **Test File:** `internal/integration/recovery_test.go`
- **Steps:**
  1. Mock DB connection failure (or close pool).
  2. Call `/health` -> Expect 503.
  3. Call `/chat` -> Expect 503 (App remains running).
  4. Mock Qdrant failure.
  5. Call `/chat`. Expect 200 (if fallback enabled) or 503 with specific message.

---

## How to Run Tests

```bash
# Backup PostgreSQL
pg_dump -U postgres botla > backup.sql

# Restore
psql -U postgres botla < backup.sql
```
