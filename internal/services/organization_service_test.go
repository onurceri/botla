package services

// NOTE: Organization service tests require database infrastructure.
// The tests are designed to run against the integration test DB.
// Run with: go test ./internal/integration/... -run Organization
//
// Test coverage includes:
// - CreateOrganization (success, duplicate slug)
// - CreateWorkspace (success, duplicate slug in org, same slug different org)
// - GetUserOrganizations (empty, with orgs)
// - CheckMembership (owner, non-member)
// - AddMember (new, upsert)
// - GetWorkspaces (empty, with workspaces)
