# 8.2 Workspace Management Test Plan

## Overview
This test plan covers workspace CRUD within organizations.

---

## Test Cases

### 8.2.1 Create Workspace
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/organizations/{id}/workspaces` | 201 Created |
| 2 | Workspace belongs to org | organization_id set |
| 3 | Slug unique within org | Enforced |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Create Org.
- **Steps:**
  1. POST `/organizations/{id}/workspaces` with `{"name": "Dev", "slug": "dev"}`.
  2. Verify 201.
  3. Verify `organization_id` in DB matches.

---

### 8.2.2 Duplicate Slug Within Org
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create workspace "dev" | Success |
| 2 | Create another "dev" in same org | 409 Conflict |
| 3 | Create "dev" in different org | Success |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Create Org A, Org B.
- **Steps:**
  1. Create "dev" in Org A. Expect 201.
  2. Create "dev" in Org A. Expect 409.
  3. Create "dev" in Org B. Expect 201.

---

### 8.2.3 List Workspaces
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/organizations/{id}/workspaces` | 200 OK |
| 2 | All org workspaces returned | Array |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Create Org with 2 workspaces.
- **Steps:**
  1. GET `/organizations/{id}/workspaces`.
  2. Verify array length 2.

---

### 8.2.4 Update Workspace
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Update name, slug | 200 OK |
| 2 | Changes persisted | Query confirms |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Create Workspace.
- **Steps:**
  1. PUT `/workspaces/{id}` with new name/slug.
  2. Verify 200.
  3. Fetch workspace and verify changes.

---

### 8.2.5 Delete Workspace
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Delete workspace | 200 OK |
| 2 | Cascade to chatbots | All deleted |
| 3 | Cannot delete last workspace | 400 Bad Request |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Create Org with 2 workspaces.
  - Create bot in WS 1.
- **Steps:**
  1. DELETE WS 1. Expect 200.
  2. Verify bot is deleted.
  3. DELETE WS 2 (last one). Expect 400 (if business rule exists).

---

### 8.2.6 Chatbots in Workspace Context
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot with workspace_id | Assigned |
| 2 | GET chatbots with X-Workspace-ID | Filtered |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Create WS.
- **Steps:**
  1. Create chatbot with `X-Workspace-ID`.
  2. Verify `workspace_id` column.
  3. List chatbots with `X-Workspace-ID`. Verify returned.
  4. List chatbots without header (or different header). Verify NOT returned (or filtered correctly).

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Workspace"
```
