> Extracted from: organization-workspace.md — Section 1

# Organization & Workspace Management — Feature Testing

> **Document Version**: 1.0
> **Last Updated**: July 2026
> **Scope**: Organization & Workspace Management
> **Backend**: Go 1.25+ API
> **Test Frameworks**: Go `testing` (backend unit/integration), Playwright (frontend E2E)

---

# 1. Organization & Workspace Management

## 1.1 Feature Overview

Multi-tenancy is implemented through Organizations and Workspaces. An org has members with roles, and workspaces provide further isolation within an org. Each org has an owner who can manage members. A default org and workspace are created upon registration.

## 1.2 Detailed Sub-Features

### 1.2.1 List Organizations (`GET /api/v1/organizations`)

- [ ] Return all organizations the current user is a member of
- [ ] Return default organization created at registration
- [ ] Return org with `id`, `name`, `slug`, `ownerId`, `createdAt`, `memberCount`
- [ ] Return owned orgs and orgs where user is a member
- [ ] Return empty array when user belongs to no orgs (edge case)
- [ ] Return 401 when not authenticated
- [ ] Return paginated results when user is in many orgs

### 1.2.2 Create Organization (`POST /api/v1/organizations`)

- [ ] Create organization with valid name and slug
- [ ] Auto-generate slug from name when not provided
- [ ] Auto-generate unique slug when slug is taken (append suffix)
- [ ] Reject creation with duplicate slug
- [ ] Reject creation with empty name
- [ ] Reject creation with name exceeding max length
- [ ] Reject creation with invalid slug characters (alphanumeric + hyphens only)
- [ ] Verify creator becomes the org owner
- [ ] Verify default workspace is created inside the new org
- [ ] Verify org limit enforcement (max orgs per user from plan limits)
- [ ] Return 401 when not authenticated

### 1.2.3 Update Organization (`PATCH /api/v1/organizations/{id}`)

- [ ] Update org name
- [ ] Update org slug
- [ ] Reject update by non-owner member
- [ ] Reject update with empty name
- [ ] Reject update with duplicate slug (different org)
- [ ] Return 404 for non-existent org
- [ ] Return 401 when not authenticated

### 1.2.4 Delete Organization (`DELETE /api/v1/organizations/{id}`)

- [ ] Delete own organization (owner only)
- [ ] Verify all workspaces under org are deleted (cascade)
- [ ] Verify chatbots in workspaces are disassociated or moved
- [ ] Reject deletion by non-owner
- [ ] Return 404 for non-existent org
- [ ] Return 409 if org has active chatbots (prevent accidental deletion)
- [ ] Verify tenant data isolation — other orgs' data unchanged

### 1.2.5 List Members (`GET /api/v1/organizations/{id}/members`)

