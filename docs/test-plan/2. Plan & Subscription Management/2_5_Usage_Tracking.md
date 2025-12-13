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

**Implementation Plan:**
- **Test File:** `internal/integration/usage_tracking_api_test.go`
- **Setup:**
  - Create a user and chatbot.
- **Steps:**
  1. Send a chat request (mocking usage of 150 tokens).
  2. Call `GET /api/v1/me`.
  3. Verify `monthly_tokens_used` in response matches expected value (previous + 150).

---

### 2.5.2 Embedding Token Usage Tracked
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add source to chatbot | Source processed |
| 2 | Check embedding usage | Embedding tokens tracked |

**Implementation Plan:**
- **Test File:** `internal/integration/usage_tracking_api_test.go`
- **Setup:**
  - Create a user and chatbot.
- **Steps:**
  1. Add a source (mocking embedding usage of 500 tokens).
  2. Call `GET /api/v1/me`.
  3. Verify `monthly_embedding_tokens` in response matches expected value.

---

### 2.5.3 Ingestion Count Tracked
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add URL source | Ingestion count = 1 |
| 2 | Add PDF source | Ingestion count = 2 |
| 3 | Refresh source | Ingestion count unchanged (same source) |

**Implementation Plan:**
- **Test File:** `internal/integration/usage_tracking_api_test.go`
- **Setup:**
  - Create a user and chatbot.
- **Steps:**
  1. Add a URL source. Call `/me` -> verify `monthly_ingestions` incremented.
  2. Add a PDF source. Call `/me` -> verify incremented again.
  3. Refresh the URL source. Call `/me` -> verify `monthly_ingestions` did NOT increment.

---

### 2.5.4 Storage Usage Calculated
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Upload 2MB file | Storage = 2MB |
| 2 | Upload 3MB file | Storage = 5MB |
| 3 | Delete first file | Storage = 3MB |

**Implementation Plan:**
- **Test File:** `internal/integration/usage_tracking_api_test.go`
- **Setup:**
  - Create a user and chatbot.
- **Steps:**
  1. Upload a file of known size (e.g., 2MB).
  2. Call `/me` -> verify `storage_used` is approx 2MB.
  3. Upload another 3MB file.
  4. Call `/me` -> verify `storage_used` is approx 5MB.
  5. Delete the first file.
  6. Call `/me` -> verify `storage_used` is approx 3MB.

---

### 2.5.5 Usage Resets Monthly
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Use 50,000 tokens in January | Usage = 50,000 |
| 2 | February 1st arrives | Usage reset to 0 |
| 3 | Ingestion count reset | Count reset to 0 |

**Implementation Plan:**
- **Test File:** `internal/integration/usage_tracking_api_test.go`
- **Setup:**
  - Create a user.
  - Manually insert usage records dated for the previous month.
- **Steps:**
  1. Call `GET /api/v1/me`.
  2. Verify that `monthly_tokens_used` and other monthly counters are `0`.

---

### 2.5.6 Frontend Displays Ingestion Usage
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to Plan page | Ingestion counter visible |
| 2 | Add source | Counter increments |

**Implementation Plan:**
- **Test File:** `frontend/e2e/usage_display.spec.ts`
- **Setup:**
  - Login.
- **Steps:**
  1. Go to `/settings/plan`.
  2. Locate the ingestion usage element.
  3. Verify text format (e.g., "5 / 50").
  4. Upload a file.
  5. Verify text updates (e.g., "6 / 50").

---

### 2.5.7 Frontend Displays Storage Usage
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to Plan page | Storage usage displayed |
| 2 | Shows used/total format | e.g., "5MB / 10MB" |
| 3 | Progress bar accurate | Visual representation correct |

**Implementation Plan:**
- **Test File:** `frontend/e2e/usage_display.spec.ts`
- **Setup:**
  - Login.
- **Steps:**
  1. Go to `/settings/plan`.
  2. Verify storage usage text and progress bar exist.
  3. Validate the percentage width of the progress bar matches the usage ratio.

---

### 2.5.8 Frontend Displays Token Usage
**Priority:** Medium  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Navigate to Dashboard | Token usage displayed |
| 2 | Shows used/limit format | e.g., "45,000 / 100,000" |

**Implementation Plan:**
- **Test File:** `frontend/e2e/usage_display.spec.ts`
- **Setup:**
  - Login.
- **Steps:**
  1. Go to `/dashboard`.
  2. Locate the token usage card.
  3. Verify the text matches the user's current usage state.

---

### 2.5.9 Exceeding Limits Returns Clear Error
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Exceed monthly token limit | 429 Too Many Requests |
| 2 | Response message | Clear explanation with upgrade CTA |
| 3 | Exceed ingestion limit | 403 with clear message |

**Implementation Plan:**
- **Test File:** `internal/integration/usage_tracking_api_test.go`
- **Setup:**
  - User with maxed out tokens.
- **Steps:**
  1. Trigger chat request.
  2. Verify 429.
  3. Parse body, check for `code: ERR_MONTHLY_LIMIT` (or similar) and a user-friendly message key.
  4. Repeat for ingestion limit (403 or 402) and verify message.

---

### 2.5.10 Usage Survives Database Restart
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Record usage | Usage tracked |
| 2 | Restart database | N/A |
| 3 | Query usage | Values preserved |

**Implementation Plan:**
- **Note:** This is implicitly covered by all integration tests as they run against a persistent (albeit temporary for tests) database. Explicit restart testing is usually reserved for infrastructure/ops tests.
- **Test File:** `internal/integration/usage_tracking_api_test.go`
- **Steps:**
  1. Record usage.
  2. Close DB connection pool.
  3. Re-open DB connection pool.
  4. Verify usage values.

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
