> Extracted from: analytics.md — Section: 5. Expected Behavior

# 5. Expected Behavior

### 5.1 Success Conditions

| Condition | Behavior |
|-----------|----------|
| Auth user fetches analytics for owned chatbot | 200 OK with aggregated data |

### 5.2 Failure Conditions

| Condition | Expected HTTP Status | Expected Behavior |
|-----------|---------------------|-------------------|
| No auth cookie | 401 | Unauthorized; no plan/analytics/admin data returned |
| Non-existent user/chatbot/source ID | 404 | Resource not found |
