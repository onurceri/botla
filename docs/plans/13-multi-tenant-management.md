# Plan 3.2: Multi-Tenant Management Features

## Overview
While the core multi-tenant architecture (database, basic list/create APIs) is in place, the management capabilities are missing. Users cannot update names, delete entities, or manage memberships (invite/remove users). This plan addresses these gaps.

## 1. Backend Implementation

### 1.1 Organization Handlers
**File:** `internal/api/handlers/organization.go`

Add the following methods to `OrganizationHandlers`:
- `UpdateOrganization(w, r)`: PATCH `/api/v1/organizations/{id}`
  - Updates name and slug.
  - Requires `owner` or `admin` role.
- `DeleteOrganization(w, r)`: DELETE `/api/v1/organizations/{id}`
  - Soft or hard delete.
  - Requires `owner` role.

### 1.2 Workspace Handlers
**File:** `internal/api/handlers/organization.go`

Add the following methods to `OrganizationHandlers`:
- `UpdateWorkspace(w, r)`: PATCH `/api/v1/organizations/{orgID}/workspaces/{wsID}`
  - Updates name and slug.
  - Requires `owner` or `admin` role.
- `DeleteWorkspace(w, r)`: DELETE `/api/v1/organizations/{orgID}/workspaces/{wsID}`
  - Requires `owner` or `admin` role.

### 1.3 Membership Handlers
**File:** `internal/api/handlers/organization.go` (or new `membership.go`)

Add methods for member management:
- `GetMembers(w, r)`: GET `/api/v1/organizations/{id}/members`
  - Lists all members with their roles.
- `AddMember(w, r)`: POST `/api/v1/organizations/{id}/members`
  - Adds a user by email (requires looking up user by email first).
  - Body: `{ "email": "user@example.com", "role": "member" }`
- `RemoveMember(w, r)`: DELETE `/api/v1/organizations/{id}/members/{userID}`
  - Removes a user from the organization.
- `UpdateMemberRole(w, r)`: PATCH `/api/v1/organizations/{id}/members/{userID}`
  - Updates role (e.g., member -> admin).

### 1.4 Service Layer Updates
**File:** `internal/services/organization_service.go`

Ensure service methods exist for:
- `UpdateOrganization`
- `DeleteOrganization`
- `UpdateWorkspace`
- `DeleteWorkspace`
- `RemoveMember`
- `GetUserByEmail` (might belong in UserService, but needed here).

### 1.5 Router Configuration
**File:** `cmd/server/main.go`

Register the new endpoints:
- `PATCH /api/v1/organizations/{id}`
- `DELETE /api/v1/organizations/{id}`
- `PATCH /api/v1/organizations/{id}/workspaces/{wsID}`
- `DELETE /api/v1/organizations/{id}/workspaces/{wsID}`
- `GET /api/v1/organizations/{id}/members`
- `POST /api/v1/organizations/{id}/members`
- `DELETE /api/v1/organizations/{id}/members/{userID}`
- `PATCH /api/v1/organizations/{id}/members/{userID}`

## 2. Frontend Implementation

### 2.1 Organization Settings UI
- Create `src/features/organization/pages/OrganizationSettingsPage.tsx`.
- Tabs: "General" (Name/Slug), "Members" (List/Add/Remove), "Billing" (Plan).
- "General": Form to rename or delete the organization.
- "Members": Data table of members with "Add Member" button and "Remove" actions.

### 2.2 Workspace Settings UI
- Create `src/features/organization/pages/WorkspaceSettingsPage.tsx`.
- Form to rename or delete the workspace.

### 2.3 Navigation Integration
- Add "Settings" (Ayarlar) link in the `OrganizationSwitcher` dropdown or sidebar.
- Ensure proper routing in `App.tsx` or `DashboardLayout`.

## 3. Verification
- Test creating an organization -> Updating it -> Deleting it.
- Test adding a member -> verifying they can access -> removing them -> verifying access lost.
