# 8.1 Organization Management Test Plan

## Overview
This test plan covers organization CRUD operations and member management.

---

## Test Cases

### 8.1.1 Create Organization
**Priority:** Critical  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST `/api/v1/organizations` | 201 Created |
| 2 | Creator assigned owner role | Membership created |
| 3 | Slug is unique | Stored |

---

### 8.1.2 Duplicate Slug Prevention
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create org with slug "my-org" | Success |
| 2 | Create another with same slug | 409 Conflict |

---

### 8.1.3 List User Organizations
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/organizations` | 200 OK |
| 2 | Only orgs user belongs to | Filtered list |
| 3 | Includes user's role | owner/admin/member |

---

### 8.1.4 Update Organization
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Owner updates name | 200 OK |
| 2 | Admin updates name | 200 OK |
| 3 | Member updates name | 403 Forbidden |

---

### 8.1.5 Delete Organization
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Owner deletes | 200 OK |
| 2 | Admin deletes | 403 Forbidden |
| 3 | Cascade to workspaces | All deleted |

---

### 8.1.6 Add Member
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST add member by email | 200 OK |
| 2 | Membership created | With role |
| 3 | Non-existent email | 404 Not Found |

---

### 8.1.7 Remove Member
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Remove member | 200 OK |
| 2 | Cannot remove last owner | 400 Bad Request |

---

### 8.1.8 Update Member Role
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Owner promotes to admin | 200 OK |
| 2 | Admin cannot promote to owner | 403 Forbidden |
| 3 | Member cannot change roles | 403 Forbidden |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Organization"
```
