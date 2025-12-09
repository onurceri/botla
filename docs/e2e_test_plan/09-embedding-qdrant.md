# 09. Embedding & Qdrant Tests

> **Priority**: High  
> **Test Count**: 12  
> **Source Files**: `internal/rag/qdrant.go`, `internal/rag/embedding.go`

---

## 9.1 Qdrant Integration

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| QDR-001 | Collection creation | Collection exists | ✅ |
| QDR-002 | Upsert embedding | Point stored | ✅ |
| QDR-003 | Search with filter (chatbot_id) | Correct filtering | ✅ |
| QDR-004 | Delete by source_id | Points removed | ✅ |
| QDR-005 | Qdrant timeout (>30s) | Error handled | ✅ |
| QDR-006 | Qdrant unavailable | Graceful degradation | ✅ |
| QDR-007 | Empty search results | "No info found" | ✅ |

---

## 9.2 Batch Processing

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| BAT-001 | Batch size 25 | Single request | ✅ |
| BAT-002 | Batch size 50 | Two requests | ✅ |
| BAT-003 | Rate limiting | No 429 from OpenAI | ✅ |
| BAT-004 | Point ID format | Valid UUID v4 | ✅ |
| BAT-005 | Retry on failure | One retry before error | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/rag/qdrant_test.go` | Client operations |
| `internal/integration/qdrant_*.go` | Error scenarios |
