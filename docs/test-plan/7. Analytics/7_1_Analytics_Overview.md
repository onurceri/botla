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

---

### 7.1.2 Feedback Counts
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Submit positive feedback | Stored |
| 2 | GET analytics | thumbs_up_count incremented |
| 3 | Submit negative feedback | thumbs_down_count incremented |

---

### 7.1.3 Satisfaction Rate
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | 8 thumbs up, 2 thumbs down | Stored |
| 2 | GET analytics | feedback_rate = 80% |

---

### 7.1.4 Handoff Count
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Trigger handoff | Recorded |
| 2 | GET analytics | handoff_count incremented |

---

### 7.1.5 30-Day Aggregation
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Data from 31 days ago | Not included |
| 2 | Data from 30 days ago | Included |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Analytics"
```
