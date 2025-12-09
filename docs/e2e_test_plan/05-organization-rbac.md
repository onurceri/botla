# 05. Organization & Workspace RBAC Tests

> **Priority**: Critical  
> **Test Count**: 25  
> **Source Files**: `internal/services/organization_service.go`, `pkg/middleware/organization.go`

---

## 5.1 Role Hierarchy

| Role | Weight | Capabilities |
|------|--------|--------------|
| `owner` | 3 | Everything |
| `admin` | 2 | Workspaces CRUD, manage members (except owners) |
| `member` | 1 | Read access, use chatbots |

---

## 5.2 Middleware Enforcement

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| RBAC-001 | Member accesses member-required endpoint | 200 OK | ✅ |
| RBAC-002 | Member accesses admin-required endpoint | 403 Forbidden | ✅ |
| RBAC-003 | Member accesses owner-required endpoint | 403 Forbidden | ✅ |
| RBAC-004 | Admin accesses member-required endpoint | 200 OK | ✅ |
| RBAC-005 | Admin accesses admin-required endpoint | 200 OK | ✅ |
| RBAC-006 | Admin accesses owner-required endpoint | 403 Forbidden | ✅ |
| RBAC-007 | Owner accesses all endpoints | 200 OK | ✅ |
| RBAC-008 | Non-member accesses any endpoint | 403 Forbidden | ✅ |
| RBAC-009 | Missing Authorization header | 401 Unauthorized | ✅ |
| RBAC-010 | Invalid organization ID in path | 400 Bad Request | ✅ |

### Technical Notes

```go
// pkg/middleware/organization.go:RequireOrganizationAccess
// Extracts orgID from path: /api/v1/organizations/:orgId/...
// Checks membership and minimum role
```

---

## 5.3 Route-Level RBAC Matrix

| Endpoint | Method | Required Role | Test ID |
|----------|--------|---------------|---------|
| `/api/v1/organizations/:id` | GET | member | RBAC-011 ✅ |
| `/api/v1/organizations/:id` | PATCH | owner | RBAC-012 ✅ |
| `/api/v1/organizations/:id` | DELETE | owner | RBAC-013 ✅ |
| `/api/v1/organizations/:id/workspaces` | GET | member | RBAC-014 ✅ |
| `/api/v1/organizations/:id/workspaces` | POST | admin | RBAC-015 ✅ |
| `/api/v1/organizations/:id/workspaces/:wid` | PATCH | admin | RBAC-016 ✅ |
| `/api/v1/organizations/:id/workspaces/:wid` | DELETE | admin | RBAC-017 ✅ |
| `/api/v1/organizations/:id/members` | GET | member | RBAC-018 ✅ |
| `/api/v1/organizations/:id/members` | POST | admin | RBAC-019 ✅ |
| `/api/v1/organizations/:id/members/:uid` | PATCH | admin | RBAC-020 ✅ |
| `/api/v1/organizations/:id/members/:uid` | DELETE | admin | RBAC-021 ✅ |

---

## 5.4 Service-Level Constraints

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| SVC-001 | Self-promotion (member→admin) | Error: "cannot promote yourself" | ✅ |
| SVC-002 | Self-promotion (admin→owner) | Error: "cannot promote yourself" | ✅ |
| SVC-003 | Admin assigns owner role | Error: "only owners can assign owner role" | ✅ |
| SVC-004 | Owner assigns owner role | Success | ✅ |
| SVC-005 | Demote last owner | Error: "cannot demote the last owner" | ✅ |
| SVC-006 | Remove last owner | Error: "cannot remove the last owner" | ✅ |
| SVC-007 | Delete last organization | Error: "cannot delete the last organization" | ✅ |
| SVC-008 | Delete last workspace | Error: "cannot delete the last workspace" | ✅ |
| SVC-009 | Two owners, demote one | Success | ✅ |
| SVC-010 | Invalid role value | Error: "invalid role" | ✅ |

### Technical Notes

```go
// internal/services/organization_service.go
// Lines 362-365: Prevent self-promotion
// Lines 378-381: Only owners can assign owner role
// Lines 367-376, 321-329: Last owner protection
```

---

## 5.5 Workspace Scoping

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| WSC-001 | Chatbot created with workspace_id | Scoped to workspace | ✅ |
| WSC-002 | Chatbot access from another workspace | 403 or 404 | ✅ |
| WSC-003 | `X-Workspace-ID` header extraction | Context populated | ✅ |
| WSC-004 | Sources scoped to workspace | Isolation verified | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/integration/organization_role_test.go` | Role enforcement |
| `internal/services/organization_service_test.go` | Service constraints |
