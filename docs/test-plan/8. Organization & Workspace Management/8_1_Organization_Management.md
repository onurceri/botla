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

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Create user.
- **Steps:**
  1. POST `/api/v1/organizations` with `{"name": "New Org", "slug": "new-org"}`.
  2. Verify 201.
  3. Query `organization_members` for `user_id` and `organization_id`. Verify `role` is "owner".

---

### 8.1.2 Duplicate Slug Prevention
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create org with slug "my-org" | Success |
| 2 | Create another with same slug | 409 Conflict |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Steps:**
  1. POST org `slug="dup-test"`. Expect 201.
  2. POST org `slug="dup-test"`. Expect 409.

---

### 8.1.3 List User Organizations
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | GET `/api/v1/organizations` | 200 OK |
| 2 | Only orgs user belongs to | Filtered list |
| 3 | Includes user's role | owner/admin/member |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Create User A, User B.
  - Create Org 1 (A owner). Add B as member.
  - Create Org 2 (B owner).
- **Steps:**
  1. Login as A. GET `/organizations`. Verify list contains Org 1. Verify `role` is "owner".
  2. Login as B. GET `/organizations`. Verify list contains Org 1 ("member") and Org 2 ("owner").

---

### 8.1.4 Update Organization
**Priority:** Medium  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Owner updates name | 200 OK |
| 2 | Admin updates name | 200 OK |
| 3 | Member updates name | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Org with Owner, Admin, Member.
- **Steps:**
  1. Owner updates name. Expect 200.
  2. Admin updates name. Expect 200.
  3. Member updates name. Expect 403.

---

### 8.1.5 Delete Organization
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Owner deletes | 200 OK |
| 2 | Admin deletes | 403 Forbidden |
| 3 | Cascade to workspaces | All deleted |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Org with Owner, Admin. Org has a Workspace.
- **Steps:**
  1. Admin attempts delete. Expect 403.
  2. Owner attempts delete. Expect 200.
  3. Verify Workspace is deleted (cascade check).

---

### 8.1.6 Add Member
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST add member by email | 200 OK |
| 2 | Membership created | With role |
| 3 | Non-existent email | 404 Not Found |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Org Owner. User B (not member).
- **Steps:**
  1. POST `/organizations/{id}/members` with `email="userb@example.com"`.
  2. Verify 200.
  3. Verify B is now a member.
  4. POST with `email="nobody@example.com"`. Verify 404.

---

### 8.1.7 Remove Member
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Remove member | 200 OK |
| 2 | Cannot remove last owner | 400 Bad Request |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Org with Owner and Member.
- **Steps:**
  1. Owner removes Member. Expect 200.
  2. Owner removes Self (assuming only 1 owner). Expect 400.

---

### 8.1.8 Update Member Role
**Priority:** High  
**Type:** Integration Test

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Owner promotes to admin | 200 OK |
| 2 | Admin cannot promote to owner | 403 Forbidden |
| 3 | Member cannot change roles | 403 Forbidden |

**Implementation Plan:**
- **Test File:** `internal/integration/org_workspace_mgmt_test.go`
- **Setup:**
  - Org with Owner, Admin, Member.
- **Steps:**
  1. Owner sets Member role to "admin". Expect 200.
  2. Admin sets Member role to "owner". Expect 403.
  3. Member sets Self role to "admin". Expect 403.

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Organization"
```
