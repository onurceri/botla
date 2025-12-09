# 07. Chatbot Lifecycle Tests

> **Priority**: High  
> **Test Count**: 14  
> **Source Files**: `internal/db/chatbot.go`, `internal/api/handlers/chatbot.go`

---

## 7.1 Creation

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| BOT-001 | Create with valid name | 201, chatbot returned | ✅ |
| BOT-002 | Create with duplicate name (same user) | Unique slug generated | ✅ |
| BOT-003 | Create with all optional fields | All fields persisted | ✅ |
| BOT-004 | Create without name | 400 Bad Request | ✅ |
| BOT-005 | Default values applied | temperature: 0.7, max_tokens: 4096 | ✅ |

---

## 7.2 Configuration

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| CFG-001 | Update theme_color | Persisted and returned | ✅ |
| CFG-002 | Update welcome_message (Turkish chars) | Chars preserved | ✅ |
| CFG-003 | Update suggested_questions array | Array persisted | ✅ |
| CFG-004 | Update confidence_threshold | Applied to RAG search | ✅ |

---

## 7.3 Deletion

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| DEL-001 | Delete chatbot | Soft deleted | ✅ |
| DEL-002 | Cascade delete sources | Sources removed | ✅ |
| DEL-003 | Cascade delete analytics | Analytics removed | ✅ |
| DEL-004 | Cascade delete from Qdrant | Embeddings removed | ✅ |
| DEL-005 | Delete non-existent chatbot | 404 Not Found | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/db/chatbot_test.go` | CRUD operations |
| `internal/integration/chatbot_test.go` | API flow |
| `internal/integration/lifecycle_test.go` | Full lifecycle coverage (BOT-001 to DEL-005) |
