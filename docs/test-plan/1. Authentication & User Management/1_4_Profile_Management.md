# 1.4 Profile Management Test Plan

## Overview
This test plan covers user profile retrieval and the `/me` endpoint functionality.

---

## Test Cases

### 1.4.1 Get Current User Profile [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login to get access token | Token received [x] |
| 2 | GET `/api/v1/me` with valid token | 200 OK [x] |
| 3 | Response contains user info | id, email, full_name, created_at [x] |

---

### 1.4.2 Profile Includes Plan Details [x]
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/me` | 200 OK [x] |
| 2 | Response contains plan info | plan.code, plan.config [x] |
| 3 | Plan config includes limits | allowed_models, max_monthly_tokens, etc. [x] |

---

### 1.4.3 Profile Includes Usage Statistics [x]
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/me` | 200 OK [x] |
| 2 | Response contains usage stats | monthly_tokens_used, storage_used [x] |

---

### 1.4.4 Unauthorized Profile Access [x]
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/me` without token | 401 Unauthorized [x] |
| 2 | GET `/api/v1/me` with invalid token | 401 Unauthorized [x] |

---

### 1.4.5 Profile Includes Organization Memberships [x]
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/me` | 200 OK [x] |
| 2 | Response contains organizations | Array of organization memberships [x] |
| 3 | Each org includes role | owner, admin, or member [x] |

---

### 1.4.6 Cross-User Profile Access Prevention [x]
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as User A | Token A received [x] |
| 2 | GET `/api/v1/me` with Token A | Returns User A's profile [x] |
| 3 | Attempt to access User B's data | Only User A's data accessible via /me [x] |

**Note:** The `/me` endpoint only returns the authenticated user's data, so cross-user access is inherently prevented by design.

---

## How to Run Tests

### Run Profile Tests
```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/... -run "Me|Profile"
```

---

## Manual Testing

### Get Profile via cURL
```bash
# Login first
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpassword123"}' | jq -r '.token')

# Get profile
curl -X GET http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Response Schema
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "full_name": "Test User",
  "created_at": "2024-01-01T00:00:00Z",
  "plan": {
    "code": "free",
    "config": {
      "chat": {
        "allowed_models": ["gpt-4o-mini"],
        "max_monthly_tokens": 100000
      }
    }
  },
  "organizations": [
    {
      "id": "uuid",
      "name": "Personal",
      "role": "owner"
    }
  ]
}
```

---

## Coverage Notes
- Profile update functionality not currently implemented
- Consider adding profile picture/avatar support in future
