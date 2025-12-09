# 01. RAG Pipeline Tests

> **Priority**: Critical  
> **Test Count**: 24  
> **Source Files**: `internal/rag/search.go`, `internal/rag/chunker.go`, `internal/rag/embedding.go`

---

## 1.1 Context Search & Retrieval

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| RAG-001 | Search with valid embedding | Returns top-k results sorted by score | ✅ |
| RAG-002 | Search with empty embedding | Returns empty result, no crash | ✅ |
| RAG-003 | Search with invalid chatbot ID | Returns empty result | ✅ |
| RAG-004 | Search respects `confidenceThreshold` (0.3 default) | Filters results below threshold | ✅ |
| RAG-005 | Search respects `topK` limit (default: 5, env: `RAG_TOPK`) | Returns max topK results | ✅ |
| RAG-006 | Search respects `maxContextTokens` (default: 2000) | Stops aggregating at token limit | ✅ |
| RAG-007 | Context scoring with Turkish text | Score calculation uses 1.3x multiplier | ✅ |
| RAG-008 | Score threshold at exactly 0.0 (permissive) | Returns all results | ✅ |
| RAG-009 | Score threshold at 1.0 (restrictive) | Returns no results | ✅ |
| RAG-010 | Context aggregation separator | Multiple chunks separated by `\n---\n` | ✅ |

### Technical Notes

```go
// Key function: internal/rag/search.go:SearchContext
// Parameters to test:
// - threshold (chatbot.ConfidenceThreshold)
// - topK (from RAGConfig or env RAG_TOPK, default: 5)
// - maxCtx (from RAGConfig or env RAG_MAX_CONTEXT_TOKENS, default: 2000)

// Context header (Turkish):
// "Aşağıdaki belgeler sorgularına cevap vermek için kullanılmıştır:\n\n"
```

---

## 1.2 Chunking Algorithm

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| CHK-001 | Chunk empty text | Returns nil, no error | ✅ |
| CHK-002 | Chunk with `targetTokens <= 0` | Returns error: "targetTokens must be > 0" | ✅ |
| CHK-003 | Chunk respects paragraph boundaries | Splits on `\n\n` | ✅ |
| CHK-004 | Chunk respects sentence boundaries | Uses sentence tokenizer | ✅ |
| CHK-005 | Chunk with Turkish abbreviations | Does NOT split on "Dr.", "Prof.", "vb." | ✅ |
| CHK-006 | Chunk with English abbreviations | Does NOT split on "Mr.", "Ms.", "Inc." | ✅ |
| CHK-007 | Chunk ~15% tail overlap | Overlap calculated correctly | ✅ |
| CHK-008 | Very long sentence exceeding targetTokens | Sentence emitted alone | ✅ |
| CHK-009 | Token counting Turkish (1.3x multiplier) | Correct estimation | ✅ |
| CHK-010 | Token counting English (1.0x multiplier) | Correct estimation | ✅ |

### Technical Notes

```go
// Turkish abbreviations from pkg/langconfig/config.go:
// "Dr.", "Prof.", "vb.", "Av.", "Ecz.", "Doç.", "Yrd.", "Cad.", "Sok.", "Mah."

// Test sentence that SHOULD NOT split:
// "Sayın Dr. Ahmet Bey geldi." → 1 sentence
```

---

## 1.3 Embedding Generation

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| EMB-001 | Generate embeddings for 0 chunks | Returns nil, no error | ✅ |
| EMB-002 | Generate embeddings for 25 chunks (batch limit) | Single batch processed | ✅ |
| EMB-003 | Generate embeddings for 26 chunks | Two batches processed | ✅ |
| EMB-004 | Rate limiting (58 req/sec ticker) | No rate limit errors | ✅ |

### Technical Notes

```go
// Batch size: 25 chunks per API request
// Rate limiter: time.NewTicker(time.Second / 58)
// Retry: Once on failure before returning error
```

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/rag/search_test.go` | Thresholds (incl. 0.0/1.0), separators, max tokens |
| `internal/rag/chunker_test.go` | Basic chunking, abbreviations, overlap |
| `internal/rag/embedding_test.go` | Batch processing, empty inputs |
| `internal/rag/pipeline_missing_test.go` | Invalid ID, Turkish scoring, sentence/para splitting |
