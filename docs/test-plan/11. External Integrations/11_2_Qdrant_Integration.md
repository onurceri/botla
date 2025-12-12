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

---

### 11.2.2 Collection Management
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot | Collection created |
| 2 | Delete chatbot | Collection deleted |

---

### 11.2.3 Embeddings Upsert
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Process source | Embeddings stored |
| 2 | Query Qdrant | Vectors exist |

---

### 11.2.4 Vector Search
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Send chat query | Qdrant searched |
| 2 | Results returned | Relevant chunks |

---

### 11.2.5 Top-K Parameter
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | top_k = 3 | Max 3 results |
| 2 | top_k = 10 | Max 10 results |

---

### 11.2.6 Error Handling
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Qdrant unavailable | Error handled |
| 2 | User sees message | Fallback behavior |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Qdrant|Vector"
```
