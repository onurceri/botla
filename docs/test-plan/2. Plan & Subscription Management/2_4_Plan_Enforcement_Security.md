# 2.4 Plan Enforcement Security Test Plan

## Overview
This test plan covers security-critical tests ensuring plan limits cannot be bypassed via API manipulation.

---

## Test Cases

### 2.4.1 Free User Cannot Enable Secure Embed via API
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as Free plan user | Token received |
| 2 | PUT `/api/v1/chatbots/{id}` with `secure_embed_enabled: true` | 403 Forbidden |
| 3 | Response body | `{"feature": "secure_embed", "message": "Upgrade required"}` |

---

### 2.4.2 Free User Cannot Set Auto-Refresh via API
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as Free plan user | Token received |
| 2 | PUT `/api/v1/chatbots/{id}` with `refresh_policy: "auto"` | 403 Forbidden |

---

### 2.4.3 Free User Cannot Enable Discovery Mode via API
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as Free plan user | Token received |
| 2 | PUT `/api/v1/chatbots/{id}` with `discovery_mode: "auto"` | 403 Forbidden |

---

### 2.4.4 Free User Cannot Select gpt-4o via API
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as Free plan user | Token received |
| 2 | POST create chatbot with `model: "gpt-4o"` | 403 or model coerced to gpt-4o-mini |
| 3 | PUT update chatbot to `model: "gpt-4o"` | 403 Forbidden |

---

### 2.4.5 Free User Cannot Upload Oversized Files via API
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as Free plan user | Token received |
| 2 | Upload 6MB PDF file | 413 Payload Too Large |
| 3 | Verify file not stored | No file in S3/storage |

---

### 2.4.6 Free User Cannot Exceed URL Limit via API
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as Free plan user | Token received |
| 2 | Add 1 URL source | 201 Created |
| 3 | Add 2nd URL source | 403 Forbidden |

---

### 2.4.7 Pro User Cannot Select Claude via API
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Login as Pro plan user | Token received |
| 2 | PUT update chatbot to `model: "claude-3-5-sonnet"` | 403 Forbidden |

---

### 2.4.8 Plan Bypass Returns Proper Error Format
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Attempt any plan bypass | 403 Forbidden |
| 2 | Response format | `{"error": "...", "feature": "...", "upgrade_required": true}` |

---

### 2.4.9 Backend Validates All Plan Limits
**Priority:** Critical  
**Type:** Security Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Disable frontend validation (DevTools) | N/A |
| 2 | Submit requests exceeding limits | All rejected by backend |
| 3 | No data persisted | Database unchanged |

---

### 2.4.10 Frontend Limits Match Backend
**Priority:** High  
**Type:** E2E Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Check Free plan UI limits | Matches backend config |
| 2 | Check Pro plan UI limits | Matches backend config |
| 3 | Check Ultra plan UI limits | Matches backend config |

---

## How to Run Tests

### Run Security Integration Tests
```bash
cd /Users/onur/Documents/workspace/botla-co
go test -v ./internal/integration/... -run "Security|Enforcement|Bypass"
```

### Existing Security Tests
- `internal/integration/plan_enforcement_test.go` (if exists)
- Security regression tests in `docs/plans/security-regression-suite.md`

---

## Manual Security Testing

### Test with cURL - Attempt Plan Bypass
```bash
# Get Free user token
FREE_TOKEN="..."

# Attempt to enable secure_embed (should fail)
curl -X PUT http://localhost:8080/api/v1/chatbots/{id} \
  -H "Authorization: Bearer $FREE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"secure_embed_enabled": true}'

# Expected: 403 Forbidden
```

---

## Coverage Notes
- All plan enforcement should be tested at the handler/service level
- Consider fuzzing inputs to find bypass vectors
- Regression tests should run on every PR
