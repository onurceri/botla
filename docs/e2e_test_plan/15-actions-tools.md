# 15. Actions & Tools Tests

> **Priority**: Medium  
> **Test Count**: 8  
> **Source Files**: `internal/rag/tools.go`, `internal/rag/tool_executor.go`

---

## 15.1 Tool-Enabled Chat

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| ACT-001 | Chat with enabled actions | Tool calls work | ✅ `TestChatWithTools` |
| ACT-002 | Agentic loop (max 5 iterations) | Terminates correctly | ✅ `TestAgenticLoopLimit` |
| ACT-003 | Tool execution error | Error in tool result | ✅ `TestToolExecutionError` |
| ACT-004 | HTTP action (GET) | External API called | ✅ `TestChatWithTools` |
| ACT-005 | HTTP action (POST) | Request body sent | ✅ `TestHTTPActionPOST` |
| ACT-006 | Builtin tools | Default tools work | ✅ `TestBuiltinTools` |
| ACT-007 | Action management CRUD | Create/update/delete | ✅ `TestAction_CRUD` |
| ACT-008 | Action enabled/disabled | Toggle respected | ✅ `TestDisabledAction` |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/integration/action_test.go` | Full coverage (ACT-001 to ACT-008) |
