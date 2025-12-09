# 12. Analytics Tests

> **Priority**: Medium  
> **Test Count**: 10  
> **Source Files**: `internal/db/analytics.go`, `internal/api/handlers/analytics.go`

---

## 12.1 Data Collection

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| ANL-001 | Message count incremented | total_messages +2 | ✅ |
| ANL-002 | Conversation count on new conv | total_conversations +1 | ✅ |
| ANL-003 | Token usage tracked | total_tokens_used updated | ✅ |
| ANL-004 | Feedback counts tracked | thumbs_up/down updated | ✅ |
| ANL-005 | Handoff count tracked | handoff_count updated | ✅ |
| ANL-006 | Response time tracked | Stored in ms | ✅ |

---

## 12.2 API Endpoints

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| API-001 | Get overview (30 days) | Aggregated stats | ✅ |
| API-002 | Get time series data | Daily breakdown | ✅ |
| API-003 | Filter by date range | Correct filtering | ✅ |
| API-004 | Empty analytics (new bot) | Zero values | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/integration/analytics_*.go` | All scenarios |
