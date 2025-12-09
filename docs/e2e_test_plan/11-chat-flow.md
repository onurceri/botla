# 11. Chat Flow Tests

> **Priority**: Critical  
> **Test Count**: 18  
> **Source Files**: `internal/services/chat_service.go`, `internal/api/handlers/chat.go`

---

## 11.1 Chat Success Path

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| CHT-001 | Simple question with context | Response with sources | ✅ |
| CHT-002 | Conversation continuity | Same session_id maintains context | ✅ |
| CHT-003 | New conversation creation | New conv ID assigned | ✅ |
| CHT-004 | Message persistence | User + assistant messages stored | ✅ |
| CHT-005 | Token usage recorded | Analytics updated | ✅ |
| CHT-006 | Suggested questions shown | From chatbot config | ✅ |

---

## 11.2 Chat Failure Paths

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| ERR-001 | LLM 500 error | Fallback error message | ✅ |
| ERR-002 | Embedding API failure | Error handled | ✅ |
| ERR-003 | Qdrant search failure | Error message shown | ✅ |
| ERR-004 | Context timeout (>30s) | "İşlem tamamlanamadı..." | ✅ |
| ERR-005 | Empty context (no sources) | "Yeterli bilgi bulamadım." | ✅ |
| ERR-006 | Rate limit (429) | Retry-After header | ✅ |

---

## 11.3 Feedback

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| FBK-001 | Submit positive feedback | thumbs_up recorded | ✅ |
| FBK-002 | Submit negative feedback | thumbs_down recorded | ✅ |
| FBK-003 | Feedback for non-existent message | 404 | ✅ |
| FBK-004 | Feedback updates analytics | Counters incremented | ✅ |
| FBK-005 | Source usage tracked | message_sources populated | ✅ |
| FBK-006 | Unanswered query tracked | unanswered_queries table | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/integration/chat_test.go` | Basic flow |
| `internal/integration/chat_error_test.go` | LLM errors |
| `internal/integration/chat_timeout_test.go` | Timeouts |
| `internal/integration/feedback_test.go` | Feedback |
| `internal/integration/chat_embed_error_test.go` | Embedding errors |
| `internal/integration/chat_qdrant_error_test.go` | Qdrant errors |
| `internal/integration/qdrant_empty_hits_test.go` | Empty hits |
| `internal/integration/ratelimit_test.go` | Rate limits |
| `internal/integration/analytics_feedback_test.go` | Feedback analytics |
| `internal/integration/public_suggestions_test.go` | Suggested questions |
| `internal/integration/handoff_test.go` | Session continuity (implicit) |
