# 11.2 Qdrant Integration Test Plan

## Overview
This test plan covers the Qdrant vector database integration.

---

## Test Cases

### 11.2.1 Connection
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | QDRANT_URL configured | Connection successful |
| 2 | Invalid URL | Error on startup |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_qdrant_test.go`
- **Steps:**
  1. Init Qdrant client with valid URL -> OK.
  2. Init Qdrant client with invalid URL -> Error.

---

### 11.2.2 Collection Management
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot | Collection created |
| 2 | Delete chatbot | Collection deleted |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_qdrant_test.go`
- **Setup:**
  - Mock Qdrant API.
- **Steps:**
  1. Create chatbot.
  2. Verify `CreateCollection` called on mock.
  3. Delete chatbot.
  4. Verify `DeleteCollection` called on mock.

---

### 11.2.3 Embeddings Upsert
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Process source | Embeddings stored |
| 2 | Query Qdrant | Vectors exist |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_qdrant_test.go`
- **Setup:**
  - Mock Qdrant.
- **Steps:**
  1. Call service `UpsertChunks`.
  2. Verify mock received points with vectors.

---

### 11.2.4 Vector Search
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send chat query | Qdrant searched |
| 2 | Results returned | Relevant chunks |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_qdrant_test.go`
- **Steps:**
  1. Call service `Search`.
  2. Verify mock received search query.
  3. Verify returned payload matches.

---

### 11.2.5 Top-K Parameter
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | top_k = 3 | Max 3 results |
| 2 | top_k = 10 | Max 10 results |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_qdrant_test.go`
- **Steps:**
  1. Call `Search` with `limit=3`. Verify result count <= 3.
  2. Call `Search` with `limit=10`. Verify result count <= 10.

---

### 11.2.6 Error Handling
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Qdrant unavailable | Error handled |
| 2 | User sees message | Fallback behavior |

**Implementation Plan:**
- **Test File:** `internal/integration/integration_qdrant_test.go`
- **Setup:**
  - Mock returns error.
- **Steps:**
  1. Call `Search`.
  2. Verify error is returned (not panic).
  3. Verify system handles it gracefully (e.g. empty results log warning).

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Qdrant|Vector"
```
