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

---

### 19.1.2 Database Restore
**Priority:** Critical  
**Type:** Operations Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Restore from backup | Data restored |
| 2 | Verify data | All records present |

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

---

## How to Run Tests

```bash
# Backup PostgreSQL
pg_dump -U postgres botla > backup.sql

# Restore
psql -U postgres botla < backup.sql
```
