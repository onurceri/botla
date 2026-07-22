> Extracted from: security-middleware.md — Section: 2.6 Organization Middleware (`organization.go`)

## 2.6 Organization Middleware (`organization.go`)

- [ ] `RequireOrganizationAccess` checks user's org membership
- [ ] `ExtractTenantContext` injects tenant (workspace) scope into request context
- [ ] Returns 403 if user does not belong to the org
- [ ] Returns 403 if user does not have access to the workspace
- [ ] Org/workspace ID is extracted from the URL path
- [ ] Context is available downstream for data scoping
