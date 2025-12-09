# 02. Token Usage & Billing Tests

> **Priority**: Critical  
> **Test Count**: 16  
> **Source Files**: `internal/rag/tokens.go`, `internal/db/analytics.go`, `internal/db/usage_ingestions.go`

---

## 2.1 Token Counting

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| TOK-001 | Count tokens for empty string | Returns 0 | ✅ |
| TOK-002 | Count tokens for Turkish text (şğıöüç) | 1.3x multiplier applied | ✅ |
| TOK-003 | Count tokens for English text | 1.0x multiplier applied | ✅ |
| TOK-004 | Token count minimum is 1 | Returns 1 for very short text | ✅ |
| TOK-005 | Token formula verification | `round((rune_count / 4) * multiplier)` | ✅ |

### Technical Notes

```go
// internal/rag/tokens.go:CountTokens
// Formula: tokens = round(utf8.RuneCountInString(text) / 4.0 * cfg.TokenMultiplier)

// Example Turkish text:
// "Türkiye'de yaşıyorum." (21 runes)
// Expected: round((21/4) * 1.3) = round(6.825) = 7 tokens
```

---

## 2.2 Monthly Usage Tracking

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| USG-001 | `GetMonthlyTokenUsage` aggregates from analytics | Correct sum for current month | ✅ |
| USG-002 | `IncrementAnalytics` upsert (initial) | Row created | ✅ |
| USG-003 | `IncrementAnalytics` upsert (update) | Counters incremented | ✅ |
| USG-004 | Token usage across multiple chatbots | Sum includes all user chatbots | ✅ |
| USG-005 | Usage resets at month boundary | Previous month excluded | ✅ |
| USG-006 | `GetMonthlyIngestionUsage` | Returns sources + embedding tokens | ✅ |
| USG-007 | `IncrementSuccessfulIngestion` | sources_count incremented | ✅ |
| USG-008 | `AddEmbeddingTokens` | embedding_tokens incremented | ✅ |

### Technical Notes

```go
// internal/db/analytics.go:GetMonthlyTokenUsage
// Query: SELECT COALESCE(SUM(a.total_tokens_used), 0) FROM analytics a
//        JOIN chatbots c ON a.chatbot_id = c.id
//        WHERE c.user_id = $1 AND a.analytics_date >= [month_start]

// internal/db/usage_ingestions.go
// Table: usage_ingestions(user_id, period_month, sources_count, embedding_tokens)
```

---

## 2.3 Quota Enforcement

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| QTA-001 | Chat when monthly tokens exceeded | 429/402 with "ERR_MONTHLY_TOKENS_EXCEEDED" | ✅ |
| QTA-002 | Ingestion when monthly limit exceeded | 429/402 with "ERR_MONTHLY_INGESTION_EXCEEDED" | ✅ |
| QTA-003 | Refresh when monthly limit exceeded | 429/402 with "ERR_MONTHLY_REFRESH_EXCEEDED" | ✅ |

### Technical Notes

```go
// Plan config fields:
// - max_monthly_tokens
// - max_monthly_ingestions
// - max_monthly_embedding_tokens
// - refresh.max_monthly_refreshes
```

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/rag/tokens_test.go` | Basic token counting |
| `internal/db/usage_ingestions_test.go` | Increment/get usage |
| `internal/integration/ingestion_quota_test.go` | Quota enforcement |
| `internal/integration/quota_enforcement_test.go` | Chat & Refresh quota enforcement |
| `internal/integration/usage_tracking_test.go` | Monthly usage tracking integration |