- [ ] Return all members with `id`, `email`, `fullName`, `role`, `joinedAt`
- [ ] Return owner as first member with role `owner`
- [ ] Return members sorted by `joinedAt`
- [ ] Return empty array if org has no members (shouldn't happen — owner always exists)
- [ ] Return 403 for non-member access
- [ ] Return 404 for non-existent org
- [ ] Return paginated results

### 1.2.6 Add Member (`POST /api/v1/organizations/{id}/members`)

- [ ] Add existing user by email
- [ ] Add multiple members in batch
- [ ] Reject adding non-existent email
- [ ] Reject adding user who is already a member
- [ ] Reject add by non-owner/non-admin member
- [ ] Reject add when org is at max member limit
- [ ] Return 404 for non-existent org
- [ ] Verify new member has role `member` by default
- [ ] Verify member limit enforcement from plan

### 1.2.7 Remove Member (`DELETE /api/v1/organizations/{id}/members/{userID}`)

- [ ] Remove member from org
- [ ] Reject removing the owner
- [ ] Reject remove by non-owner
- [ ] Return 404 for non-existent member
- [ ] Return 404 for non-existent org
- [ ] Verify member loses access to org resources
- [ ] Verify member still exists as a user (not deleted)

### 1.2.8 Update Member Role (`PATCH /api/v1/organizations/{id}/members/{userID}`)

- [ ] Promote member to `admin`
- [ ] Demote admin to `member`
- [ ] Reject changing owner's role
- [ ] Reject update by non-owner (only owner can change roles)
- [ ] Reject invalid role value
- [ ] Return 404 for non-existent member
- [ ] Return 404 for non-existent org

### 1.2.9 List Workspaces (`GET /api/v1/organizations/{id}/workspaces`)

- [ ] Return all workspaces in org
- [ ] Return default workspace created with org
- [ ] Return workspace with `id`, `orgId`, `name`, `slug`, `createdAt`
- [ ] Return 403 for non-member access
- [ ] Return 404 for non-existent org
- [ ] Return empty array if only default workspace exists

### 1.2.10 Create Workspace (`POST /api/v1/organizations/{id}/workspaces`)

- [ ] Create workspace with valid name
- [ ] Auto-generate slug from name
- [ ] Reject creation with duplicate name within same org
- [ ] Reject creation by non-member
- [ ] Reject creation when workspace limit is reached (from plan limits)
- [ ] Verify creator (or org owner) is workspace admin

### 1.2.11 Update Workspace (`PATCH /api/v1/organizations/{id}/workspaces/{wsID}`)

- [ ] Update workspace name
- [ ] Update workspace slug
- [ ] Reject update by non-member
- [ ] Return 404 for non-existent workspace
- [ ] Return 404 for non-existent org

### 1.2.12 Delete Workspace (`DELETE /api/v1/organizations/{id}/workspaces/{wsID}`)

- [ ] Delete workspace
- [ ] Reject deleting the last workspace in an org
- [ ] Reject delete by non-owner/non-admin
- [ ] Verify chatbots in workspace are disassociated or moved to default workspace
- [ ] Return 404 for non-existent workspace

## 1.3 API Endpoints Table

| Method | Path | Handler | Expected Status Codes |
|--------|------|---------|----------------------|
| `GET` | `/api/v1/organizations` | `oh.ListOrCreate` | 200, 401 |
| `POST` | `/api/v1/organizations` | `oh.ListOrCreate` | 201, 400, 401, 409 |
| `PATCH` | `/api/v1/organizations/{id}` | `oh.UpdateOrganization` | 200, 400, 401, 403, 404 |
| `DELETE` | `/api/v1/organizations/{id}` | `oh.DeleteOrganization` | 200, 401, 403, 404, 409 |
| `GET` | `/api/v1/organizations/{id}/members` | `oh.GetMembers` | 200, 401, 403, 404 |
| `POST` | `/api/v1/organizations/{id}/members` | `oh.AddMember` | 201, 400, 401, 403, 404, 409 |
| `DELETE` | `/api/v1/organizations/{id}/members/{userID}` | `oh.RemoveMember` | 200, 401, 403, 404 |
| `PATCH` | `/api/v1/organizations/{id}/members/{userID}` | `oh.UpdateMemberRole` | 200, 400, 401, 403, 404 |
| `GET` | `/api/v1/organizations/{id}/workspaces` | `wh.Workspaces` | 200, 401, 403, 404 |
| `POST` | `/api/v1/organizations/{id}/workspaces` | `wh.Workspaces` | 201, 400, 401, 403, 404, 409 |
| `PATCH` | `/api/v1/organizations/{id}/workspaces/{wsID}` | `wh.UpdateWorkspace` | 200, 400, 401, 403, 404 |
| `DELETE` | `/api/v1/organizations/{id}/workspaces/{wsID}` | `wh.DeleteWorkspace` | 200, 401, 403, 404, 409 |

## 1.4 Test Scenarios

### Scenario 1: Full Org Lifecycle

```
Step 1: Register → Default org created
  - [ ] Register new user
  - [ ] GET /api/v1/organizations → 1 org (default)
  - [ ] Assert: default org name matches "{User}'s Organization" pattern

Step 2: Create a second org
  - [ ] POST /api/v1/organizations with { "name": "Acme Corp" }
  - [ ] Assert: 201 Created
  - [ ] Assert: slug = "acme-corp"
  - [ ] Assert: ownerId = current user ID
  - [ ] Assert: Default workspace created inside org

Step 3: Create org with slug collision
  - [ ] POST /api/v1/organizations with { "name": "Acme Corp" }
  - [ ] Assert: 201 Created (different slug, e.g. "acme-corp-1")

Step 4: Update org
  - [ ] PATCH /api/v1/organizations/{acmeID} with { "name": "Acme Corporation" }
  - [ ] Assert: 200 OK
  - [ ] Assert: name updated

Step 5: Delete org
  - [ ] DELETE /api/v1/organizations/{acmeID}
  - [ ] Assert: 200 OK
  - [ ] GET /api/v1/organizations → default org still exists, Acme is gone
```

### Scenario 2: Member Management Workflow

```
Step 1: Setup
  - [ ] Register User A (org owner)
  - [ ] Register User B
  - [ ] Register User C

Step 2: Owner adds members
  - [ ] POST /api/v1/organizations/{orgID}/members with { "email": "user_b@example.com" }
  - [ ] Assert: 201 Created
  - [ ] POST /api/v1/organizations/{orgID}/members with { "email": "user_c@example.com" }
  - [ ] Assert: 201 Created

Step 3: Promote User B to admin
  - [ ] PATCH /api/v1/organizations/{orgID}/members/{userB_ID} with { "role": "admin" }
  - [ ] Assert: 200 OK

Step 4: Admin adds member
  - [ ] Login as User B
  - [ ] POST /api/v1/organizations/{orgID}/members with { "email": "non_existent@example.com" }
  - [ ] Assert: 400 Bad Request (user doesn't exist)

Step 5: Member cannot add members
  - [ ] Login as User C (role: member)
  - [ ] POST /api/v1/organizations/{orgID}/members with { "email": "someone@example.com" }
  - [ ] Assert: 403 Forbidden

Step 6: Owner removes User C
  - [ ] Login as User A
  - [ ] DELETE /api/v1/organizations/{orgID}/members/{userC_ID}
  - [ ] Assert: 200 OK
  - [ ] GET /api/v1/organizations/{orgID}/members → User C not in list

Step 7: Cannot remove owner
  - [ ] DELETE /api/v1/organizations/{orgID}/members/{userA_ID}
  - [ ] Assert: 400 or 403 (cannot remove owner)
```

### Scenario 3: Workspace Isolation

```
Step 1: Setup org with two workspaces
  - [ ] POST /api/v1/organizations/{orgID}/workspaces with { "name": "Development" }
  - [ ] Assert: 201 Created, slug = "development"
  - [ ] POST /api/v1/organizations/{orgID}/workspaces with { "name": "Production" }
  - [ ] Assert: 201 Created, slug = "production"

Step 2: Create chatbots in different workspaces
  - [ ] POST /api/v1/chatbots with { "name": "Dev Bot", "workspaceId": "{devWsID}" }
  - [ ] Assert: 201
  - [ ] POST /api/v1/chatbots with { "name": "Prod Bot", "workspaceId": "{prodWsID}" }
  - [ ] Assert: 201

Step 3: List chatbots per workspace
  - [ ] GET /api/v1/chatbots?workspaceId={devWsID} → only Dev Bot
  - [ ] GET /api/v1/chatbots?workspaceId={prodWsID} → only Prod Bot

Step 4: Delete workspace with chatbots
  - [ ] DELETE /api/v1/organizations/{orgID}/workspaces/{devWsID}
  - [ ] Assert: 200 OK or 409 (if prevention enabled)
  - [ ] If allowed: Verify Dev Bot is moved to default workspace

Step 5: Non-member cannot see workspaces
  - [ ] Login as User B (not a member of this org)
  - [ ] GET /api/v1/organizations/{orgID}/workspaces → 403
```

### Scenario 4: Tenant Data Isolation

```
Step 1: Setup two independent orgs
  - [ ] User A creates "Org A"
  - [ ] User B registers, gets "Org B" (default)

Step 2: Cross-org access attempts
  - [ ] User A tries GET /api/v1/organizations/{orgB_ID} → 403 or 404
  - [ ] User A tries GET /api/v1/organizations/{orgB_ID}/workspaces → 403 or 404
  - [ ] User A tries GET /api/v1/organizations/{orgB_ID}/members → 403 or 404
  - [ ] User A tries PATCH /api/v1/organizations/{orgB_ID} → 403 or 404
  - [ ] User A tries DELETE /api/v1/organizations/{orgB_ID} → 403 or 404
```

### Scenario 5: Org and Workspace Slug Auto-Generation

```
  - [ ] POST /api/v1/organizations with { "name": "My Organization" }
  - [ ] Assert: slug = "my-organization"
  - [ ] POST /api/v1/organizations with { "name": "My Organization" }
  - [ ] Assert: slug = "my-organization-1" (unique suffix)
  - [ ] POST /api/v1/organizations with { "name": "My___Org!!!", "slug": "" }
  - [ ] Assert: slug = "my-org" (sanitized)
  - [ ] POST /api/v1/organizations with { "name": "ABC", "slug": "custom-slug" }
  - [ ] Assert: slug = "custom-slug"
  - [ ] POST /api/v1/organizations with { "name": "ABC", "slug": "CUSTOM-SLUG" }
  - [ ] Assert: slug = "custom-slug" (lowercased)
```

## 1.5 Edge Cases

| Edge Case | Expected Behavior |
|-----------|------------------|
| Create org with duplicate slug (race condition) | Unique constraint in DB; second request gets 409 Conflict |
| Delete last workspace in an org | 409 Conflict — org must have at least one workspace |
| Add member who is already a member | 409 Conflict |
| Remove oneself from org | 409 Conflict — cannot remove yourself (must transfer ownership or delete org) |
| Org name with special characters | Allowed; slug is sanitized to alphanumeric + hyphens |
| Org name that is 100% special characters | Slug falls back to "organization" or shortened form |
| Very long org name (>255 chars) | 400 Bad Request |
| Member limit exceeded | 409 Conflict with message indicating plan limit |
| Add member by UUID (not email) | Endpoint expects email; UUID treated as email → 400 or 404 |
| Transfer org ownership | Not supported directly; requires manual DB operation |
| Concurrent member add for same email | Unique constraint prevents duplicates |
| Deleted org appears in member's list | No — cascade or soft-delete filters it out |

## 1.6 Expected Behavior

| Condition | HTTP Status | Response Body |
|-----------|-------------|---------------|
| List orgs (member) | 200 | `{ "organizations": [{ "id", "name", "slug", "ownerId", "createdAt", "memberCount" }] }` |
| Create org success | 201 | `{ "id", "name", "slug", "ownerId", "createdAt" }` |
| Create org duplicate slug | 409 | `{ "error": { "code": "SLUG_TAKEN", "message": "..." } }` |
| Update org success | 200 | `{ "id", "name", "slug", "ownerId", "createdAt" }` |
| Update org forbidden (non-owner) | 403 | `{ "error": { "code": "FORBIDDEN" } }` |
| Delete org success | 200 | `{ "message": "Organization deleted" }` |
| Delete org with active chatbots | 409 | `{ "error": { "code": "ORG_HAS_CHATBOTS" } }` |
| List members success | 200 | `{ "members": [{ "id", "email", "fullName", "role", "joinedAt" }] }` |
| Add member success | 201 | `{ "id", "email", "fullName", "role": "member", "joinedAt" }` |
| Add member duplicate | 409 | `{ "error": { "code": "ALREADY_MEMBER" } }` |
| Remove member success | 200 | `{ "message": "Member removed" }` |
| Remove owner | 400/403 | `{ "error": { "code": "CANNOT_REMOVE_OWNER" } }` |
| Update role success | 200 | `{ "id", "email", "fullName", "role", "joinedAt" }` |
| List workspaces success | 200 | `{ "workspaces": [{ "id", "orgId", "name", "slug", "createdAt" }] }` |
| Create workspace success | 201 | `{ "id", "orgId", "name", "slug", "createdAt" }` |
| Delete last workspace | 409 | `{ "error": { "code": "LAST_WORKSPACE" } }` |
| Non-member access | 403 | `{ "error": { "code": "FORBIDDEN" } }` |
| Non-existent resource | 404 | `{ "error": { "code": "NOT_FOUND" } }` |
