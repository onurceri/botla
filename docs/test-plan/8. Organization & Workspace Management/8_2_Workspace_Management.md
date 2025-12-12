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

---

### 8.2.2 Duplicate Slug Within Org
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create workspace "dev" | Success |
| 2 | Create another "dev" in same org | 409 Conflict |
| 3 | Create "dev" in different org | Success |

---

### 8.2.3 List Workspaces
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/organizations/{id}/workspaces` | 200 OK |
| 2 | All org workspaces returned | Array |

---

### 8.2.4 Update Workspace
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Update name, slug | 200 OK |
| 2 | Changes persisted | Query confirms |

---

### 8.2.5 Delete Workspace
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Delete workspace | 200 OK |
| 2 | Cascade to chatbots | All deleted |
| 3 | Cannot delete last workspace | 400 Bad Request |

---

### 8.2.6 Chatbots in Workspace Context
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create chatbot with workspace_id | Assigned |
| 2 | GET chatbots with X-Workspace-ID | Filtered |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Workspace"
```
