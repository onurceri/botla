> Extracted from: plans-entitlements.md — Section: 5. Expected Behavior

# 5. Expected Behavior

### 5.1 Success Conditions

| Condition | Behavior |
|-----------|----------|
| Valid request to public plan endpoints | 200 OK with complete plan data including limits and features |
| Authenticated user fetches own plan | 200 OK with plan + available models resolved per plan |
| Authenticated user fetches own usage | 200 OK with usage statistics |
| Plan cache invalidation | 200 OK, subsequent request fetches fresh data |
| Token quota within limits | Request proceeds normally |

### 5.2 Failure Conditions

| Condition | Expected HTTP Status | Expected Behavior |
|-----------|---------------------|-------------------|
| No auth cookie | 401 | Unauthorized; no plan/analytics/admin data returned |
| Invalid plan code | 404 | Plan not found |
| Token quota exceeded | — | `ReserveTokens` returns error; chat request blocked |
| Rate limit exceeded | 429 | Too Many Requests (middleware layer) |
| Plan validation failure at startup | — | Application fails to boot; combined validation errors logged |
| Redis cache write failure | — | Logged but non-fatal; in-memory cache used |
