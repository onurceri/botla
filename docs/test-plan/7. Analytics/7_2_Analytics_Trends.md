# 7.2 Analytics Trends Test Plan

## Overview
This test plan covers time-series analytics data for charts and reporting.

---

## Test Cases

### 7.2.1 Daily Message Counts
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}/analytics/trends?days=7` | 200 OK |
| 2 | Response includes daily breakdown | 7 data points |
| 3 | Each day has total_messages | Count per day |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_series_test.go`
- **Setup:**
  - Insert analytics data for `now()`, `now()-1day`, `now()-2days`.
- **Steps:**
  1. GET `trends?days=7`.
  2. Verify response array length is 7.
  3. Verify `total_messages` matches inserted data for specific dates.

---

### 7.2.2 Daily Conversation Counts
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET trends | Daily data |
| 2 | total_conversations per day | Correct counts |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_series_test.go`
- **Setup:**
  - Insert conversation data distributed over days.
- **Steps:**
  1. GET `trends`.
  2. Verify `total_conversations` field in each day object matches DB.

---

### 7.2.3 Daily Token Usage
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET trends | Daily data |
| 2 | total_tokens_used per day | Correct counts |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_series_test.go`
- **Setup:**
  - Insert token usage data.
- **Steps:**
  1. GET `trends`.
  2. Verify `total_tokens_used` sums correctly per day.

---

### 7.2.4 Daily Feedback Breakdown
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET trends | Daily data |
| 2 | thumbs_up_count per day | Correct |
| 3 | thumbs_down_count per day | Correct |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_series_test.go`
- **Setup:**
  - Insert feedback with timestamps.
- **Steps:**
  1. GET `trends`.
  2. Verify feedback counts per day.

---

### 7.2.5 Days Parameter
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | days=7 | 7 data points |
| 2 | days=30 | 30 data points |
| 3 | days=1 | 1 data point |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_series_test.go`
- **Steps:**
  1. GET `trends?days=7`. Verify len=7.
  2. GET `trends?days=30`. Verify len=30.
  3. GET `trends?days=1`. Verify len=1.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Trends"
```
