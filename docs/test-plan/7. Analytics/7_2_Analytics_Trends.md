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

---

### 7.2.2 Daily Conversation Counts
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET trends | Daily data |
| 2 | total_conversations per day | Correct counts |

---

### 7.2.3 Daily Token Usage
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET trends | Daily data |
| 2 | total_tokens_used per day | Correct counts |

---

### 7.2.4 Daily Feedback Breakdown
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET trends | Daily data |
| 2 | thumbs_up_count per day | Correct |
| 3 | thumbs_down_count per day | Correct |

---

### 7.2.5 Days Parameter
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | days=7 | 7 data points |
| 2 | days=30 | 30 data points |
| 3 | days=1 | 1 data point |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Trends"
```
