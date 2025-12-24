package services

import "errors"

// MI-004: Sentinel errors for organization/workspace operations
// These replace string-based error checking for more robust error handling

// Organization errors
var (
	ErrOrgSlugExists      = errors.New("organization slug already exists")
	ErrLastOrganization   = errors.New("cannot delete the last organization")
	ErrNotMember          = errors.New("not a member of this organization")
	ErrLastOwner          = errors.New("cannot remove the last owner")
	ErrInvalidRole        = errors.New("invalid role")
	ErrCannotPromoteSelf  = errors.New("cannot promote yourself")
	ErrCannotDemoteOwner  = errors.New("cannot demote the last owner")
	ErrOnlyOwnersCanGrant = errors.New("only owners can assign owner role")
)

// Workspace errors
var (
	ErrWorkspaceSlugExists = errors.New("workspace slug already exists in this organization")
	ErrLastWorkspace       = errors.New("cannot delete the last workspace in the organization")
)
