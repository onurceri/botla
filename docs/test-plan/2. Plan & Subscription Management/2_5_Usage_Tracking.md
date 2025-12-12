# 2.5 Usage Tracking & Display Test Plan

## Overview
This test plan covers usage tracking, monthly resets, and frontend display of usage statistics.

---

## Test Cases

### 2.5.1 Monthly Token Usage Tracked
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Chat with chatbot | Response received |
| 2 | Check usage tracking | Tokens added to monthly_tokens_used |
| 3 | GET `/api/v1/me` | Usage reflected in response |

---

### 2.5.2 Embedding Token Usage Tracked
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add source to chatbot | Source processed |
| 2 | Check embedding usage | Embedding tokens tracked |

---

### 2.5.3 Ingestion Count Tracked
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL source | Ingestion count = 1 |
| 2 | Add PDF source | Ingestion count = 2 |
| 3 | Refresh source | Ingestion count unchanged (same source) |

---

### 2.5.4 Storage Usage Calculated
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload 2MB file | Storage = 2MB |
| 2 | Upload 3MB file | Storage = 5MB |
| 3 | Delete first file | Storage = 3MB |

---

### 2.5.5 Usage Resets Monthly
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use 50,000 tokens in January | Usage = 50,000 |
| 2 | February 1st arrives | Usage reset to 0 |
| 3 | Ingestion count reset | Count reset to 0 |

---

### 2.5.6 Frontend Displays Ingestion Usage
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to Plan page | Ingestion counter visible |
| 2 | Add source | Counter increments |

---

### 2.5.7 Frontend Displays Storage Usage
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to Plan page | Storage usage displayed |
| 2 | Shows used/total format | e.g., "5MB / 10MB" |
| 3 | Progress bar accurate | Visual representation correct |

---

### 2.5.8 Frontend Displays Token Usage
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to Dashboard | Token usage displayed |
| 2 | Shows used/limit format | e.g., "45,000 / 100,000" |

---

### 2.5.9 Exceeding Limits Returns Clear Error
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Exceed monthly token limit | 429 Too Many Requests |
| 2 | Response message | Clear explanation with upgrade CTA |
| 3 | Exceed ingestion limit | 403 with clear message |

---

### 2.5.10 Usage Survives Database Restart
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Record usage | Usage tracked |
| 2 | Restart database | N/A |
| 3 | Query usage | Values preserved |

---

## How to Run Tests

```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/... -run "Usage|Token|Storage"
```

---

## Database Schema

```sql
-- Usage is typically tracked in:
-- - users table (monthly_tokens_used, monthly_ingestions)
-- - Or separate usage_tracking table

SELECT 
  monthly_tokens_used,
  monthly_embedding_tokens,
  monthly_ingestions,
  storage_used_bytes
FROM user_usage
WHERE user_id = ?;
```
