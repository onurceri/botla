> Extracted from: admin-operations.md — Section: 5. Expected Behavior

# 5. Expected Behavior

### 5.1 Success Conditions

| Condition | Behavior |
|-----------|----------|
| Platform admin accesses admin endpoints | 200 OK with requested data |
| Platform admin updates user/plan/source | 200 OK with status confirmation + audit log entry |
| Health check with all systems up | 200 OK, `status: "healthy"` |
| Privacy request approval | Request processed, status updated, audit logged |
| Source reprocessed | Vectors deleted, source reset to pending, queued |

### 5.2 Failure Conditions

| Condition | Expected HTTP Status | Expected Behavior |
|-----------|---------------------|-------------------|
| No auth cookie | 401 | Unauthorized; no plan/analytics/admin data returned |
| Non-admin user on admin endpoint | 403 | Forbidden; `RequirePlatformAdmin` guard blocks |
| Non-existent user/chatbot/source ID | 404 | Resource not found |
| Empty request body on PATCH endpoints | 400 | Bad request |
| Expired privacy export download | 410 | Gone |
| Missing env-dependent service (Storage) | 503 | Service unavailable |
| Postgres down (health check) | 503 | `status: "unhealthy"` |
| Non-critical dependency down (health check) | 200 | `status: "degraded"` |
