# 7.1 Analytics Overview Test Plan

## Overview
This test plan covers the analytics overview endpoint and data aggregation.

---

## Test Cases

### 7.1.1 Get Chatbot Analytics Overview
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/chatbots/{id}/analytics` | 200 OK |
| 2 | Response includes total_messages | Count |
| 3 | Response includes total_conversations | Count |
| 4 | Response includes total_tokens_used | Count |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_detailed_test.go`
- **Setup:**
  - Create bot.
  - Insert analytics data (10 msgs, 2 convs, 500 tokens).
- **Steps:**
  1. GET `/api/v1/chatbots/{id}/analytics`.
  2. Verify 200 OK.
  3. Verify JSON fields match inserted data.

---

### 7.1.2 Feedback Counts
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit positive feedback | Stored |
| 2 | GET analytics | thumbs_up_count incremented |
| 3 | Submit negative feedback | thumbs_down_count incremented |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_detailed_test.go`
- **Setup:**
  - Insert 5 thumbs up, 3 thumbs down.
- **Steps:**
  1. GET `/analytics`.
  2. Verify `thumbs_up_count` is 5.
  3. Verify `thumbs_down_count` is 3.

---

### 7.1.3 Satisfaction Rate
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | 8 thumbs up, 2 thumbs down | Stored |
| 2 | GET analytics | feedback_rate = 80% |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_detailed_test.go`
- **Setup:**
  - Insert 8 up, 2 down.
- **Steps:**
  1. GET `/analytics`.
  2. Verify `feedback_rate` is 0.8 (or 80.0 depending on format).

---

### 7.1.4 Handoff Count
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Trigger handoff | Recorded |
| 2 | GET analytics | handoff_count incremented |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_detailed_test.go`
- **Setup:**
  - Insert 4 handoff requests.
- **Steps:**
  1. GET `/analytics`.
  2. Verify `handoff_count` is 4.

---

### 7.1.5 30-Day Aggregation
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Data from 31 days ago | Not included |
| 2 | Data from 30 days ago | Included |

**Implementation Plan:**
- **Test File:** `internal/integration/analytics_detailed_test.go`
- **Setup:**
  - Insert row at `now() - 31 days`.
  - Insert row at `now() - 10 days`.
- **Steps:**
  1. GET `/analytics`.
  2. Verify counts only reflect the 10-day old data.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Analytics"
```
