# 5.2 RAG (Retrieval Augmented Generation) Test Plan

## Overview
This test plan covers the RAG pipeline including vector search and confidence scoring.

---

## Test Cases

### 5.2.1 Query Qdrant for Relevant Chunks
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add source about "pricing" | Processed |
| 2 | Ask "What is the pricing?" | Relevant chunks retrieved |
| 3 | Sources used in response | Pricing source included |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Bot with 1 processed source.
- **Steps:**
  1. Send chat message "pricing".
  2. Verify mock Qdrant `Search` method was called.
  3. Verify mock Qdrant returned dummy chunks.
  4. Verify response includes source metadata.

---

### 5.2.2 Top-K Limit Respected
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free plan: top_k = 3 | Config |
| 2 | Send query | Max 3 chunks used |
| 3 | Pro plan: top_k = 5 | Max 5 chunks used |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Free Bot, Pro Bot.
- **Steps:**
  1. Free Bot: Chat. Verify Qdrant `limit` param is 3.
  2. Pro Bot: Chat. Verify Qdrant `limit` param is 5.

---

### 5.2.3 Max Context Tokens Respected
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Free plan: max_context = 2000 | Config |
| 2 | Send query | Context <= 2000 tokens |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Free Bot. Mock Qdrant returning large chunks totaling 5000 tokens.
- **Steps:**
  1. Chat.
  2. Verify the `system` or `context` prompt passed to LLM mock is truncated to ~2000 tokens.

---

### 5.2.4 Confidence Tiers
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Query with high match | confidence_tier = "high" |
| 2 | Query with medium match | confidence_tier = "medium" |
| 3 | Query with low match | confidence_tier = "low" |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Bot with thresholds: High=0.8, Medium=0.5.
- **Steps:**
  1. Mock Qdrant score 0.9. Chat. Verify `confidence_tier: "high"`.
  2. Mock Qdrant score 0.6. Chat. Verify `confidence_tier: "medium"`.
  3. Mock Qdrant score 0.2. Chat. Verify `confidence_tier: "low"`.

---

### 5.2.5 High Confidence Response
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | High confidence query | Normal response |
| 2 | No warning message | Clean answer |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Bot with `show_confidence_warning=true`.
- **Steps:**
  1. Mock high score.
  2. Chat.
  3. Verify response text does NOT contain warning prefix.

---

### 5.2.6 Medium Confidence Warning
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | show_confidence_warning = true | Config |
| 2 | Medium confidence query | Warning included |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Bot with `show_confidence_warning=true`.
- **Steps:**
  1. Mock medium score.
  2. Chat.
  3. Verify response text starts with (or contains) warning message.

---

### 5.2.7 Low Confidence Fallback
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Low confidence query | Fallback triggered |
| 2 | fallback_mode = "static" | Static message returned |
| 3 | fallback_mode = "smart" | AI-generated fallback |
| 4 | fallback_mode = "escalate" | Handoff triggered |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Bot with low thresholds.
- **Steps:**
  1. Set `fallback_mode="static"`. Mock low score. Chat. Verify `fallback_messages.no_info` returned.
  2. Set `fallback_mode="smart"`. Chat. Verify LLM called with fallback prompt.
  3. Set `fallback_mode="escalate"`. Chat. Verify `request_human_handoff` tool triggered.

---

### 5.2.8 No Sources Found
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Query unrelated topic | No chunks found |
| 2 | Response | no_info_found message |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Mock Qdrant returning empty list.
- **Steps:**
  1. Chat.
  2. Verify response matches `no_info_found` message.

---

### 5.2.9 Topic Restrictions Applied
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Block topic "politics" | Config |
| 2 | Ask about politics | blocked_message returned |

**Implementation Plan:**
- **Test File:** `internal/integration/chat_rag_test.go`
- **Setup:**
  - Bot with blocked topic "politics".
  - Mock LLM moderation check to return `blocked=true`.
- **Steps:**
  1. Chat.
  2. Verify response matches `blocked_message`.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "RAG|Confidence"
```
