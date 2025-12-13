# 14.1 Edge Cases Test Plan

## Overview
This test plan covers edge cases and error handling scenarios.

---

## Network & Connectivity

| Test | Action | Expected Result |
|------|--------|-----------------|
| API timeout | Server slow | Timeout handled |
| Network offline | No connection | Error displayed |
| Retry logic | Transient failure | Automatic retry |

**Implementation Plan:**
- **Test File:** `internal/integration/edge_cases_test.go`
- **Setup:**
  - Mock external service (OpenAI/Qdrant) with latency > timeout.
- **Steps:**
  1. Trigger action requiring external call.
  2. Verify 504 Gateway Timeout or graceful error message.

---

## Data Edge Cases

| Test | Action | Expected Result |
|------|--------|-----------------|
| Empty strings | Submit "" | Validation or handled |
| Null values | Missing fields | Defaults used |
| Long strings | 10,000 char name | Truncated or rejected |
| Special characters | Emojis, Unicode | Handled correctly |
| Large files | Max size | Rejected with message |

**Implementation Plan:**
- **Test File:** `internal/integration/edge_cases_test.go`
- **Steps:**
  1. Create bot with `name=""`. Expect 400.
  2. Create bot with `name=null`. Expect 400.
  3. Create bot with `name="A" * 10000`. Expect 400.
  4. Create bot with `name="🤖"`. Expect 201.

---

## Quota Edge Cases

| Test | Action | Expected Result |
|------|--------|-----------------|
| At exact limit | Add last item | Succeeds |
| Over limit | Add one more | Rejected |
| Race condition | Concurrent adds | Only one succeeds |
| Monthly reset | New month | Quota restored |

**Implementation Plan:**
- **Test File:** `internal/integration/edge_cases_test.go`
- **Setup:**
  - User with limit=5. Usage=4.
- **Steps:**
  1. Add item -> 201. Usage=5.
  2. Add item -> 403.
  3. Reset usage to 4. Spawn 5 concurrent goroutines adding items.
  4. Verify only 1 succeeded (or usage is accurately capped).

---

## State Edge Cases

| Test | Action | Expected Result |
|------|--------|-----------------|
| Update deleted chatbot | API call | 404 Not Found |
| Refresh processing source | API call | 409 Conflict |
| Delete processing source | API call | Handled appropriately |

**Implementation Plan:**
- **Test File:** `internal/integration/edge_cases_test.go`
- **Steps:**
  1. Delete bot. Try PUT update -> 404.
  2. Create source, set status "processing". Try Refresh -> 409.
  3. Delete "processing" source. Verify 200/204 and cleanup.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Edge|Error"
```
